# GoSC

GoSC is an artefact for the PLDI2027 submission for the paper "Stable Scheduling Constraints for Reliable Deadlock Discovery in Go".

![Arechitecture](./doc/img/architecture/architecture_new.png)

Gosc provides multiple different modes:

- fuzzing: Apply different fuzzing approaches to increase the reach of the analysis.
- experiments: Run specific experiments, especially the execution indexing implementation. 

For the stable scheduling constraint, the `fuzzing` mode is the mode to use.

The implementation also contains some additional modes like `record` or `replay`. Which are not directly required for the artefact.

## Usage

The following is a short overview on how to build and use the implementation. For more detailes see [./doc/usage.md](./doc/usage.md).

The implementation consists of a modified version of the go runtime in the [goPath](./goPatch/) directory and a controller program in [gosc](./gosc/).

### Build

The GoSC program is a Go program. To build it change into the corresponding directory and run 

```shell
go build
```


To build the modified go runtime, move into
[go-path/src](../goPatch/src/) directory and run the

```shell
./src/make.bash
```

or 

```shell
./src/make.bat
```

script. This will create a go executable in the `bin` directory.


On Linux you can use the make file by running `make` to automatically build both gosc and the modified runtime.

### Run

To run an analysis run the previously build `gosc` executable with

```
./gosc fuzzing -path ~/pathToProg/progDir/
```

As a default, this will iterate over all unit tests in the program. To execute the main function or one specific unit test add one of the following tags:

- `-main`
- `-exec [testName]`

To add a timeout add 

- `-timeoutRun [t]` for the timeout in seconds for one execution, or
- `-timeoutFuz [t]` for the timout in seconds used to analyze one test.

To select a closure mode, select it with `-closure [name]` with the following names:

- `hb`: happens-before based closure (default)
- `mhb`: must-happen-before based closure
- `pf`: routine-intern and fork closure
- `none`: no closure

Running a program will create an `goscResult` folder in the directory of the analyzed program, containing all output files. 