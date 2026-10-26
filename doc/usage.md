# How To Use Advocate



## Download

Download or clone the repository into a local folder.


If you are using Windows, make sure that any automatic conversion of line endings (LF -> CRLF) is turned off. For git, this can be done with 

```shell
git config core.autocrlf false
```


## Build

Before Advocate can be used, it must first be build.

There are two elements that need to be build.

### Runtime

To run the recording and replay for Go, a modified version of the Go runtime
has been provided. It can be found in the [go-path](../goPatch/) folder.

Before it can be used, it needs to be build. To do this, move into
[go-path/src](../goPatch/src/) directory and run the

```shell
./src/make.bash
```

or 

```shell
./src/make.bat
```

script. This will create a go executable in the `bin` directory.

### Advocate

Additionally, the advocate program needs to be build. This is a standard Go
program. To build it, move into the [advocate](../advocate/) directory
and build it with the standard

```shell
go build
```

command. This will create an `advocate` executable, which will be used to
run all recordings, replays, analysis and fuzzing.

Advocate provides an optional gui. For this, it uses [fyne](https://fyne.io/). To build advocate with gui, the required dependencies for fyne must be installed. 
See [here](https://docs.fyne.io/started/quick/) for details.

If you want to build advocate with gui, you have to enable it by building the program with 

```shell
go build -tags gui
```

In this case fyne is not required.


## Docker

We also provide a docker file to create the environment.

To build the docker file, run

```shell
docker build -t advocate-app .
```

To run the analysis or fuzzing on a program, you can call the following:

```shell
docker run --rm -it \
  -v <pathToProg>:/prog \
  advocate-app [mode] -path /prog [args]
```

e.g.

```shell
docker run --rm -it \
  -v /home/erik/progToTest:/prog \
  advocate-app fuzzing -path /prog
```

For the modes and args, see [usage](#usage).
Note that the -path argument has already been set and does not need to be set again.

## Usage

All modes of advocates are started and controlled using the [advocate](../advocate/)
program. This program implements multiple modes:

- [Gui](#mode-gui)
- [Recording](#mode-recording)
- [Replay](#mode-replay)
- [Analysis](#mode-analysis)
- [Fuzzing](#mode-fuzzing)

### Help

To get an overview over the possible modes and arguments, you can run

```shell
./advocate -help
```

### Mode: Gui

This modes starts a graphical interface allowing you to run all other modes.
No other flags are required. 

This is not available, if advocate was build with `-tags nogui`.

### Mode: Recording

This mode allows us to record all relevant concurrency operations of a given
program. Running this mode will run a given program or test and create a
trace of all the elements executed in the program.

For more information about the recording and the created trace, see [here](./recording.md).

To run a recording, we use the advocate program implemented [here](../advocate/).
To record a program, run

```shell
./advocate record [args]
```

This will run the program or tests and create the traces. They are placed
into a folder called `advocateResult`.

Normally, the print outputs of the executed programs are not printed to
the terminal, but printed into an `output.log` file. To show them on the terminal,
you can set the `-output` flag.

#### Tests

If you want to execute and record the unit tests of a program, you need to
specify the following arg:

- `-path [pathToTests]`

This provides the path to the tests that should be recorded.
The path should point to the folder containing the tests.

As a default, this will record all tests
in this folder (for each test a separate trace will be created).\
If you only want to execute a specific test, you can specify this test with

- `-exec [testName]`

Be aware that this will record all tests with this name, meaning if multiple
tests share the same name, all of them will be executed.

An example command would be

```
./advocate record -path ~/program/testFolder/ -exec TestOne
```

#### Program

If you want to record the program itself, the following args need to be set

- `-main`
- `-path [pathToMainFile]`

Here, main tells the program to run the main function instead of the tests.
Path should point to the file containing the main function of this program.

Go will try to determine the executable name of the program from the `go.mod`
file. If advocate is unable to find this file, it needs to be set manually
using

- `-exec [executableName]`

Please note, that programs that cannot be directly be build using `go build`
cannot be recorded (the tests can still be recorded in this case).

A possible command would therefore be

```
./advocate record -main -path ~/program/main.go -exec progName
```

### Mode: Replay

The replay mode allows us to replay a previously recorded trace.

This can be done by calling

```
./advocate replay [args]
```

The following args are required

- `-path [pathToProg]`: path to the folder containing the tests of the main file
- `-trace [pathToTrace]`: path to the folder containing the trace files to be replayed

If the `-main` flag is not set, meaning some tests are replayed, and there
is more than one test in `pathToProg`, `-exec [testName]` must be set to specify
the test, the trace belongs to.\
For main, `-exec [executableName]` should only be set if the go.mod file cannot
be found.

Possible command would therefore be

```
./advocate replay -path ~/program/testFolder/ -trace ~/traceFolder
./advocate -main replay -path ~/program/main.go -trace ~/traceFolder
```

Please not, that the trace folder should not be inside the `AdvocateResult` folder.
This means, if you want to replay a given recording, you first need to copy it
to a location outside the `AdvocateResult` folder.

Please note, that the replay relies on the program code not being altered
between recording and replay. Each change, even on non-concurrency elements
can cause the replay to fail.\
Additionally, all non-concurrency indeterminism, like random numbers
or unpredictable api calls can cause the replay to get stuck. For more info
see [here](./replay.md#things-that-can-go-wrong).

### Mode: analysis

The analysis mode is the main mode to analyzer tests. It will run the program
or test, [record](./recording.md) the trace, [analyze](./analysis.md) and analyze
it. It a potential bug was found and it is possible, the mode will [rewrite](./rewrite.md)
the trace in such a way, that it should contain the potential bug and
[replay](./replay.md) it, trying to trigger the bug.

For the (specified) tests or the main function, it will run the program and
record the trace, analyze the trace and, if something was found,
create a trace that should trigger the found bug and replay this trace,
to confirm the bug.

It can be run with

```
./advocate analysis [args]
```

The arguments for `-path` and if necessary `-main` are required. They are
defined the same way as in for the [recording](#mode-recording) and [replay](#mode-replay)
modes.

If `-path` does not point to the root of the program, `-root [path]` must be set. 

The default behavior is to run all analysis scenarios. You can select to run
only certain scenarios to by setting

- `-scen [scenarios]`

with the following possible scenarios:

- `s`: Send on closed channel
- `r`: Receive on closed channel
- `w`: Done before add on waitGroup
- `n`: Close of closed channel
- `b`: Concurrent receive on channel
- `l`: Leaking routine
- `u`: Unlock of unlocked mutex
- `c`: Cyclic deadlock (resource deadlocks)

To select multiple by adding them together, e.g.

```
  -scen sc
```

to run the analysis for send on closed and cyclic (resource) deadlocks.\
If `-scen` is not set, all scenarios will be searched for.

While running, the analyzer will create a `advocateResult` folder. In it, it will create on
folder for each of the analyzed tests. In this folder it will create a file
for the output of the program runs, as well as two files showing an
overview over all detected bugs. Additionally, it will create a bug folder.
This folder contains one file for each of the found bugs, detailing the
type and position of the bug and information about the replay (if performed).

An example command would be

```
./advocate analysis -path ~/pathToProg/progDir/main.go -prog progName -main
```

to run the analysis on the main function of a program, or

```
./advocate analysis -path ~/pathToProg/progDir/ -prog progName -scen c -exec TestOne
```

to analyze the test `TestOne` in the given path, only checking for cyclic (resource) deadlocks.

The analysis will try to find the following situations:

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
- L..: "Leak" (Blocked but not necessarily finally blocked routine),

Some of them are only considered warnings. To ignore them, you can set `-noWarning`.

The analysis will try to rewrite and replay found potential bugs. To disable this,
you can set `-noRewrite`.\
The default behavior is to not replay bugs that have already been replayed successfully.
To still replay them, you can set `-replayAll`.

The traces can become very large. When using advocate for many tests, or multiple
times, this can lead to a large amount of data being stored in the trace files.
For this reason, advocate can delete the trace files, as soon as the analysis
has finished. To activate this, you can set `-deleteTrace`.

The [Go Memory-Model](https://go.dev/ref/mem#chan) does not specify, that
channels behave as a FIFO queue. Since in practice they are implemented as such,
the analysis assumes that they behave this way. If you do not want to use this
assumption, you can disable it using `-ignCritSec`.

Additionally, the used happens-before modes assumes, that critical sections
(mutex) conform to a happens before relation, meaning they cannot be reordered.
This may be too strong of an assumption in some case. To ignore this relation
you can set `-ignCritSec`.

If you do not want to perform the happens-before analysis, but only check
for actually occurring panics or leaks, you can set the `-onlyActual` flag.

### Mode: fuzzing

To run the fuzzing as described [here](doc/fuzzing.md), the following command can be used:

```
./advocate fuzzing [args]
```

To use the fuzzing, you need to apply a fuzzing mode with `-fuzzingMode [mode]`.
The available modes are:

- `GFuzz`: Run the [GFuzz](doc/fuzzing/GFuzz.md) based fuzzing
- `GFuzzHB`: Run the improved [GFuzz](doc/fuzzing/GFuzz.md#improvement-over-original-gfuzz) based fuzzing using happens-before information
- `Flow`: Run the [Flow](doc/fuzzing/Flow.md) based fuzzing
- `GFuzzHBFlow`: Run a combination of [GFuzzHB](doc/fuzzing/GFuzz.md) and the [Flow](doc/fuzzing/Flow.md) based fuzzing
- `GoPie`: Run the [GoPie](doc/fuzzing/GoPie.md#gopie) based fuzzing
- `GoCR`: Run an improved [GoPie](doc/fuzzing/GoPie.md#gopie-1) based fuzzing
- `GoPieHB`: Run an improved [GoPie](doc/fuzzing/GoPie.md#gopiehb) based fuzzing using happens-before information

All other required and additional args as well as the output files are the same as for the analysis mode.

For fuzzing, the `-prog [progName]` flag with the name of the program can be set.

The number of fuzzing runs per test/prog can be limited by setting `-maxFuzzingRun [maxRun]` (default: 100). To disable this, set `-maxFuzzingRun -1`
Alternatively, a maximum time can be set using `-timeoutFuz [to in s]` (default 7 min). To disable this, set `-timeoutFuz -1`

An example command would therefore be

```
./advocate fuzzing -path ~/pathToProg/progDir/ -fuzzingMode GoPieHB -prog progName
```

## Additional Tags

To set timeouts, you can set

- `-timeoutRec [to in s]`: Timeout for the recording in seconds (Default: 10 min)
- `-timeoutRep [to in s]`: Timeout for the replay (Default: 500 \* recording time)

To get additional information, the following tags can also be set:

- `-time`: measure the runtime for the different phases and create a time file
- `-stats`: create multiple statistic files as described [here](doc/statistics.md)
- `-notExec`: Find operations, that have never been executed

If one of these are set, the `-prog [name]` tag can be set to indicate the name of the program.

The created statistic and time files can also be found in the `advocateResult` folder.

In some situations, especially when only limited storage is available, it may
be useful to ignore atomic operations during recording and analysis. To do this,
you can set the `-ignoreAtomics`.

If the analysis of multiple tests was interrupted, running the toolchain
again would start from the beginning. If you want to skip all the already
finished tests, you can set `-cont`.

During the runtime, multiple status messages are shown in the terminal.
To disable them and only show found bugs, you can set `-noInfo`.

Sometimes the analysis or execution of a test may result in a panic in
advocatego. Since it would be annoying to terminate the total analysis of all tests,
only because the analysis of one of the tests crashed, a catch mechanism
has been implemented, that will only terminate the analysis of the given tests
and continue with the next, without crashing the whole program. To disable this,
you can set the `-panic` flag.

## Warning

It is the users responsibility of the user to make sure, that the input to
the program, including e.g. API calls are equal for the recording and the
tracing. Otherwise the replay is likely to get stuck.

Do not change the program code between trace recording and replay. The identification of the operations is based on the file names and lines, where the operations occur. If they get changed, the program will most likely block without terminating. If you need to change the program, you must either rerun the trace recording or change the effected trace elements in the recorded trace.
This also includes the adding of the replay header. Make sure, that it is already in the program (but commented out), when you run the recording.

## Settings

There are multiple constants that can be changed from the outside. This is
normally not necessary, but I you want to experiment with some of the settings,
you are free to do so.

Those constants can be set using the -settings [args] flag, where the args
must consists of the values that should be set in the form

```
-settings name1=value1,name2=value2
```

Make sure to not use spaces in this argument.

The following values can be changed:

| name           | default value | range                       | description                                                                                                            |
| -------------- | ------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| GFuzzW1        | 10            | $\mathbb{Q}$                | w1 weight for score in GFuzz as described [here](./fuzzing/GFuzz.md#determine-the-score)                               |
| GFuzzW2        | 10            | $\mathbb{Q}$                | w2 weight for score in GFuzz as described [here](./fuzzing/GFuzz.md#determine-the-score)                               |
| GFuzzW3        | 10            | $\mathbb{Q}$                | w3 weight for score in GFuzz as described [here](./fuzzing/GFuzz.md#determine-the-score)                               |
| GFuzzW4        | 10            | $\mathbb{Q}$                | w4 weight for score in GFuzz as described [here](./fuzzing/GFuzz.md#determine-the-score)                               |
| GFuzzFlipP     | 0.99          | $\mathbb{Q}, 0 <= val <= 1$ | probability of at least one of the selects to flip as described [here](./fuzzing/GFuzz.md#flip-probability)            |
| GFuzzFlipPMin  | 0.1           | $\mathbb{Q}, 0 <= val <= 1$ | minimum probability for each individual select to get flipped as described [here](./fuzzing/GFuzz.md#flip-probability) |
| GoPieW1        | 1             | $\mathbb{Q}$                | w1 weight for score in GoPie as described [here](./fuzzing/GoPie.md#mutation)                                          |
| GoPieW2        | 1             | $\mathbb{Q}$                | w2 weight for score in GoPie as described [here](./fuzzing/GoPie.md#mutation)                                          |
| GoPieBound     | 3             | $\mathbb{N}, val \geq 2$    | Maximum length of scheduling chain (BOUND) as described [here](./fuzzing/GoPie.md#mutation)                            |
| GoPieMutabound | 3             | $\mathbb{N}_{\neq 0}$       | Mutabound as described [here](./fuzzing/GoPie.md#mutation)                                                             |
| GoPieSCStart   | 5             | $\mathbb{N}_{\neq 0}$       | Number of starting point for scheduling chains as described [here](./fuzzing/GoPie.md#mutation)                        |
| SameElementTypeInSC | 0 (false) | $\{0,1\}$ | Only allow elements of the same type in a SC |
| WithoutReplay | 0 (false) | $\{0,1\}$ | Disable replay for goPie+ fuzzing runs |