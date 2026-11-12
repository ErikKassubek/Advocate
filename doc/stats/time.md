
# Time

If the `-time` flag is set, a two files with the runtimes will be created.

This will create a `times_detail_nameOfProg.csv` and a `times_total_nameOfProf.csv` file in the `goscResult` folder.
The first line of the files contains the column names. All times are given in seconds.

The details file has one line for each test run. For fuzzing it has one line
for each fuzzing run, meaning it can have multiple lines for each test. The columns are
as follows:

- `TestName`: Name of the test
- `Run`: Time to run the test without recording or replay
- `Recording`: Time for the recording og the trace
- `Io`: Total time for a read and write operations of the analyzer. This does not include the write of the trace by the recording.
- `Analysis`: Total time of the analysis
- `AnaHb`: Time of the analysis spend with calculating the HB clocks
- `AnaExitCode`:  Time of the analysis spend with analyzing the recording exit codes
- `AnaLeak`:  Time of the analysis spend with analyzing leaks
- `AnaClose`:  Time of the analysis spend with analyzing send on closed channel
- `AnaConcurrent`:  Time of the analysis spend with analyzing concurrent receive on the same channel
- `AnaResource`:  Time of the analysis spend with analyzing resource deadlocks (cyclic deadlocks)
- `AnaSelWithoutPartner`:  Time of the analysis spend with analyzing selects without partner
- `AnaUnlock`:  Time of the analysis spend with analyzing unlock before lock
- `AnaWait`: Time of the analysis spend with analyzing negative wait group counter
- `AnaMixed`: Time of the analysis spend with analyzing mixed deadlocks (should currently be zero)
- `FuzzingAna`: Time spend with gathering informations used to create new mutations
- `FuzzingMut`: Time spend with creating new mutations
- `Rewrite`: Time of the rewrite
- `Replay`: Total time for the replay
- `NumberReplay`: Number of performed replays

The total files contains the total time for each test and the total time for the program. It has the two columns

- `TestName`: Name of the test
- `Time`: Total time for the test

For fuzzing there is also only one line for each test, showing the total time for all fuzzing runs of this test.

The last time has the `TestName` `Total` and contains the total runtime for running all tests.
