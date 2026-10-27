# Replay

Replay allows us to force the execution of a program to follow a given trace.

Being able to replay has a great benefit for analyzing or bugging programs with
multiple possible applications.

Many analysis tools may produce false positives when trying to detect bugs.
With recording, it is possible to [rewrite](./rewrite.md) a program trace,
in which no actual bug has been executed but the analysis has indicated
a potential bug, in such a way, that the bug should occur. By replaying
this new trace, we may be able to confirm that such a bug can actually occur.

Additionally, the replay mechanism can be used for implementing a guided
fuzzing approach. By recording, mutating and then executing scheduling orders,
we may able to directly detect bugs or to execute new code, which was not
run in previous runs. Using such a form of guided fuzzing has way better
potential, especially in concurrent programs, to execute unlikely, but still
possible paths or execution schedules in the program than
approaches, where only the input of a program is mutated. This allows for a much
more guided approach, reducing the number of required runs and therefore
the expected runtime to detect possible bugs.

Here we explain the complete and partial replay used in the replay of
[rewritten traces](./rewrite.md) and the [GoPie](./fuzzing/GoPie.md) fuzzing.

For the [complete replay](#complete-replay), we force the order of the execution of all
concurrency operations. For the [partial replay](#partial-replay), we provide a list of
active operations. For those operations we force the order of the executions,
while all other operations can run freely.

In most cases, especially for the replay of rewritten traces, we want the
replay to execute the given trace as closely as possible. For this the
complete replay is used. But in certain situations this is to strong of an
assumption. E.g. in fuzzing, a mutation of the order of some element in
a trace may be difficult to merge back into the trace in such a way, that a
valid trace is created, even if the order of the mutated elements is valid.
In this case it may be useful to switch to the partial replay, where only
the order of the mutated elements is enforced, preventing the program from
getting stuck due to an impossible trace.

For the explanation of the simplified replay used for
flow fuzzing, see [here](./fuzzing/Flow.md#implementations).

## Toolchain

To replay a trace, import 

```
import _ "advocatego"
``` 

and

build the program with 

```
./go build -gcflags=all="-advocatereplay -advocatepath=[path]"
```

where [path] is the relative path to the trace.

When using the toolchain to run replays, this is automatically added.

## Complete replay

The following is a description of the implementation of the trace replay.

First we will give an [overview](#replay-mechanism) over the replay mechanism. Then
we will give an [detailed explanation](#detail) of the implementation.

### Replay Mechanism

The replay is implemented as an wait and release mechanism.

When an operation wants to execute, it checks if it is the next operation
in the trace. If it is, it will execute. If not, it will wait until it is the
operations turn.

There are two main parts in this mechanism, the operation and the replay
manager.\
The operations wants to execute, and if it is not the next element
to be executed, it will wait.\
The replay manager is a routine running in the background to
release waiting operations.

The replay in the operation works as follows:

```
func operation(op) {
	if ignored(op) {  // (E1)
		execute(op)
		return
	}

	evt = headEvt()  // (E2)
	if match(op, evt) {  // (E3)
		execute(op)
		ackDir()  // (E6)
	} else {
		sig = suspend(op)  // (E4)
	    waitMap[op] = sig
		if sig.ok {  // (E5)
			execute(op)
			ack()  // (E6)
		}
	}
}

E1: If the operations is an ignored (internal) operation, execute directly.
E2: Get the next element in the trace.
E3: If current operations matches the next event, then executed and send a direct acknowledgement to the manager.
E4: Otherwise, suspend operation and store a signal to wake-up the operation in map.
E5: When the wake-up signal is triggered, execute the operation.
E6: After execution, send an acknowledgement to the manager. For `Once.Do`, the
    acknowledgement is send after the `Do` has decided whether it will execute
    the given function (whether it is the first call to this once) or not, but
    before the argument function is run.
```

The replay manager runs in a background routine and is implemented as followed
(here we ignore the timeout mechanism. For details on this, see the [detailed](#detail) explanation):

```
func replayManager() {
	while(replayInProgress()) {
		evt = headEvt()  // (M1)

		if hasRelDirectly() {  // (M2)
			if !isChannel(evt) {
				waitAck()
			}
			nextEvt()
		}

		if sig, ok := waitMap[evt]; ok {  // (M3)
			release(sig)
			delete waitMap[evt]

			if !isChannelCom {  // (M4)
				waitAck()
			}

			nextEvt()  // (M5)
		}
	}
}

M1: Get the next element in the trace.
M2: Check if the next element in the trace has been directly
    executed (ackDir). If this is the case, and it is not a channel, wait for
    the operation to fully execute (wait for ack). Then advance the trace to
    the next element.
M3: If the next operation corresponds to an already waiting operation,
    send the signal to release it and remove the operation from the waitMap.
M4: If it is not a channel communication or select, wait for the operation
    to fully execute (wait for ack).
M5: Advance to the next operation in the trace.
```

#### Acknowledgement

The implemented acknowledgements are necessary to prevent situations like the following:

Assume we have the following program code

```go
var m sync.Mutex

go func() {     // R1
	o.Do(f1())  // Do 1
}

go func() {     // R2
	o.Do(f2())  // Do 2
}
```

Only one of the `do` operations will execute its argument function. Assume in the replay, we have first executed both `go` statements and now want to execute first the `Do 1` and then the `Do 2`, therefore executing `f1` but not `f2`. If we release `Do 1` and then directly release `Do 2`, we can get the following situations. Since go routines do not directly correspond to hardware threads, it is possible that both routines are mapped to the same thread. Lets assume, the underlying thread first executes the replay release on `Do 1`, but before the `Do 1` can be executed, the scheduler switches to routine 2. Here the `Do 2` is released and then executed, therefore running `f2`. Then the scheduler switches back to the first routines, running `Do 1`. Since the once `o` has already executed a function it will not execute `f1`. Additionally, if `f2` contains operations which are effected by replay, the routine will get stuck, since the replay mechanism will never release them.

Similar situations can be constructed for situations, where operations with different execution times are executed directly next to each other. It the longer operation is executed first, it could happen that the next element, which executes much faster effectively executes first, even though it should have executed second.

To prevent this, we use an acknowledgement. When an operation is released, the replay manager will pause. When the operations is completed, it will send an acknowledgement to the replay manager. Only when the acknowledgement is received, the current element will be advanced to the next element in the trace, meaning the next element can only be released when the previous operations has fully executed.

Additionally, we only directly release elements (E3) if the replay manager is currently not waiting for an acknowledgement.

For most operation this works, since they can be executed consecutively. The only operation where this does not work are the channels. Assume we have the following code:

```go
c := make(chan int, 0)

go func() {
	c <- 1
}

<- c
```

If we would wait for the send to fully execute and send an acknowledgement before we release the receive, the program would get stuck, because the send and the receive need to execute at the same time. We therefore do not wait for acknowledgements on channel communication operations and selects (M4).

### Detail

The code for the replay is mainly in [advocate/advocate_replay.go](../goPatch/src/advocate/advocate_replay.go) and [runtime/advocate_replay.go](../goPatch/src/runtime/advocate_replay.go) as well as in the code implementation of all recorded operations.

[advocate/advocate_replay.go](../goPatch/src/advocate/advocate_replay.go) mainly contains the code to read in the trace and initialize the replay. When reading in the trace, all trace files are read.

The internal representation of the replay consists of a slice of `ReplayElements` called `replayData`.
Each `ReplayElement` represents one operation that is to be executed. It contains data about
the type of operation, the timestamp, the routine ID and the position of the operation
in the file. This position is used to connect the `ReplayElement` to the actual operation during the execution. Additionally it may contain
information about whether the operation is blocked, meaning it should start but never finish or whether it should execute successfully,
e.g. for `once.Do` and `TryLock`. For selects, it also contains the internal
index of the case that should be executed.\
After the trace local trace files have been read, the elements in `replayData`, representing
a global trace, are sorted by there timestamp. This allows us to always get the
next operation that should be executed.

To connect an operation from the trace to an operation in an executing program,
we create a key for each operation. This key consists of the file and line
in the program code, where the operation is located, and the ID of the routine
in the trace file, where the operation should be executed.
Using the routine is important because of the following situations:

```go
c := make(chan int, 1)

a := func() {
	c <- 1
}

go a()
go a()
```

If we assume, that both routines running the function `a` are already running
when the first send on `c` should be executed, we would not be able to
uniquely identify by the code position, which send of the sends should be executed first.

Each spawn operation in the trace
contains the ID of the new routine, it creates. When a spawn is executed,
it will store this `replayID` in the new routine (in the `*g` object, where we
also record the traces, [implemented here](../goPatch/src/runtime/proc.go#L5083)). When a operation wants to execute, it will get this
ID from the routine and its code position from `runtime.Caller`. With those
information, an operation can be (mostly) uniquely connected to an operation in
the trace. The only case, where operation would have the same key, is for operations
at the same code position and in the same routine. Since operations
in the same routine are executed consecutively, the order of the executions
of those operations must be the same as the order of those operations in the trace,
which allows us to uniquely connect such operations without any additional information.

To guarantee, that the recorded code positions are always the positions
where the operation actually occurs, optimization and
inlining is disables for running a replay.\
This is done by setting the `-gcflags="all=-N -l"` when building a program or running a test.

#### Replay in operations

In each of the implementations of the operations that are considered in the
replay, the following (or similar) code snipped has been added:

```go
chWait, chAck := runtime.WaitForReplay(runtime.OperationMutexTryLock, 2, true)
if chWait != nil {
	replayElem := <-chWait
}

...

if chWait != nil {
	close(chAck)
}

```

When a operation wants to execute
it will call the [WaitForReplay](../goPatch/src/runtime/advocate_replay.go#L587) function.
The arguments of the function
contain information about the operation (type of operation and
skip value for `runtime.Caller`) as well as information about whether the
operation will send an acknowledgement (`wAck`) or not.

This function will first check, if the operations is part of the replay.
It is not part of the replay, if replay is disabled, or if the operation is
an internal operation.\
We have decided not to replay (or record) internal operations.
The reason for this is that for most uses (e.g. bug analysis), they are not relevant
and unnecessarily increase the trace file size and the replay and recording time.
Additionally, they may be part of unpredictable operations like e.g. the garbage
collector, which would make the replay much more complicated to implements.\
If the operation is not part of the replay, `WaitForReplay` will return `wait = false`
and the function can directly be executed.

If the operations is part of the replay, `WaitForReplay` creates a wait channel
`chWait` and and acknowledgement channel `cka`, each with
buffer size 1.

It will then check if the replay manager is currently not waiting for an acknowledgement,
and if so if the operation is the next operation in the trace.
If this is the case, it will directly send the
`replayElement` from the trace over the `chWait` channel, to
directly release the operation. If the operation should send an acknowledgement,
`WaitForReplay` will inform the replay manger, by setting a global variable, that it should wait for an acknowledgement.
If not, it will inform it to advance to the next element in the trace.

If now direct release is possible, it will store those channels in a map `waitingOps`, where the key
is the operation key explained above.

When `WaitForReplay` has finished, if well return the channels to the operation.

The operation will then start to read on the the `chWait` channel.
When the operation is cleared to run, a message containing the trace element will
be send over this channel. This replay element contains the
information about whether the operation blocked and whether
it was successful. If it should block (tPost = 0), it will block the operation
forever. If the operation was not successful (only possible for once.Do or Try(R)Lock),
it will force the execution of the operation to follow this behavior.
When the operation is not a channel or select, it will send an empty message over the `ackCh` as soon as the operation is finished
(normally implemented by `defer`). This allows the manager to release the
next operation.

<center><img src="img/replayInOp.png" alt="Replay in Operations" width="400px" height=auto></center>

<center><img src="img/waitForReplay.png" alt="WaitForReplay" width="800px" height=auto></center>

#### Replay Manager

The replay manager releases the operations in the correct order.

To release the operations, a separate routine [ReplayManager](../goPatch/src/runtime/advocate_replay.go#L353) is run in the
background.

This routing loops as long as the replay is active.

The main loop of this manager is as follows:\
First the manager checks, if an operation has executed directly
without it being added to the `waitOps` map first. If this is the case,
it will, if required, check if the operation has already send its acknowledgement.
If not, it will wait for it. If the acknowledgement has arrived, it will
advance to the next element in the trace.

Then, the manager gets the next element that should be executed (see [here](#getnextreplayelement)).
If no element is left, the replay is disabled.
The same is true if the next element in the trace is the `replayEnd` element.
If the bug is confirmed by reaching a certain point in the code (e.g leak or
resource deadlock), this will confirm the replay.

If the element is an
operation element, the manager will check if the element should be ignored
for the replay. In this case, there is no waiting operation and the
manager will simply assume, that the operation was executed and will continue
the loop from the beginning.

If this is not the case, the it is checked, if the
operation is already waiting, meaning the corresponding key is in `waitingOps`.
If this is the case, it is [released](#release).

After that, the manager will continue the loop to get the next trace element.
If the operation is not yet waiting, the manager will directly restart the
loop without advancing to the next replay element, until the operation is
waiting.

<center><img src="img/replayManager.png" alt="Replay Manager" width="1200px" height=auto></center>

##### Timeout

It is possible, that in the replay something went wrong (see [here](#things-that-can-go-wrong))
causing the replay to get stuck.

THis may be cause by a single unexpected or invalid operation in the trace or
execution.To still get a chance that is may resolve itself, the manager
is able skip operations, if it senses that the
replay is stuck. This will be done, if the next replayElement is the same for
a too long time. In this case, the replay will trigger the timeout mechanism.

In this, either the oldest element in `waitOps` is released
or the next element in the trace is skipped. We choose one of them at random
(if no operation is waiting, we always skip the trace element).
We hope that this will clear the blockage, so that the replay can continue to be executed.
If an element is released as an oldest waiting, we add this element to a map
of operations. If the next element in the trace is skipped, we increase the
counter pointing to the next element in the trace
(compare [getNextReplayElement](#getnextreplayelement) and [release](#release)).

Sometimes this clears the problem, and the program can continue as normal,
if no element has been cleared regularly
for a certain time, the replay will assume that it it is completely stuck and cannot be
brought back by those irregular releases or skips.
It will therefore
disable the replay completely, meaning the program will continue without
any guidance (all waiting operations are released). If the replay is already
far enough, this may still result in interesting behavior, e.g. if the replay
is used for confirming a bug, the bug may still be triggered, if the program has already been
pushed far enough in the correct direction.

##### getNextReplayElement

The function to get the next element to be replayed is implemented [here](../goPatch/src/runtime/advocate_replay.go#L789).

The trace is stored in a sorted slice. The replay is therefore done in the
order, in which the operations occur in this trace. To keep track of this,
we have a counter, always pointing to the next operation to be executed.

When the replay manager or an operation requests the next replay element,
we return the element at the position, where this counter points to.

Before we return the element, we check if it is in the list of elements,
that where released by the timeout mechanism. If it is, we advance the counter
by one, and recursively call `getNextReplayElement` again.

If the counter is greater than the number of elements in the trace, we return and
empty element, signaling to the manager that all elements in the trace have been
replayed.

##### release

To [release](../goPatch/src/runtime/advocate_replay.go#L693) a waiting element,
we send the element info over the corresponding
channel, on which the operation is waiting.

If an acknowledgement is expected,
the release function will then read on the acknowledgement channel until
the acknowledgement is received.

To prevent a failed acknowledgement from
getting the replay stuck, we have a timeout. If no acknowledgement has been received
after a certain time, we continue the replay anyways. To prevent the case,
where an operation ties to send an acknowledgement after the timeout has been triggered
and therefore cannot send, we set the buffer size of all acknowledgement
channels to 1. They can therefore also send if no one is receiving it any more.

The release function than delete the operation from `waitingOps` and
increase the trace counter by 1.

<center><img src="img/release.png" alt="Release" width="1000px" height=auto></center>

#### Select

While most operations are fully determined by the order in which they are executed,
this may not be the case for selects. Here we need to determine which case
the replay should execute. For this an alternative implementation of the
[select structure with a preferred case](../goPatch/src/runtime/select.go#L151) has been implemented.

In the [original implementation](../goPatch/src/runtime/select.go#L629), the select has three phases. In the
first phase the code iterates over all cases and checks, if one of then can
immediately be executed.\
If this is the case, it is executed and the select returns. If this is not
the case, it is checked if there is a default case. If there is, it is executed.
If not, the select start its second phase. Here all channels in all the cases
enqueue the corresponding operations. The routine then parks, until
it is woken up by some routine that wants to communicate with one of the channels.
The communication is then executed. In the third phase, all communications
that have not been executed are dequeued again.

For the select with preferred case, the following changes have been made.
In the first pass we only check the case, where the internal index `casi`
is equal to the preferred index. If the default case is the preferred index,
we don't do anything here.\
If the preferred case is the default case, we directly execute the default case.\
In the second pass, we only enqueue the preferred case.

Similar to release of the oldest waiting element, we also want to release the
wait on the select if the runtime senses, that the replay may be stuck. For
this reason, we implement a `gopark` function with a timeout, that
automatically wakes it up after a certain time. The park with timeout is implemented
in [prog.go](../goPatch/src/runtime/proc.go#L439) as follows:

```go
func goparkWithTimeout(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceReason traceBlockReason, traceskip int, timeout int64) {
	mp := acquirem()
	gp := mp.curg

	// Setup timer if timeout is non-zero
	if timeout > 0 {
		go func() {
			sleep(float64(timeout))
			if readgstatus(gp) == _Gwaiting {
				goready(gp, traceskip)
			}
		}()
	}

	// original gopark implementation
  ...
}
```

The additionally routine will wake the sleeping routine if the timer
runs out before it is awoken by a communication partner.

When the routine is woken up by a communication partner, we do not need to
pass the third phase, since there are no more waiting channel in the select.
In this case the select returns. If it was awoken by the timeout, we dequeue
the waiting channel operation and return. In this case, the select is run again,
now with the original implementation, meaning without any guidance.

Selects with only one case and a default are implemented separately in go ([here](../goPatch/src/runtime/chan.go#L990) and [here](../goPatch/src/runtime/chan.go#L1057)).
This implementation tries to directly [send](../goPatch/src/runtime/chan.go#L201) or
[receive](../goPatch/src/runtime/chan.go#L645) on the channel in the non default case.
If this is not possible because no communication partner is available, it directly
calls the default.\
For the replay this is changed as follows: If the default case is the preferred case,
we directly run the default case without trying to send/recv. If the non-default
case is the preferred one, we busy-wait with a for loop, each time trying to send
or receive, until the communication succeeds or a timer runs out. If the timer
runs out first, we run the default case regardless.

#### Atomics

While for most operations, we could add the code for the [replay](#replay-in-operations)
directly into the implementation of those operations, this was not
possible for the atomic operations, since they are partially implemented in
assembly. We therefore needed to intersect an additional function call.
For more information about this, see the [atomic recording documentation](./trace/atomics.md#implementation).

#### FinishReplay

Go stops the execution of all routines in a program, as soon as all operations
in the main routine have been executed. In the replay, we want to avoid that the
program terminates, before all operations in the trace have been executed.
To do this, we implement a function [FinishReplay](../goPatch/src/advocate/advocate_replay.go#L447).
This function should be executed at the end of the main functions.
It will run, until all operations in the trace have been executed, stopping the
main routine from terminating to early.

## Partial replay

To implement some of the [Order based fuzzing](./fuzzing/GoPie.md), we use a
simplified partial replay. Here, we do not restrict the order of all
element in the trace and program. Instead, we have a list of active
elements from the trace.

When reading in a trace for a partial replay, we also read in
a file containing the information about this partiality. This file
contains the active elements and an information when the replay should
switch from the strict to the partial replay (this information can also
indicate to directly start with the partial replay). When reading in the
trace, the mechanism will ignore all elements that are after the specified
starting point and not in the set of active operations.
For operations that are in the list of active operations, we also store
how often the same operation (same file and line) appears before this
specific operation execution in the trace.

When the partial replay is active, it is executed as follows:

Elements that want to execute, but are not in the list of active elements
are directly executed without there order being influenced by the
replay mechanism.

For elements that want to execute, and are in the list of active
operations, we track how often the same operation has been executed before.
We then compare this number with the number we have stored. If it is not equal,
the execution represents the correct operation (file and line), but not the
correct execution of this operation. In this case the operation is directly
executed. If it is the same, the execution represents an active execution,
and is handled by the replay mechanism in the same way as in the total
replay.

## Things that can go wrong

It is possible, that either an element in the trace never tries to execute
of that an operation tries to execute that is not in the trace. This could
happen e.g. if the program uses randomness, if its behavior depends
on an outside communication (e.g. API call) or if the
control-flow changes due to non-atomic memory operations.

An example would be the following

```go

m := sync.Mutex{}
c := make(chan int, 1)

if rand.Float64() < 0.5 {
	c <- 1
}

m.Lock()

```

Assume that during the recording, the random number was less then 0.5,
meaning the channel send is part of the trace. If we now try to
replay this trace, it could happen that the value is now greater than 0.5.
When we now arrive at the lock operation, the mutex wants to execute,
but the replay manager still assumes, that the channel send should be the
next operations. This causes the replay mechanism to get stuck.

Similar situations could happen when using non-atomic operations depending on shared memory.
Let's assume we have the following program.

```go
a := 0

m := sync.Mutex{}

go func () {
	a = 1          // write
}()

go func () {
	if a == 0 {   // read
		m.Lock()
		...
		m.Unlock()
	}
}()
```

Our mechanism cannot influence the order of the write and read on `a`, since
they are not atomic operations. Whether the lock and unlock on `m` can or must
be executed, therefore depends on an order we cannot control, which may
lead to the program getting stuck.

To not get completely stuck if such operations occur, the replay mechanism
is able to release waiting elements without them being the next trace element
or to completely disable the replay, if it senses, that it is stuck (as
described in the [details](#timeout) section).
