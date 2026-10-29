// Copyright (c) 2024 Erik Kassubek
//
// File: fuzzing.go
// Brief: Main file for fuzzing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_fuzzing

import (
	"advocate/advoc/toolchain"
	"advocate/analysis/a_base"
	"advocate/fuzzing/f_base"
	"advocate/fuzzing/f_constraints"
	"advocate/fuzzing/modes/f_gfuzz"
	"advocate/fuzzing/modes/f_gopie"
	f_guided "advocate/fuzzing/modes/g_guided"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/results"
	"advocate/utils/results/stats"
	"advocate/utils/timer"
	"advocate/utils/types"
	"fmt"
	"path/filepath"
	"time"
)

// Fuzzing creates the fuzzing data and runs the fuzzing executions
func Fuzzing() error {
	modes := []string{f_base.GoPie, f_base.GoCR, f_base.GoCRHB, f_base.GFuzz, f_base.GFuzzHBFlow, f_base.GFuzzHB, f_base.Guided}
	if !types.Contains(modes, flags.FuzzingMode) {
		return fmt.Errorf("Invalid fuzzing mode '%s'. Possible values are Guided, GoPie, GoCR, GoCRHB, GFuzz, GFuzzFlow, GFuzzHB, Flow", flags.FuzzingMode)
	}

	f_base.MaxNumberRuns = flags.MaxFuzzingRun
	if flags.TimeoutFuzzing > 0 {
		f_base.MaxTime = time.Duration(flags.TimeoutFuzzing) * time.Second
		f_base.MaxTimeSet = true
	}

	f_base.FuzzingModeGoPie = (flags.FuzzingMode == f_base.GoPie || flags.FuzzingMode == f_base.GoCR || flags.FuzzingMode == f_base.GoCRHB)
	f_base.FuzzingModeGoCRHBPlus = (flags.FuzzingMode == f_base.GoCR || flags.FuzzingMode == f_base.GoCRHB)
	f_base.FuzzingModeGFuzz = (flags.FuzzingMode == f_base.GFuzz || flags.FuzzingMode == f_base.GFuzzHBFlow || flags.FuzzingMode == f_base.GFuzzHB)
	f_base.FuzzingModeGuided = (flags.FuzzingMode == f_base.Guided || flags.FuzzingMode == f_base.Default)
	f_base.UseHBInfoFuzzing = (flags.FuzzingMode == f_base.Guided || flags.FuzzingMode == f_base.GFuzzHB || flags.FuzzingMode == f_base.GFuzzHBFlow || flags.FuzzingMode == f_base.GoCR || flags.FuzzingMode == f_base.GoCRHB)

	if flags.Continue {
		log.Info("Continue fuzzing")
	} else {
		log.Infof("Start fuzzing in mode %s", flags.FuzzingMode)
	}

	// run either fuzzing on main or fuzzing on one test
	if flags.ModeMain || flags.ExecName != "" {
		if flags.ModeMain {
			log.Info("Run fuzzing on main function")
		} else {
			log.Info("Run fuzzing on test ", flags.ExecName)
		}

		err := runFuzzing("", true, 0, 0)

		if flags.CreateStatistics {
			err := stats.CreateStatsFuzzing(f_base.GetPath(flags.ProgPath))
			if err != nil {
				log.Error("Failed to create fuzzing stats: ", err.Error())
			}
			err = stats.CreateStatsTotal(f_base.GetPath(flags.ProgPath))
			if err != nil {
				log.Error("Failed to create total stats: ", err.Error())
			}
		}

		clearDataFull()
		timer.ResetFuzzing()

		if flags.DeleteTraces {
			toolchain.RemoveTraces(flags.ProgPath)
		}

		return err
	}

	log.Info("Run fuzzing on all tests")

	// run fuzzing on all tests
	testFiles, maxFileNumber, totalFiles, err := toolchain.FindTestFiles(flags.ProgPath, flags.Continue)
	if err != nil {
		return fmt.Errorf("Failed to find test files: %v", err)
	}

	log.Infof("Found %d test files", totalFiles)

	// Process each test file
	fileCounter := 0
	if flags.Continue {
		fileCounter = maxFileNumber
	}

	for i, testFile := range testFiles {
		fileCounter++
		// log.Progressf("Progress %s: %d/%d", flags.ProgName, fileCounter, totalFiles)
		log.Infof("Processing file: %s\n", testFile)

		testFunctions, err := toolchain.FindTestFunctions(testFile)
		if err != nil || len(testFunctions) == 0 {
			log.Info("Could not find test functions in ", testFile)
			continue
		}

		for j, testFunc := range testFunctions {
			flags.ExecName = testFunc

			for control.WasCanceledRAM() {
				log.Error("Wait RAM")
				time.Sleep(6 * time.Second)
			}

			a_base.Clear(true)
			ResetFuzzing()
			timer.ResetTest()

			timer.Start(timer.TotalTest)

			log.Progressf("Run fuzzing for %s (%d/%d) -> %s (%d/%d)", testFile, fileCounter, totalFiles, testFunc, j+1, len(testFunctions))

			firstRun := (i == 0 && j == 0)

			err := runFuzzing(testFile, firstRun, fileCounter, j+1)
			if err != nil {
				log.Error("Error in fuzzing: ", err.Error())
				clearDataRun()
			}

			timer.Stop(timer.TotalTest)

			timer.UpdateTimeFileOverview(testFunc)
		}
	}

	if flags.CreateStatistics {
		err := stats.CreateStatsFuzzing(f_base.GetPath(flags.ProgPath))
		if err != nil {
			log.Error("Failed to create fuzzing stats: ", err.Error())
		}
		err = stats.CreateStatsTotal(f_base.GetPath(flags.ProgPath))
		if err != nil {
			log.Error("Failed to create total stats: ", err.Error())
		}
	}

	if flags.DeleteTraces {
		toolchain.RemoveTraces(flags.ProgPath)
	}

	return nil
}

// Run Fuzzing on one program/test
//
// Parameter:
//   - testPath string: path to the test file
//   - hBInfoFuzzing bool: whether to us HB info in fuzzing
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
func runFuzzing(testPath string, firstRun bool, fileNumber, testNumber int) error {
	clearDataFull()

	// while there are available mutations, run them
	startTime := time.Now()
	for f_base.NumberFuzzingRuns == 0 || f_base.MutationQueue.Size() != 0 {

		// clean up
		clearDataRun()
		timer.ResetFuzzing()

		if flags.CancelTestIfBugFound && results.GetBugWasFound() {
			log.Infof("Cancel test after %d runs", f_base.NumberFuzzingRuns)
			break
		}

		log.Info("Fuzzing Run: ", f_base.NumberFuzzingRuns+1)

		fuzzingPath := ""
		progPathDir := paths.GetDirectory(flags.ProgPath)
		var order f_base.Mutation
		if f_base.NumberFuzzingRuns != 0 {
			order = popMutation()
			if order.MutType == f_base.MutPiType {
				fuzzingPath = filepath.Join(progPathDir,
					filepath.Join("fuzzingTraces",
						fmt.Sprintf("fuzzingTrace_%d", order.MutPie)))
			} else {
				err := f_base.WriteMutationToFile(progPathDir, order)
				if err != nil {
					return err
				}
			}
		}

		firstRun = firstRun && (f_base.NumberFuzzingRuns == 0)

		// Run the test/mutation
		runAnalysis := true
		runRecord := true
		traceID, numberResults, err := toolchain.Run(testPath, runRecord, runAnalysis, runAnalysis,
			f_base.NumberFuzzingRuns, fuzzingPath, firstRun, fileNumber, testNumber)

		f_base.NumberFuzzingRuns++

		if numberResults > flags.MaxNumberElements {
			continue
		}

		if err != nil {
			log.Error("Fuzzing run failed: ", err.Error())
		} else {
			log.Info("Parse recorded trace for fuzzing information")

			// collect the required data to decide whether run is interesting
			// and to create the mutations
			ParseTrace(&a_base.MainTrace)

			if control.WasCanceled() {
				log.Error("Fuzzing run was canceled due to memory")
				f_gopie.ClearDataRun()
				a_base.ClearTrace()
				a_base.ClearData()
				continue
			}

			log.Infof("Create mutations")

			// add mutation based on guided fuzzing
			if f_base.FuzzingModeGuided {
				f_guided.CreateMutations()
			}
			// Add mutation based on GFuzz
			if f_base.FuzzingModeGFuzz {
				f_gfuzz.CreateMutations(false)
			}
			// add mutations based on GoPie
			if f_base.FuzzingModeGoPie {
				f_gopie.CreateMutations(order.MutPie)
			}

			if flags.CreateStatistics {
				stats.CreateStats(flags.ExecName, traceID, f_base.NumberFuzzingRuns-1)
			}

			log.Infof("Current fuzzing queue size: %d", f_base.MutationQueue.Size())

			if f_base.FuzzingModeGFuzz {
				f_gfuzz.MergeTraceInfoIntoFileInfo()
			}
		}

		// cancel if max number of mutations have been reached
		if f_base.MaxNumberRuns != -1 && f_base.NumberFuzzingRuns >= f_base.MaxNumberRuns {
			log.Infof("Finish fuzzing because maximum number of mutation runs (%d) have been reached", f_base.MaxNumberRuns)
			return nil
		}

		// cancel if max fuzzing time has been reached
		if f_base.MaxTimeSet {
			since := time.Since(startTime)
			if since > f_base.MaxTime {
				log.Infof("Finish fuzzing because maximum runtime for fuzzing (%d min) has been reached", int(f_base.MaxTime.Minutes()))
				return nil
			} else {
				remaining := f_base.MaxTime - time.Since(startTime)
				log.Infof("Remaining fuzzing time: %d:%d min", int(remaining.Minutes()), int(remaining.Seconds())%60)
			}
		}

		// cancel if bug was found
		if f_base.FinishIfBugFound && numberResults > 0 {
			return nil
		}

		a_base.ClearTrace()
		a_base.ClearData()

	}

	if f_base.FuzzingModeGoPie {
		toolchain.ClearFuzzingTrace()
	}

	log.Infof("Finish fuzzing after %d runs\n", f_base.NumberFuzzingRuns)

	return nil
}

// Remove and return the first mutation from the mutation queue
//
// Returns:
//   - the first mutation from the mutation queue
func popMutation() f_base.Mutation {
	return f_base.MutationQueue.Pop()
}

// Reset fuzzing
func ResetFuzzing() {
	f_base.NumberFuzzingRuns = 0
	f_base.MutationQueue = types.NewQueue[f_base.Mutation]()
	// count how often a specific mutation has been in the queue
	f_base.AllMutations = make(map[string]int)
	f_base.ChainFiles = make(map[int]f_constraints.Constraint)

}

func clearDataFull() {
	f_base.ClearDataFull()
	f_gopie.ClearData()
	f_gfuzz.ClearDataFull()
}

func clearDataRun() {
	f_gopie.ClearDataRun()
	f_gfuzz.ClearDataRun()
}
