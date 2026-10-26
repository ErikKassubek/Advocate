
# Statistics

If the `-stats` flag is set when running the `analysis` or `fuzzing` mode, multiple statistic files will be created.

To create the statistic files, set the `-stats` flag.\
This will create three files

- `statsProgram_progName.csv`: Information about the whole program
- `statsTrace_progName.csv`: general statistics about the trace
- `statsAnalysis_progName.csv`: general statistics about the analysis results
- `statsAll_progName.csv`: detailed results about the trace and analysis
- `statsFuzzing_progName.csv`: Information about the fuzzing. Merged results for all runs of a test

## statsProgram

While for all the other stat files, the data is collected on a per test base, this file contains data about the whole analyzed program. This contains some information about the program, like the number of files and lines, as well as the number of unique found bugs
for the whole problem. Meaning if the same bug was found in multiple tests, it is just counted once here. For the replay we again
use the best possible result for the bug. The fields are

- `NoFiles`
- `NoLines`
- `NoNonEmptyLines`
- `NoTests`
- `NoRuns`
- `NoDetectedA01`
- `NoDetectedA02`
- `NoDetectedA03`s
- `NoDetectedA04`
- `NoDetectedA05`
- `NoDetectedA06`
- `NoDetectedA07`
- `NoDetectedA08`
- `NoDetectedP01`
- `NoDetectedP02`
- `NoDetectedP03`
- `NoDetectedP04`
- `NoDetectedP05`
- `NoDetectedL00`
- `NoDetectedL01`
- `NoDetectedL02`
- `NoDetectedL03`
- `NoDetectedL04`
- `NoDetectedL05`
- `NoDetectedL06`
- `NoDetectedR01`
- `NoDetectedR02`
- `NoReplayWrittenA01`
- `NoReplayWrittenA02`
- `NoReplayWrittenA03`
- `NoReplayWrittenA04`
- `NoReplayWrittenA05`
- `NoReplayWrittenA06`
- `NoReplayWrittenA07`
- `NoReplayWrittenA08`
- `NoReplayWrittenP01`
- `NoReplayWrittenP02`
- `NoReplayWrittenP03`
- `NoReplayWrittenP04`
- `NoReplayWrittenP05`
- `NoReplayWrittenL00`
- `NoReplayWrittenL01`
- `NoReplayWrittenL02`
- `NoReplayWrittenL03`
- `NoReplayWrittenL04`
- `NoReplayWrittenL05`
- `NoReplayWrittenL06`
- `NoReplayWrittenR01`
- `NoReplayWrittenR02`
- `NoReplaySuccessfulA01`
- `NoReplaySuccessfulA02`
- `NoReplaySuccessfulA03`
- `NoReplaySuccessfulA04`
- `NoReplaySuccessfulA05`
- `NoReplaySuccessfulA06`
- `NoReplaySuccessfulA07`
- `NoReplaySuccessfulA08`
- `NoReplaySuccessfulP01`
- `NoReplaySuccessfulP02`
- `NoReplaySuccessfulP03`
- `NoReplaySuccessfulP04`
- `NoReplaySuccessfulP05`
- `NoReplaySuccessfulL00`
- `NoReplaySuccessfulL01`
- `NoReplaySuccessfulL02`
- `NoReplaySuccessfulL03`
- `NoReplaySuccessfulL04`
- `NoReplaySuccessfulL05`
- `NoReplaySuccessfulL06`
- `NoReplaySuccessfulR01`
- `NoReplaySuccessfulR02`
- `NoUnexpectedPanicA01`
- `NoUnexpectedPanicA02`
- `NoUnexpectedPanicA03`
- `NoUnexpectedPanicA04`
- `NoUnexpectedPanicA05`
- `NoUnexpectedPanicA06`
- `NoUnexpectedPanicA07`
- `NoUnexpectedPanicA08`
- `NoUnexpectedPanicP01`
- `NoUnexpectedPanicP02`
- `NoUnexpectedPanicP03`
- `NoUnexpectedPanicP04`
- `NoUnexpectedPanicP05`
- `NoUnexpectedPanicL00`
- `NoUnexpectedPanicL01`
- `NoUnexpectedPanicL02`
- `NoUnexpectedPanicL03`
- `NoUnexpectedPanicL04`
- `NoUnexpectedPanicL05`
- `NoUnexpectedPanicL06`
- `NoUnexpectedPanicR01`
- `NoUnexpectedPanicR02`

## statsTrace

This file contains statistics about the program traces. It contains one
line with the column names and then one line for each tests, or just one line if it is run on a main file.\
The columns are

- `TestName`: Name of the test`, "Main" if main function
- `NoEvents`: Total number of all events in the trace`
- `NoGoroutines`: Number of go routines
- `NoAtomicEvents`: Total number of atomic events`
- `NoChannelEvents`: Total number of channel events`
- `NoSelectEvents`: Total number of select events`
- `NoMutexEvents`: Total number of mutex events
- `NoWaitgroupEvents`: total number of wait group events`
- `NoCondVariablesEvents`: Total number of events on conditional vars
- `NoOnceOperations`: Total number of once operations

## statsAnalysis

This file contains general statistics about the program analysis. It contains one
line with the column names and then one line for each test, or just one line if it is run on a main file.
For fuzzing, it will create one line per fuzzing run.
The data columns can be split into to groups. The first part contains the total number of relevant events in the analysis.
In e.g. loops. the same bug may be detected multiple times. For this reason we also show the number of
unique found bugs per test. For the rewrite and replay we always count the best possible value, meaning if
a bug was found twice and once the replay worked, but the other time it did not, we still count it as a successful replay.\
The columns are

- `TestName`: Name of the test, "Main" for main func
- `NumberActualBugTotal`: Total number of actually occurring bugs
- `NoLeaksTotal`: Total number of detected leaks
- `NoLeaksWithRewriteTotal`: Total number of leaks where a rewrite was possible
- `NoLeaksResolvedViaReplayTotal`: Total number of leaks that could be resolved with replay
- `NoPanicsTotal`: Total number of possible panic bugs
- `NoPanicsVerifiedViaReplayTotal`: Total number of possible panic bugs confirmed with replay
- `NoUnexpectedPanicsInReplayTotal`: Number of unexpected bugs in replay
- `NoProbInRecordingTotal`: Number of panics or timeouts in recording
- `NumberActualBugUnique`: Total number of actually occurring bugs
- `NoLeaksUnique`: Total number of detected leaks
- `NoLeaksWithRewriteUnique`: Total number of leaks where a rewrite was possible
- `NoLeaksResolvedViaReplayUnique`: Total number of leaks that could be resolved with replay
- `NoPanicsUnique`: Total number of possible panic bugs
- `NoPanicsVerifiedViaReplayUnique`: Total number of possible panic bugs confirmed with replay
- `NoUnexpectedPanicsInReplayUnique`: Number of unexpected bugs in replay


## statsAll

This file contains full statistics about the program trace and analysis. It contains one
line with the column names and then one line for each tests, or just one line if it is run on a main file.
For fuzzing, it will create one line per fuzzing run.\

The columns can be categories into the following sets:

- TestName: Name of the test, "Main" for main function
- Trace info
  - Total number of events
  - Number of routines (empty/not empty)
  - Number of elements per type (e.g. channel, mutex)
  - Number of operations per type
- Analysis info
  - Both for total and unique
    - number detected for each bug type
    - number replay created for each bug type
    - number replay successful for each bug type
    - number of unexpected panics in replay for each bug type


The full list of columns is as follows:

- `TestName`
- `NoEvents`
- `NoGoroutines`
- `NoNotEmptyGoroutines`
- `NoSpawnEvents`
- `NoRoutineEndEvents`
- `NoAtomics`
- `NoAtomicEvents`
- `NoChannels`
- `NoBufferedChannels`
- `NoUnbufferedChannels`
- `NoChannelEvents`
- `NoBufferedChannelEvents`
- `NoUnbufferedChannelEvents`
- `NoSelectEvents`
- `NoSelectCases`
- `NoSelectNonDefaultEvents`
- `NoSelectDefaultEvents`
- `NoMutex`
- `NoMutexEvents`
- `NoWaitgroup`
- `NoWaitgroupEvent`
- `NoCondVariables`
- `NoCondVariablesEvents`
- `NoOnce`
- `NoOnceOperations`
- `NoTotalDetectedA01`
- `NoTotalDetectedA02`
- `NoTotalDetectedA03`
- `NoTotalDetectedA04`
- `NoTotalDetectedA05`
- `NoTotalDetectedA06`
- `NoTotalDetectedA07`
- `NoTotalDetectedA08`
- `NoTotalDetectedP01`
- `NoTotalDetectedP02`
- `NoTotalDetectedP03`
- `NoTotalDetectedP04`
- `NoTotalDetectedP05`
- `NoTotalDetectedL00`
- `NoTotalDetectedL01`
- `NoTotalDetectedL02`
- `NoTotalDetectedL03`
- `NoTotalDetectedL04`
- `NoTotalDetectedL05`
- `NoTotalDetectedL06`
- `NoTotalDetectedL07`
- `NoTotalDetectedL08`
- `NoTotalDetectedL09`
- `NoTotalDetectedL10`
- `NoTotalDetectedR01`
- `NoTotalDetectedR02`
- `NoUniqueDetectedA01`
- `NoUniqueDetectedA02`
- `NoUniqueDetectedA03`
- `NoUniqueDetectedA04`
- `NoUniqueDetectedA05`
- `NoUniqueDetectedA06`
- `NoUniqueDetectedA07`
- `NoUniqueDetectedA08`
- `NoUniqueDetectedP01`
- `NoUniqueDetectedP02`
- `NoUniqueDetectedP03`
- `NoUniqueDetectedP04`
- `NoUniqueDetectedP05`
- `NoUniqueDetectedL00`
- `NoUniqueDetectedL01`
- `NoUniqueDetectedL02`
- `NoUniqueDetectedL03`
- `NoUniqueDetectedL04`
- `NoUniqueDetectedL05`
- `NoUniqueDetectedL06`
- `NoUniqueDetectedL07`
- `NoUniqueDetectedL08`
- `NoUniqueDetectedL09`
- `NoUniqueDetectedL10`
- `NoUniqueDetectedR01`
- `NoUniqueDetectedR02`
- `NoTotalReplayWrittenA01`
- `NoTotalReplayWrittenA02`
- `NoTotalReplayWrittenA03`
- `NoTotalReplayWrittenA04`
- `NoTotalReplayWrittenA05`
- `NoTotalReplayWrittenA06`
- `NoTotalReplayWrittenA07`
- `NoTotalReplayWrittenA08`
- `NoTotalReplayWrittenP01`
- `NoTotalReplayWrittenP02`
- `NoTotalReplayWrittenP03`
- `NoTotalReplayWrittenP04`
- `NoTotalReplayWrittenP05`
- `NoTotalReplayWrittenL00`
- `NoTotalReplayWrittenL01`
- `NoTotalReplayWrittenL02`
- `NoTotalReplayWrittenL03`
- `NoTotalReplayWrittenL04`
- `NoTotalReplayWrittenL05`
- `NoTotalReplayWrittenL06`
- `NoTotalReplayWrittenL07`
- `NoTotalReplayWrittenL08`
- `NoTotalReplayWrittenL09`
- `NoTotalReplayWrittenL10`
- `NoTotalReplayWrittenR01`
- `NoTotalReplayWrittenR02`
- `NoUniqueReplayWrittenA01`
- `NoUniqueReplayWrittenA02`
- `NoUniqueReplayWrittenA03`
- `NoUniqueReplayWrittenA04`
- `NoUniqueReplayWrittenA05`
- `NoUniqueReplayWrittenA06`
- `NoUniqueReplayWrittenA07`
- `NoUniqueReplayWrittenA08`
- `NoUniqueReplayWrittenP01`
- `NoUniqueReplayWrittenP02`
- `NoUniqueReplayWrittenP03`
- `NoUniqueReplayWrittenP04`
- `NoUniqueReplayWrittenP05`
- `NoUniqueReplayWrittenL00`
- `NoUniqueReplayWrittenL01`
- `NoUniqueReplayWrittenL02`
- `NoUniqueReplayWrittenL03`
- `NoUniqueReplayWrittenL04`
- `NoUniqueReplayWrittenL05`
- `NoUniqueReplayWrittenL06`
- `NoUniqueReplayWrittenL07`
- `NoUniqueReplayWrittenL08`
- `NoUniqueReplayWrittenL09`
- `NoUniqueReplayWrittenL10`
- `NoUniqueReplayWrittenR01`
- `NoUniqueReplayWrittenR02`
- `NoTotalReplaySuccessfulA01`
- `NoTotalReplaySuccessfulA02`
- `NoTotalReplaySuccessfulA03`
- `NoTotalReplaySuccessfulA04`
- `NoTotalReplaySuccessfulA05`
- `NoTotalReplaySuccessfulA06`
- `NoTotalReplaySuccessfulA07`
- `NoTotalReplaySuccessfulA08`
- `NoTotalReplaySuccessfulP01`
- `NoTotalReplaySuccessfulP02`
- `NoTotalReplaySuccessfulP03`
- `NoTotalReplaySuccessfulP04`
- `NoTotalReplaySuccessfulP05`
- `NoTotalReplaySuccessfulL00`
- `NoTotalReplaySuccessfulL01`
- `NoTotalReplaySuccessfulL02`
- `NoTotalReplaySuccessfulL03`
- `NoTotalReplaySuccessfulL04`
- `NoTotalReplaySuccessfulL05`
- `NoTotalReplaySuccessfulL06`
- `NoTotalReplaySuccessfulL07`
- `NoTotalReplaySuccessfulL08`
- `NoTotalReplaySuccessfulL09`
- `NoTotalReplaySuccessfulL10`
- `NoTotalReplaySuccessfulR01`
- `NoTotalReplaySuccessfulR02`
- `NoUniqueReplaySuccessfulA01`
- `NoUniqueReplaySuccessfulA02`
- `NoUniqueReplaySuccessfulA03`
- `NoUniqueReplaySuccessfulA04`
- `NoUniqueReplaySuccessfulA05`
- `NoUniqueReplaySuccessfulA06`
- `NoUniqueReplaySuccessfulA07`
- `NoUniqueReplaySuccessfulA08`
- `NoUniqueReplaySuccessfulP01`
- `NoUniqueReplaySuccessfulP02`
- `NoUniqueReplaySuccessfulP03`
- `NoUniqueReplaySuccessfulP04`
- `NoUniqueReplaySuccessfulP05`
- `NoUniqueReplaySuccessfulL00`
- `NoUniqueReplaySuccessfulL01`
- `NoUniqueReplaySuccessfulL02`
- `NoUniqueReplaySuccessfulL03`
- `NoUniqueReplaySuccessfulL04`
- `NoUniqueReplaySuccessfulL05`
- `NoUniqueReplaySuccessfulL06`
- `NoUniqueReplaySuccessfulL07`
- `NoUniqueReplaySuccessfulL08`
- `NoUniqueReplaySuccessfulL09`
- `NoUniqueReplaySuccessfulL10`
- `NoUniqueReplaySuccessfulR01`
- `NoUniqueReplaySuccessfulR02`
- `NoTotalUnexpectedPanicA01`
- `NoTotalUnexpectedPanicA02`
- `NoTotalUnexpectedPanicA03`
- `NoTotalUnexpectedPanicA04`
- `NoTotalUnexpectedPanicA05`
- `NoTotalUnexpectedPanicA06`
- `NoTotalUnexpectedPanicA07`
- `NoTotalUnexpectedPanicA08`
- `NoTotalUnexpectedPanicP01`
- `NoTotalUnexpectedPanicP02`
- `NoTotalUnexpectedPanicP03`
- `NoTotalUnexpectedPanicP04`
- `NoTotalUnexpectedPanicP05`
- `NoTotalUnexpectedPanicL00`
- `NoTotalUnexpectedPanicL01`
- `NoTotalUnexpectedPanicL02`
- `NoTotalUnexpectedPanicL03`
- `NoTotalUnexpectedPanicL04`
- `NoTotalUnexpectedPanicL05`
- `NoTotalUnexpectedPanicL06`
- `NoTotalUnexpectedPanicL07`
- `NoTotalUnexpectedPanicL08`
- `NoTotalUnexpectedPanicL09`
- `NoTotalUnexpectedPanicL10`
- `NoTotalUnexpectedPanicR01`
- `NoTotalUnexpectedPanicR02`
- `NoUniqueUnexpectedPanicA01`
- `NoUniqueUnexpectedPanicA02`
- `NoUniqueUnexpectedPanicA03`
- `NoUniqueUnexpectedPanicA04`
- `NoUniqueUnexpectedPanicA05`
- `NoUniqueUnexpectedPanicA06`
- `NoUniqueUnexpectedPanicA07`
- `NoUniqueUnexpectedPanicA08`
- `NoUniqueUnexpectedPanicP01`
- `NoUniqueUnexpectedPanicP02`
- `NoUniqueUnexpectedPanicP03`
- `NoUniqueUnexpectedPanicP04`
- `NoUniqueUnexpectedPanicP05`
- `NoUniqueUnexpectedPanicL00`
- `NoUniqueUnexpectedPanicL01`
- `NoUniqueUnexpectedPanicL02`
- `NoUniqueUnexpectedPanicL03`
- `NoUniqueUnexpectedPanicL04`
- `NoUniqueUnexpectedPanicL05`
- `NoUniqueUnexpectedPanicL06`
- `NoUniqueUnexpectedPanicL07`
- `NoUniqueUnexpectedPanicL08`
- `NoUniqueUnexpectedPanicL09`
- `NoUniqueUnexpectedPanicL10`
- `NoUniqueUnexpectedPanicR01`
- `NoUniqueUnexpectedPanicR02`

## statsFuzzing

This file contains information of the fuzzing for each test. This file is only created if the analysis is run in fuzzing mode.
The info for each test contains the name, the number of runs in fuzzing and the
number of unique bugs found over all fuzzing runs.\

The columns are

- `TestName`
- `NoRuns`
- `NoDetectedA01`
- `NoDetectedA02`
- `NoDetectedA03`
- `NoDetectedA04`
- `NoDetectedA05`
- `NoDetectedA06`
- `NoDetectedA07`
- `NoDetectedA08`
- `NoDetectedP01`
- `NoDetectedP02`
- `NoDetectedP03`
- `NoDetectedP04`
- `NoDetectedP05`
- `NoDetectedL00`
- `NoDetectedL01`
- `NoDetectedL02`
- `NoDetectedL03`
- `NoDetectedL04`
- `NoDetectedL05`
- `NoDetectedL06`
- `NoDetectedL07`
- `NoDetectedL08`
- `NoDetectedL09`
- `NoDetectedL10`
- `NoDetectedR01`
- `NoDetectedR02`
- `NoReplayWrittenA01`
- `NoReplayWrittenA02`
- `NoReplayWrittenA03`
- `NoReplayWrittenA04`
- `NoReplayWrittenA05`
- `NoReplayWrittenA06`
- `NoReplayWrittenA07`
- `NoReplayWrittenA08`
- `NoReplayWrittenP01`
- `NoReplayWrittenP02`
- `NoReplayWrittenP03`
- `NoReplayWrittenP04`
- `NoReplayWrittenP05`
- `NoReplayWrittenL00`
- `NoReplayWrittenL01`
- `NoReplayWrittenL02`
- `NoReplayWrittenL03`
- `NoReplayWrittenL04`
- `NoReplayWrittenL05`
- `NoReplayWrittenL06`
- `NoReplayWrittenL07`
- `NoReplayWrittenL08`
- `NoReplayWrittenL09`
- `NoReplayWrittenL10`
- `NoReplayWrittenR01`
- `NoReplayWrittenR02`
- `NoReplaySuccessfulA01`
- `NoReplaySuccessfulA02`
- `NoReplaySuccessfulA03`
- `NoReplaySuccessfulA04`
- `NoReplaySuccessfulA05`
- `NoReplaySuccessfulA06`
- `NoReplaySuccessfulA07`
- `NoReplaySuccessfulA08`
- `NoReplaySuccessfulP01`
- `NoReplaySuccessfulP02`
- `NoReplaySuccessfulP03`
- `NoReplaySuccessfulP04`
- `NoReplaySuccessfulP05`
- `NoReplaySuccessfulL00`
- `NoReplaySuccessfulL01`
- `NoReplaySuccessfulL02`
- `NoReplaySuccessfulL03`
- `NoReplaySuccessfulL04`
- `NoReplaySuccessfulL05`
- `NoReplaySuccessfulL06`
- `NoReplaySuccessfulL07`
- `NoReplaySuccessfulL08`
- `NoReplaySuccessfulL09`
- `NoReplaySuccessfulL10`
- `NoReplaySuccessfulR01`
- `NoReplaySuccessfulR02`
- `NoUnexpectedPanicA01`
- `NoUnexpectedPanicA02`
- `NoUnexpectedPanicA03`
- `NoUnexpectedPanicA04`
- `NoUnexpectedPanicA05`
- `NoUnexpectedPanicA06`
- `NoUnexpectedPanicA07`
- `NoUnexpectedPanicA08`
- `NoUnexpectedPanicP01`
- `NoUnexpectedPanicP02`
- `NoUnexpectedPanicP03`
- `NoUnexpectedPanicP04`
- `NoUnexpectedPanicP05`
- `NoUnexpectedPanicL00`
- `NoUnexpectedPanicL01`
- `NoUnexpectedPanicL02`
- `NoUnexpectedPanicL03`
- `NoUnexpectedPanicL04`
- `NoUnexpectedPanicL05`
- `NoUnexpectedPanicL06`
- `NoUnexpectedPanicL07`
- `NoUnexpectedPanicL08`
- `NoUnexpectedPanicL09`
- `NoUnexpectedPanicL10`
- `NoUnexpectedPanicR01`
- `NoUnexpectedPanicR02`


## Error Codes

- A01: Actual Send on Closed Channel
- A02: Actual Receive on Closed Channel
- A03: Actual Close on Closed Channel
- A04: Actual Close on nil channel
- A05: Actual negative Wait Group
- A06: Actual unlock of not locked mutex
- A07: Actual Leak (Non-Cyclic blocking bug)
- A08: Actual Deadlock (Cyclic blocking bug)
- A09: Actual Mixed Deadlock
- P01: Possible send on closed channel
- P02: Possible receive on closed channel
- P03: Possible negative waitgroup counter
- P04: Possible unlock of not locked mutex
- L00: Leak on routine without blocking element
- L01: Leak on unbuffered channel with possible partner
- L02: Leak on unbuffered channel without possible partner
- L03: Leak on buffered channel with possible partner
- L04: Leak on buffered channel without possible partner
- L05: Leak on nil channel
- L06: Leak on select with possible partner
- L07: Leak on select without possible partner
- L08: Leak on mutex
- L09: Leak on waitgroup
- L10: Leak on cond
- R01: Unknown panic in recording
- R02: Timeout in recording
