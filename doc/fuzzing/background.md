# Fuzzing For Go Concurrency Analysis

## Background and Motivation

We consider a specific program run represented by a trace.
Based on this trace we employ predictive methods (based on HB) to check
for a reordering of the trace that leads to a bug.

The effectiveness of this approach for bug finding relies on the assumption
that we have a sufficient number of unit tests that exercise all program parts.
For example, there may be a bug in function A but we miss this bug if
function A is never executed.

In the concurrent context, it is not only about exeucting all program parts
for various input values. We also need to consider alternative schedules
that lead to concurrency bugs. For example, consider deadlocking situations etc.

Fuzzing (aka fuzz testing) is an approach that influences the program flow (by mutation of program inputs, scheduling of threads, ...) to exhibit bugs that might be missed.

[GFuzz](https://github.com/system-pclub/GFuzz)
and [GoPie](https://github.com/CGCL-codes/GoPie) mutate the execution
of concurrent Go programs to exhibit bugs.

The recent work [Sound Predictive Fuzzing for Multi-threaded Programs](https://www.computer.org/csdl/proceedings-article/compsac/2023/269700a810/1PhCMNJZYlO)
in the context of C/C++
seems to be the first work that combines fuzzing with trace-based analysis methods. Check out the [talk](https://www.computer.org/csdl/video-library/video/1PrE27npamY) that gives a nice overview.

We want to do something similar but in the context of Go.
Combine fuzzing and trace-based analyses for the analysis of concurrent
Go programs.


## How to get started


### Combine GoPie and GoscGo

Would need to incorporate GoPie patches into GoscGo.

How much effort is required?

### Go comes with a "built-in" [fuzzer](https://go.dev/doc/security/fuzz/) since Go 1.18.

Here are some examples.

Let's take a look at fuzzing via some Go examples.

#### Even

The following function is meant to return true if the number is even.
Otherwise, we return false.

```go
func Even(i int) bool {
	if i > 100 {
		return false
	}
	if i%2 == 0 {
		return true
	}

	return false

}
```

There is an obvious bug (for numbers above 100).
We may miss this bug if we only consider a limited number of test inputs
(that are all below 100).

As we can see below, thanks to fuzzing we can quickly locate the bug.


##### Sources

Put the following files in a separate folder.

even.go:

```go
package main

import "fmt"

func Even(i int) bool {
	if i > 100 {
		return false
	}
	if i%2 == 0 {
		return true
	}

	return false

}

func main() {

	fmt.Printf("%d => %t", 5, Even(5))

	fmt.Printf("%d => %t", 0, Even(0))

}
```

even_test.go:

```go
package main

import "testing"


type testPair struct {
	input    int
	expected bool
}

func TestEven(t *testing.T) {

	testcases := []testPair{
		{5, false}, {0, true}, {50, true}}

	for _, tc := range testcases {
		res := Even(tc.input)
		if res != tc.expected {
			t.Errorf("isEven: %d => %t, want %t", tc.input, res, tc.expected)
		}

	}

}


func FuzzEvent(f *testing.F) {
	testinputs := []int{5,0,50}

    for _, tc := range testinputs {
        f.Add(tc)  // Use f.Add to provide a seed corpus
    }
    f.Fuzz(func(t *testing.T, in int) {
        res := Even(in)
        res2 := Even(in+1)
			if res == res2 {
			t.Errorf("Fail: %d => %t, %d => %t", in, res, in+1, res2)
        }
    })
}
```

Sample runs.

```
> go mod init example/even
>
> go test . --fuzz=Fuzz
fuzz: elapsed: 0s, gathering baseline coverage: 0/4 completed
failure while testing seed corpus entry: FuzzEvent/b61b643957dc4b40ef87b96e142889311e42671d0964206e68b63839d9ff72cf
fuzz: elapsed: 0s, gathering baseline coverage: 3/4 completed
--- FAIL: FuzzEvent (0.01s)
    --- FAIL: FuzzEvent (0.00s)
        even_test.go:37: Fail: 104 => false, 105 => false

FAIL
exit status 1
FAIL	example/even	0.247s
```



#### Channel

Here's another example that makes use of channel-based concurrency.
The program below contains a deadlock.
The deadlock does not manifest itself for all program runs
but only under a specific schedule.

The deadlock is independent of the input value.
But it seems that if we wait long enough that the fuzzer eventually
uncovers the deadlock.

##### Sources

Put the following files in a separate folder.

channel.go:

```go
package main

// Deadlock possible!
func Runner(i int) {

	ch := make(chan int)

	go func() {
		ch <- i
	}()

	go func() {
		<-ch
	}()

	ch <- i

}

func main() {

	for i := 0; i < 50; i++ {
		Runner(i)
	}

}
```

channel_test.go:

```go
package main

import "testing"

func FuzzRunner(f *testing.F) {
	testinputs := []int{5}

	for _, tc := range testinputs {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}
	f.Fuzz(func(t *testing.T, in int) {
		Runner(in)

	})
}
```

Sample runs.

```
> go mod init example/channel
>
> go run channel.go
>
> go test . --fuzz=Fuzz
fuzz: elapsed: 0s, gathering baseline coverage: 0/22 completed
fuzz: elapsed: 0s, gathering baseline coverage: 22/22 completed, now fuzzing with 10 workers
fuzz: elapsed: 3s, execs: 41407 (13798/sec), new interesting: 1 (total: 23)
fuzz: elapsed: 6s, execs: 41407 (0/sec), new interesting: 1 (total: 23)
fuzz: elapsed: 9s, execs: 41407 (0/sec), new interesting: 1 (total: 23)
fuzz: elapsed: 11s, execs: 41980 (350/sec), new interesting: 1 (total: 23)
--- FAIL: FuzzRunner (10.64s)
    fuzzing process hung or terminated unexpectedly: exit status 2
    Failing input written to testdata/fuzz/FuzzRunner/a4d1cc446bc342b8e28355319c67fde7264e5b2c398f91977758641d6ee4ca54
    To re-run:
    go test -run=FuzzRunner/a4d1cc446bc342b8e28355319c67fde7264e5b2c398f91977758641d6ee4ca54
FAIL
exit status 1
FAIL	example/channel	10.875s
```


Comment.
The fuzzer does not explicitly issue that there is a deadlock (all threads are blocked). But "hung" is a likely indication for a deadlock.

#### Race

##### Sources

Put the following files in a separate folder.

race.go:

```go
package main

import "fmt"

// import "time"
import "sync"

// Data race possible
func Race(i int) {
	var m sync.Mutex
	x := 1

	go func() {
		m.Lock()
		x = 2
		m.Unlock()
	}()

	//	time.Sleep(time.Second)
	x = 3
	m.Lock()
	fmt.Printf("\n%d", x)
	m.Unlock()

}

func main() {

	for i := 0; i < 10; i++ {
		Race(i)
	}

}
```

race_test.go:

```go
package main

import "testing"

func FuzzRace(f *testing.F) {
	testinputs := []int{5}

	for _, tc := range testinputs {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}
	f.Fuzz(func(t *testing.T, in int) {
		Race(in)

	})
}
```


Sample runs.

```
> go mod init example/race
>
> go run -race race.go
...
many, many times, no race warning is issued
```

The reason why no race warning is issued is as follows.

1. The main thread starts and highly likely will run all code to completion

2. As part of the main thread, we create a new thread.

3. This thread might not get started at all (=> no data race)

4. If this thread gets started, it is highly that the thread only starts after the main thread has executed all its statement

5. As the (FastTrack style) data race predictor underlying go-race maintains the order among critical section, go-race is unable to detect the data race.

6. If we include the sleep statement, the main thread will run after the new thread. Then, go-race is able to detect the data race.

Let's try Go fuzzing.

```
> go test . --fuzz=Fuzz -race
fuzz: elapsed: 0s, gathering baseline coverage: 0/8 completed
fuzz: elapsed: 0s, gathering baseline coverage: 8/8 completed, now fuzzing with 10 workers
fuzz: elapsed: 0s, execs: 4115 (14303/sec), new interesting: 1 (total: 9)
--- FAIL: FuzzRace (0.29s)
    --- FAIL: FuzzRace (0.01s)
        testing.go:1319: race detected during execution of test

    Failing input written to testdata/fuzz/FuzzRace/acdea50f284db86d7252b2b70923c57f7dd7486a517f5eb29619998b2a00d164
    To re-run:
    go test -run=FuzzRace/acdea50f284db86d7252b2b70923c57f7dd7486a517f5eb29619998b2a00d164
FAIL
exit status 1
FAIL	example/race	0.569s
```

Interesting. Data race detected! How?
Does Go fuzzing tap in the Go run-time and mutate thread scheduling?

Not sure. A more likely explanation (for detecting the data race) is that
the instrumentation done by Go fuzzing affects thread scheduling
and thus we are able to detect the data race.

#### Summary

Fuzzing in Go seems well supported.

We can effectively uncover the bug in `Even`.

However, fuzzing in Go can only applied to test units of certain input type.
What if our test unit expects a custom type?

Fuzzing works for all Go programs including programs that make use of concurrency.
It seems that we can uncover the deadlock bug in `Runner` and the data race bug in `Race'. But these are simple examples.
In case of the channel example it took the fuzzer a long time (10secs+) till the Go fuzzer reports "hung". It is not clear if the Go fuzzing is aware of deadlocks. See the following [issue](https://github.com/golang/go/issues/48591).
It also remains to be seen how effective Go fuzzing is for concurrent programs where there are many more alternative schedules to consider.

## Summary of related works

### GFuzz and GoPie (Go)

GFuzz only mutates select. GoPie improves on GFuzz by mutuating message orders.

### Greybox Fuzzing for Concurrency Testing (C/C++)

Modify AFL.

Main new idea seems to identify traces as equivalent if the the "last write" relations are the same.
In the paper, this is referred to as the "reads-from" relation.
So, the goal is to explore schedules where a read is connected to a different write.

### Sound Predictive Fuzzing for Multi-threaded Programs (C/C++)

https://www.computer.org/csdl/proceedings-article/compsac/2023/269700a810/1PhCMNJZYlO

Combine the SeqCheck method described
in [Sound and Efficient Concurrency Bug Prediction](https://web.cs.ucla.edu/~palsberg/paper/fse21.pdf) and AFL++.

They seem to use a different "equivalence" criteria among traces
where the two traces are different if they contain a distinct sequence of critical sections.

They argue to consider only lock/unlock operations to select new mutation.
No need to consider the many read/write operations (like "Greybox").
