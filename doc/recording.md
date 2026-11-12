# Execution, Recording and Trace

Gosc is able to record a program or test run, and to store the executed
interleaving in a trace. Those traces can be used to deterministically
[replay](./replay.md) a program run or to perform different analysis on the recorded execution.

To record the relevant operation into traces, we have added the required
functionality directly into the implementations of the operations. For this
we provide a modified [modified runtime](../goPatch/src/runtime).

Running a test or program with the correct preamble (see below) using this runtime,
will result in a trace being recorded.

For each of the operations we want to record, additional function calls
have been added to the operation implementations.
We record the following operations:

- [Channel](trace/channel.md): Send, Receive, Close
- [Select](trace/select.md)
- [Mutex](trace/mutex.md): Lock, RLock, TryLock, TryRLock, Unlock, RUnlock
- [WaitGroup](trace/waitGroup.md): Add, Done
- [Once](trace/once.md): Do
- [Conditional Variable](trace/conditionalVariables.md): Wait, Signal, Broadcast
- [Atomics](trace/atomics.md): Load, Store, Add Swap, CompareAndSwap
- [Alloc](trace/alloc.md)
- [Fork/Spawn](trace/fork.md) (Start of new routine)
- [End of Routine](trace/routineEnd.md)

## Toolchain

To record a trace, import 

```
import _ "goscgo"
``` 

and

build the program with 

```
-gcflags=all="-gosctrace"
```

e.g.

```go
./go build -gcflags=all="-gosctrace"
```

The run the compiled program.

For tests, run 

```
./go test ./... -gcflags=all="-gosctrace"
```

When the [toolchain](../gosc/) is used, this is done automatically.

## Trace local recording

To minimize the overhead introduced by the recording, we try to implement
the recording to be as routine local as possible. This includes storing
the trace element in a variable directly connected to the internal routine
implementation, preventing the necessity of synchronization mechanism (locks) when recording
elements. In total, we are able to do almost all of the recording
trace local, except for two global counter, namely the ids for the routines and
the [timestamps](#timestamp).

## Implementation

To record the execution trace local, we add an new variable `goscRoutineInfo`
into the [g struct](../goPatch/src/runtime/runtime2.go#L517).
This struct is automatically created for each routine by the runtime.
This variable (defined [here](../goPatch/src/runtime/gosc_routine.go#L28)), stores the routine id,
the maximum id of any element used in this routine and the
trace of this routine as a list of elements.

They are set [when the routine is started](../goPatch/src/runtime/proc.go#L5080).

To prevent those elements from being removed by the garbage collector and to
make it easy to collect them at the end, we store a reference to those elements
in a global map.

Before the runtime starts to run the main functions, multiple routines are created
and executed. They would always result in completely empty trace files, since
the recording only starts after the
[InitTracing](../goPatch/src/gosc/gosc_tracing.go#25) has been executed.
We therefore ignore those routines by setting there IDs to 0
and don't add there `goscRoutineInfo` into the global map.

To identify operations, that where executed on the same element, we assign
an ID to each element (channel, mutex, ...). Since most of those elements are internally
implemented as struct, we add a new field for this ID to the struct.

When a recorded function
is executed, we first check if the id was already set. If it was not, we set a new id. We want to minimize the number of
global counters we need to use. We therefor construct the new
id as $routine.id \cdot 1000000000 + routine.maxObjectId$ and then increase the maxObjectId field of the routine.
With this, we can guarantee that each object ID is unique.

For atomics, we use the memory position of the variable as id.

For each recorded operation a Pre and sometimes a Post function is implemented (multiple operations on the same type may share a Post function). The Pre function is called when the operation
is started, but before it executes. The Post function is called after it finished executing.

The calls to those functions are directly added into the implementations of
the recorded operations.

To get information about the specific implementation, see the descriptions
of the different recorded types as linked above.

To be able to reconstruct a global trace from those elements, we store a
time value for each pre and post. This values is created by an atomic counter,
that is increased every time a time value is requested. For more info, see [here](#timestamp).

The Pre function records that the operation was executed, the time when it was started, as well as all required information that are already available . The Pre function
also always records the location where the operation was recorded. This is (except for fork, see [here](./trace/fork.md#implementation)) done using the [runtime.Caller](https://pkg.go.dev/runtime#Caller) function.

The Post function then records the successful completion of the
operation as well as all information that was not available at the beginning, e.g. if a trylock was successful or not or which select case was executed.

For each operation we only store one element in the trace, representing
both the pre and post signal. The functions creating the pre signal
add a new element to the trace slice of the routine and return the index of this element
in the slice. The post functions take this index as an argument, to find
and update this value. We can determine if an operation had an pre, but no
post element, meaning the execution got stuck, by checking if the post timestamp is 0.
A post element without a pre element is not possible.

For elements that are executed directly without the possibility of the operations
being delayed by other operations (e.g. close on channel), the pre and post
signal may be recorded with the same function.

When the program execution has finished, it will create a folder `goscTrace`
in which it stores the trace files. For each routine, one trace file will be
generated. In it, each line contains the information about one recorded
event. The events are sorted by the time when the operations was executed.

We have decided not to record internal operations.\
The reason for this is that for most uses (e.g. bug analysis), they are not relevant
and unnecessarily increase the trace file size and the replay and recording time.
Additionally, they may be part of unpredictable operations like e.g. the garbage
collector, which would make the replay much more complicated to implements.\
If an internal operation is executed (meaning if the file path is in "goPatch/src/"),
it is ignored.

## Optimization and inlining

The build process of go can perform optimizations and inlining. This can lead
to problems, e.g. in the following program:

```go
func main() {
	c := make(chan int, 2)

	c <- 1
	c <- 2

	a := func(s string) {
		<-c                       // line 8
		println(s)
	}

	go a("A")                  // line 12
	go a("B")                  // line 13

	time.Sleep(time.Second)
}
```

The receive on the channel is at the same code position for both spawned
routines. But, because of inlining, they may be recorded at being at different
positions. Let's assume, that for the first routine in line 12, the compiler
applies inlining, but for the second routine in line 13, it does not. In this
case, the receive on `c` will in one case be recorded as being in line 12,
and in the other as being in line 8. This can lead to problems, especially
if during a replay of the recorded trace, where we use the line information
to connect the trace elements to the operations in the execution. Since we
cannot guarantee, that inlining will be applied consistently across recording
and replay, this may lead to problems.

To guarantee, that the recorded code positions are always the positions
where the operation actually occurs, optimization and
inlining is disables for running the recording, and also when running a replay.
This is done by setting the `-gcflags="all=-N -l"` when building a program or running a test.

## Timestamp

To reconstruct the global trace from the recorded local traces, we need a consistent
way to apply timestamps to all trace elements.

To synchronize the thread-local traces, we use a global atomic counter.
This counter is increased every time an operation requests a time value.

We tried to implement the timestamps routine local, using time values provided by the operating systems. For this, we looked at two possible time functions in the go runtime. Both of them have there problems.

The `cputicks` function is described by the go tracer team as\
"On most platforms, this function queries the CPU for a tick count with a single instruction. (Intuitively a "tick" goes by roughly every CPU clock period, but in practice this clock usually has a constant rate that's independent of CPU frequency entirely.) [...] Unfortunately, many modern CPUs don't provide such a clock that is stable across CPU cores, meaning even though cores might synchronize with one another, the clock read-out on each CPU is not guaranteed to be ordered in the same direction as that synchronization. This led to traces with inconsistent timestamps." [^1]

Additionally, the signature of the implementation notes:\
"careful: cputicks is not guaranteed to be monotonic! In particular, we have noticed drift between cpus on certain os/arch combinations. See issue 8976." [^2]

For this reason, we cannot guarantee that a trace with this form of timestamp reflects the actual execution order.

The nanotime() returns a time value from the operating system. It it consistent over all routines.

Here the problem lyes in the precision of those counters. Atomic operations like Load and Store only take a few nano seconds to execute. Assume we have the following program snippet.

```go
var a atomic.Uint64
go func() {
  a.Store(1)
}()
go func() {
  x := a.Load()
}
```

If the following code in the second routine depends on the value of `x`, it is necessary
to determine the exact order of the Load and Store operations, to get an accurate analysis and especially replay.
If we use the nanosecond counter as a timestamp and the precision of it is to small, this may result in
unclear orders.

For Linux based systems, this uses `CLOCK_MONOTONIC` (normally defined in `/usr/include/time.h`) with a precision of, in most cases, `1 ns` (can by checked by running `clock_getres(CLOCK_MONOTONIC, &res); printf("%ld\n", res.tv_nsec)` as a c program).
It is therefore enough. The problem, lies in the use in windows and macos. For windows, the `QueryPerformanceCounter` is used. According to [^3], this only has a resolution of about `100 ns`. This could lead to a situation, where the two atomic
operations receive the same timestamp value, resulting in an incorrect replay.

Even though this would allow us to get a routine local, accurate time signal, at least for linux systems, our experiments show, that the executing using
the `nanotime` function is slightly slower than the use of a global atomic variable, even for programs where a lot of recorded
operations are executed at the same time. Since the main reason for routine local recording is performance, it was not useful to use the `nanotime` function.

We therefore simply use a global, atomic
variable (atomic.Int64) initialized to 0, which is incremented each time a new
timestamp is requested.

## Writing

When the main function terminates, it calls the [FinishTracing](../goPatch/src/gosc/gosc_tracing.go#65) function. This will collect all the local traces and write them into files.
We ignore all internal routines, that where created and run before the main
routine started to execute.

The trace files are stored trace local, meaning one file per trace in a folder called
`goscTrace_[number]` inside the `goscResult` folder.
The files are called `trace_[id].log`. In each file the elements executed
in this routine are stored, one element per line sorted by the pre counter.
The elements stored for each operations are described in the
explanations for the different types linked above.

Additionally a `trace_info.log` file is created with some additional infos,
e.g. whether the program terminated normally or because of a panic.

[^1]: M. Knyszek. "Execution tracer overhaul". https://github.com/golang/proposal/blob/master/design/60773-execution-tracer-overhaul.md (Accessed 2025-03-29)\
[^2]: [runtime/cputicks.go](../goPatch/src/runtime/cputicks.go#L11)\
[^3]: S, White et al. "Acquiring high-resolution time stamps". https://learn.microsoft.com/en-us/windows/win32/sysinfo/acquiring-high-resolution-time-stamps#resolution-precision-accuracy-and-stability (Accessed 2025-03-29)
