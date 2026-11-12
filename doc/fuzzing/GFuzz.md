# Select Based Fuzzing

The select based fuzzing is based on [GFuzz](../relatedWorks/PaperAndTools/Fuzzing/GFuzz.md).

## Idea

GFuzz works on the assumption, that selects are the main points, where
execution paths in concurrent Go programs diverge. It therefore
mutates the use of those selects.

The main idea is to set a preferred case for each select execution.
The tool will, if possible, force the program to execute this preferred case,
even if another case would have been chosen naturally. In doing so,
new and maybe unlikely, but still possible paths are explored.

Let's look at the following example to illustrate the Idea

```go
func main() {
  c := make(chan int)
  d := make(chan int)

  go func() {
    // some code
    time.Sleep(100 * time.Millisecond)

    c <- 1
  }

  go func() {
    // some code
    time.Sleep(300 * time.Millisecond)

    d <- 1
  }

  select {
  case <-c:
  case <-d:
    codeWithBug()
  }
}
```

The program consists of a select with two cases, one of which leads
to a program path with a bug.
Let's assume, that due to the surrounding code (here simulated with sleep),
it is very likely, that the case of the c channel is chosen. Since no
dynamic analysis can detect bugs in codes, it did not run, this makes it
very hard to detect the bug in this program.

Given this program, GFuzz will now create runs, where one of the cases is
the preferred case. When executing the mutation where the case containing
the receive on d is the preferred case, it will execute the program containing
the bug, giving the program the chance to detect the bug (GFuzz only detects
actual bugs, but a more detailed analysis may be able to detect
bugs that are possible but not executed).

<!-- The main fuzzing loop is implemented as follows.\
The fuzzing contains a queue with all the mutations to run.
When the program or test is run for the first time, it will be run in the normal recording mode of the toolchain. Otherwise it will pop a mutation
from the queue and run this mutation. This is done by storing the
relevant information in a file called `fuzzingData.log` and adding a
different header to the program/test. This will run the mutation (see [Running a mutation](#running-a-mutation)) and record the trace for this mutation.\
Then the analyzer will be applied to the recorded trace to find potential bugs. If replays are possible, they will be performed here as well.\
Afterwards the fuzzing will parse the internal trace and calculate all values required to determine whether the run was interesting and if so, how many new mutations should be created (see [GFuzz](#gfuzz)).\
If the run was interesting, the new mutations are created. For this, a [flip probability](#flip-probability), meaning the probability that a select changes its preferred case is calculated.\
For the selects that are flipped, a case, including the default, is selected randomly as the new preferred case, making sure that the new preferred case is not equal to the last preferred case.\
Different to the original GFuzz implementation, which needs to run in to a bug to detect it and therefor may need to run the same mutation multiple times, gosc can also detect a bug if it does not occur directly. For this reason, the same mutation may only be run a limited number of times (maybe even just once). We therefore check if the created mutation has been added to the mutation queue before and if it has how often it has been added and only add the new mutation if the number of runs for the mutation does not exceed a set limit.

This loop is repeated until the mutation queue is empty. Additionally a maximum number of runs or a maximum time can be set. -->

## Implementation

<!-- - It is possible to determine all values needed to determine how interesting a run is from the trace
- Checking if a select case is possible using the HB relation would only make sense until the program run first executes a select, where a different channel is used than in the last recording. After that, the HB relation is no longer valid and can therefore not be used to determine, if a select case has a possible partner.
- The score calculation could include information from the HB relation. E.g., a run where many not executed select cases have a possible partner, could be more interesting. -->

### Determine whether the run was interesting

Since the number of possible combinations of preferred cases can be very large,
it is useful to focus on interesting runs and mutate them to get new runs.

A run is interesting, if one of the following conditions is met. The underlying information need to be stored in a file for the following runs.

- The run contains a new pair of channel operations
  - All pairs of channel operations (send-recv) must be stored
- If an operation pair's execution counter changes significantly from previous order.
  - For each operation pair determine the average number of executions per run
  - Determine a run to be interesting, if the number diverges from this average by at least 50%
  - Details changed from original GFuzz paper, because definition in GFuzz paper makes no sense.
- If a new channel operations (creation, close, not close) is triggered for the first time
  - We must be able to identify each channel
    - Add a trace element on create -> channels can be identified by path of create and channel ID
  - We must store all channels ever created
  - We must store for all channels that have ever been created, whether they have been ever closed/not closed or both
- If a buffered channel gets a larger maximum fullness than in all previous executions
  - For each channel we must store the maximum fullness over all runs
- Additionally (not in the original GFuzz), we determine a run interesting if
  a select case, that has never been selected before is selected

### Determine the score

For GFuzz, we need to extract the following information from the trace:

- CountChOpPair_i: For each pair i of send/receive, how often was it executed
- CreateCh: How many distinct channels have been created
  - Can be determined based on channel id
- CloseCh: Number of closed channel
  - Count close operations
- MaxChBufFull: Maximum fullness for each buffer
  - Each buffered channel info in the trace contains the current qSize. Pass all send and get the biggest

The score is then calculated as
$$score = \sum\log_2 CountChOpPair + w_1 \cdot CreateCh + w_2 \cdot CloseCh + s_3 \cdot \sum MaxChBufFull$$
where $w$ are scaling factors.

### Creating mutations

From the score of an interesting mutation we get the number of mutations to create
by calculating $$mut = \lceil 5 * score/maxScore \rceil.$$

For each of the mutations we do the following based on the recorded run.
We traverse all selects. We then decide whether the select should flip its
preferred case (see [flip probability](#flip-probability)). If the select will
flip its preferred case, it will chose one of the cases randomly.

### Flip Probability

The flip probability is the probability that a single select in the fuzzingData will change its preferred case compared to the previous mutation. If its set to high, the mutation mechanism basically becomes completely random. If its to small, the program will result in the same mutation being created over and over again. For now the probability is calculated as $$P = max(0.1, 1 - (1 - 0.99)^{1/numSel})$$ where $numSel$ is the number of selects in the previous mutation. This is selected in such a way, that the probability of at least one of the selects to flip its preferred case is at least $99\%$, but the probability for each individual select to get flipped is at least $10\%$. This may be changed based on experimental results.

### Running a mutation

When executing a mutation, we must make sure, that first the select only
waits on the preferred case and only falls back to the whole select if
this case cannot be executed (timeout). The original implementation and
the implementation used in Gosc are different. While for the original
GFuzz, the program code is directly implemented, in Gosc the runtime
is changes to be able to prefer a case.

#### Original GFuzz implementation

In the original GFuzz, each select in the program code is directly instrumented.

If as an example we have the following select:

```go
select {
case <-c:
  someCodeC()
case <-d:
  someCodeD()
}
```

it is instrumented to look as follows

```go
switch FetchOrder() {
case 0:
  select {
  case <-c:  // first case preferred
    someCodeC()
  case <-time.After(T):  // timeout
    select {  // copy of full, original select
    case <-c:
      someCodeC()
    case <-d:
      someCodeD()
    }
  }
case 1:
  select {
  case <-d:  // second case preferred
    someCodeD()
  case <-time.After(T):  // timeout
    select {  // copy of full, original select
    case <-c:
      someCodeC()
    case <-d:
      someCodeD()
    }
  }
}
```

FetchOrder returns the index of the preferred case in the select.
Based on this, a select with exactly two cases is run. The first case
is the preferred case, that is associated with this index, including the body
of this case. The second is a `time.After(T)` case, where `T` is a predefined
time for a timeout. If this timeout case is triggered, before the preferred
case has executed, a copy of the original, full select will be called.

#### Gosc GFuzz

In the here implemented GFuzz version, we do not instrument each select in the
code, but instead change the implementation of the select in the runtime.

We create a file `fuzzingData.log` containing for each select a list the preferred cases.
The selects are identified by its position in the code base. When initializing the mutation run, the file is read in and stored in a `map[string][]int` `fuzzingSelectData` with an entry per select containing the list of preferred indexes. Additionally there is a map `fuzzingSelectDataIndex: map[string]int`, storing for each of the select the index of the next preferred case in `fuzzingSelectData`. Then a select is supposed to be executed, assuming it is not an internal select, the preferred case is retrieved via `fuzzingSelectData[selPos][fuzzingSelectDataIndex[selPos]]` followed by `fuzzingSelectDataIndex[selPos]++`.

The select is implemented in the [selectgo](../../goPatch/src/runtime/select.go#L123) function. This
function is split into two parts. On of them is the modified version and is called if the fuzzing is active.The other part is the original select and
is called if fuzzing is not active or if the fuzzing
select was terminated by a time out.

```go
func selectgo(cas0 *scase, order0 *uint16, pc0 *uintptr, nsends, nrecvs int, block bool) (int, bool) {
  fuzzingEnabled, fuzzingIndex, fuzzingTimeout := GoscFuzzingGetPreferredCase(2)
  if fuzzingEnabled {
    if ok, i, b := fuzzingSelect(cas0, order0, pc0, nsends, nrecvs, block, fuzzingIndex, fuzzingTimeout); ok {
      return i, b
    }
  }

  return originalSelect(cas0, order0, pc0, nsends, nrecvs, block)
}
```

The select in its original implementation contains 3 passes. In the first pass the select loops over all cases and checks whether on of them has a
possible already waiting partner. If this is the case, it is immediately executed and the select finishes. If not, and there is a default case, it is executed.\
If there is no default case and non of the cases could be executed, all case operations are enqueued into the queues of the corresponding channels.\
Then the routine is parked using `gopark` until it is woken up by another routine that wants to communicate with one of the channels in the cases.\
The communication is then executed.\
In the third pass, all case operations that have not been executed are removed from there respective queues.

The implementation for the fuzzing is changed in the following way:\
For the first pass, only the case where the internal `casi` is equal to the `fuzzingIndex` is checked for possible communication partners.
If this case can directly communicate, it is executed and the select is done.
If the default case is selected as the default cause it is directly executed.\
If neither of those happened, the preferred case is enqueued into its waiting queue. Then a modified park
function is called.

```go
func goparkWithTimeout(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceReason traceBlockReason, traceskip int, timeout int64) {
  mp := acquirem()
  gp := mp.curg

  // Setup timer if timeout is non-zero
  if timeout > 0 {
    go func() {
      sleep(float64(timeout))
      goready(gp, traceskip)
    }()
  }

  // Run the original gopark logic
  ...
```

This park method will first start a go routine with a timer. If the timer has run out, it will wake up the routine regardless of whether it has found a partner. Then the normal park function is run. If the routine was woken up because the enqueued channel found a partner, the select continues as in the unmodified version. If it was woken up by the timeout, it will remove the enqueued channel operation, do some clean up and return from the modified select returning $ok = false$. In this case the unmodified select will be run as can be seen in the `selectgo` func above.

## Improvement over original GFuzz

We can improve the original GFuzz by using the HB information.

The improvements over GFuzz are mainly in finding interesting runs faster and
therefore needing to run fewer runs.

The main improvement is, as always, that
we do not need to run each mutation multiple times, hoping we run into a
bug. Running a mutation once (or a small number of times) is enough,
since the HB analysis can detect possible but not executed bugs. This
also means, that bugs, which are theoretically possible (and can maybe even be replayed),
but which are so unlikely, that they are not detected by simply running the
program a limited amount of runs, can be detected.

GFuzz only mutates the selects. But for some reason, neither the decision
whether a run is interesting, nor the calculation of the number of mutations
depend on the selects in the recorded code. But it should be obvious, that the
number of selects, and especially
the number of select cases with possible partners influence how likely a
mutation of selects will result in new paths being explored.

This is done by extending the score function to be
$$score = \sum\log_2 CountChOpPair + w_1 \cdot CreateCh + w_2 \cdot CloseCh + s_3 \cdot \sum MaxChBufFull$$

In the original GFuzz, the two following examples, assuming in both the
communication on `x` is executed, result in the same
decision about whether the runs are are interesting and in the same number
of mutations (given that the rest of the program is identical)

```go
a := make(chan int)
x := make(chan int)

go func() {
  x <- 1
}

go func() {
  a <- 1
}


<-x
// do stuff
```

```go
a := make(chan int)
x := make(chan int)

go func() {
  x <- 1
}

go func() {
  a <- 1
}

select {
  case <-x:
    // do stuff
  case <-a:
    // potential bug
  ...
}
```

But it should be obvious, that a mutation of the second program
is more useful than a mutation of the first problem. By including the
number of select cases with possible partners, we are more likely to create
mutations which lead to new and interesting runs.

Additionally, the changes in the calculation of the score and the
calculation of the mutations may result in interesting runs been found faster
and preventing to many mutations that do not change the actual mutation from
being created.

```go
go func() {
  a <- 1
}()

go func() {
  z <- 1
}()

select {
  case <-a:
    //does not contain bug
  case <-b:
    //does not contain bug
  ...
  case <- z:
    // does contain bug
  default:
    // does not contain the bug
}
```

This run contains a select with many different cases. Lets assume, that based
on the HB analysis only the cases with channel `a` and channel `z` actually
have a partner and that the case with channel `a` has been used in the recorded run.
In a mutated run, GFuzz will select each of the cases with the same probability,
whether they have a potential partner or not. This means, that in many cases,
the mutation run will result in the timeout and therefore no new interesting
behavior will be triggered. By choosing cases with possible partners with a higher
probability (3:1), they will be chosen more often, resulting in fewer runs being
effected by timeout. This can reduce the number of runs needed to explore the
possible paths.
