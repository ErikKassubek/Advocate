# AdvocateGo

![Arechitecture](./doc/img/architecture/architecture_new.png)

## What is AdvocateGo

AdvocateGo is an analysis tool for concurrent Go programs.
It tries to detects concurrency bugs and gives diagnostic insight.

AdvocateGo tries to detect the following situations:

- A01: "Actual Send on Closed Channel",
- A02: "Actual Receive on Closed Channel",
- A03: "Actual Close on Closed Channel",
- A04: "Actual Close on Nil Channel",
- A05: "Actual Negative Wait Group",
- A06: "Actual Unlock of Not Locked Mutex",
- A07: "Actual Non-Cyclic Blocking Bug",
- A08: "Actual Cyclic Deadlock with Mutex",
- A09: "Actual Concurrent Receive on Same Channel",
- A10: "Actual Cyclic Deadlock with Mutex and Channel
- P01: "Possible Send on Closed Channel",
- P02: "Possible Negative WaitGroup cCounter",
- P03: "Possible unlock of not locked mutex",
- P04: "Possible Cyclic Deadlock with Mutex",
- P05: "Possible Cyclic Deadlock with Mutex and Channel",
- L00: "Leak with unknown cause",
- L01: "Leak on a channel",
- L02: "Leak on nil channel",
- L03: "Leak on select",
- L04: "Leak on sync.Mutex",
- L05: "Leak on sync.WaitGroup",
- L16: "Leak on sync.Cond",

Additionally it is able to record and deterministically replay
executions of concurrent GO programs.

## Modes

Advocate provides 4 different modes:

- record: record the execution of a program or test into a trace
- replay: given a trace file, execute a program in such a way, that it follows the trace
- analysis: record a program and analyze the recorded trace to detect potential concurrency bugs. If a potential bug is found, rewrite the trace in such a way that the bug is triggered and replay this trace to confirm that the bug is possible.
- fuzzing: Apply different fuzzing approaches to increase the reach of the analysis.

## Usage

For an explanation on how to use advocate, see [here](./doc/usage.md).

> [!IMPORTANT]
> advocate is implemented for go version 1.27.
> Make sure, that the program does not choose another version/toolchain and is compatible with go 1.27.
> The output `package advocate is not in std` or similar indicates a problem with the used version.
> It can help to shorten the go version number in the go.mod file of the analyzed program from go 1.27.x to go 1.27.

## Documentation

A detailed description of how advocate works can be found in the [doc](doc) folder.
