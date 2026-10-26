// Copyright (c) 2025 Erik Kassubek
//
// File: analyzer.go
// Brief: Starting point for the analyzer and replay
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/analysis/a_analysis"
	"advocate/analysis/a_base"
	"advocate/fuzzing/f_base"
	"advocate/fuzzing/f_roc"
	"advocate/fuzzing/modes/f_active"
	"advocate/utils/consts"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"advocate/utils/io"
	"advocate/utils/log"
	"advocate/utils/results/results"
	"advocate/utils/timer"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// runAnalyzer is the starting point to the analyzer.
// This function will read the trace at a stored path, analyze it and,
// if needed, rewrite the trace.
//
// Parameter:
//   - pathTrace string: path to the trace to be analyzed
//   - outReadable string: path to the readable result file
//   - outMachine string: path to the machine result file
//   - fifo bool: assume, that the channels work as a fifo queue
//   - newTrace string: path to where the rewritten trace should be created, set to "" to disable rewrite
//   - fuzzingRun int: number of fuzzing run (0 for recording, then always add 1)
//   - testFile string: name of the analzed test file
//   - testName string: name of the analzed test, "main" if main
//   - replay bool: true if called on relpay
//
// Returns:
//   - error
func runAnalyzer(pathTrace string,
	outReadable string, outMachine string,
	newTrace string, fuzzingRun int, testFile, testName string, replay bool) error {

	log.Info("Run Analyzer")

	if pathTrace == "" {
		return fmt.Errorf("Please provide a path to the trace files. Set with -trace [folder]")
	}

	// run the analysis and, if requested, create a reordered trace file
	// based on the analysis results

	if !replay {
		results.InitResults(outReadable, outMachine)
	}

	_, numberOfRoutines, numberElems, err := io.CreateTraceFromFiles(pathTrace, true)

	if err != nil && fuzzingRun <= 0 {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			if testFile != "" {
				return fmt.Errorf("Test run failed and no trace was recorded for: %s -> %s", testFile, testName)
			} else {
				return fmt.Errorf("Program run failed and no trace was recorded")
			}
		}
		return err
	}

	if numberElems == 0 {
		log.Errorf("Trace at %s does not contain any elements", pathTrace)
	}

	log.Infof("Read trace with %d elements in %d routines", numberElems, numberOfRoutines)

	a_analysis.RunAnalysis(fuzzingRun >= 0)

	if control.WasCanceled() {
		// analysis.LogSizes()
		if control.WasCanceledRAM() {
			a_base.Clear()
			return fmt.Errorf("Analysis was canceled due to insufficient RAM")
		}
		a_base.Clear()
		return fmt.Errorf("Analysis was canceled due to unexpected panic")
	}
	log.Info("Analysis finished")

	numberOfResults, err := results.CreateResultFiles(true)
	if err != nil {
		log.Error("Error in printing summary: ", err.Error())
	}

	if flags.OnlyAPanicAndLeak {
		return nil
	}

	if flags.NoRewrite {
		log.Info("Skip rewrite")
		return nil
	}

	numberRewrittenTrace := 0
	failedRewrites := 0
	notNeededRewrites := 0

	if err != nil {
		log.Error("Failed to create result files: ", err)
		return nil
	}

	rewrittenBugs := make(map[helper.ResultType][]string) // bug type -> paths string

	file := filepath.Base(pathTrace)
	rewriteNr := "0"
	spl := strings.Split(file, "_")
	if len(spl) > 1 {
		rewriteNr = spl[len(spl)-1]
	}

	if newTrace != "" {
		if numberOfResults != 0 {
			log.Info("Start rewrite")
		}

		for resultIndex := 0; resultIndex < numberOfResults; resultIndex++ {
			needed, err := rewriteTrace(outMachine,
				newTrace+"_"+strconv.Itoa(resultIndex+1)+consts.Sep, resultIndex, &rewrittenBugs)

			if !needed {
				notNeededRewrites++
				fmt.Printf("Bugreport info: %s_%d,fail\n", rewriteNr, resultIndex+1)
			} else if err != nil {
				failedRewrites++
				fmt.Printf("Bugreport info: %s_%d,fail\n", rewriteNr, resultIndex+1)
			} else { // needed && err == nil
				numberRewrittenTrace++
				fmt.Printf("Bugreport info: %s_%d,suc\n", rewriteNr, resultIndex+1)
			}

			if control.WasCanceled() {
				failedRewrites += max(0, numberOfResults-resultIndex-1)
				break
			}
		}
		if control.WasCanceledRAM() {
			log.Error("Rewrite Canceled: Not enough RAM")
		} else {
			log.Info("Finished Rewrite")
		}
		log.Info("Number Results: ", numberOfResults)
		log.Info("Successfully rewrites: ", numberRewrittenTrace)
		log.Info("No need/not possible to rewrite: ", notNeededRewrites)
		if failedRewrites > 0 {
			log.Info("Failed rewrites: ", failedRewrites)
		} else {
			log.Info("Failed rewrites: ", failedRewrites)
		}
	}

	return nil
}

// Rewrite the trace file based on given analysis results
//
// Parameter:
//   - outMachine string: The path to the analysis result file
//   - newTrace string: The path where the new traces folder will be created
//   - resultIndex int: The index of the result to use for the reordered trace file
//   - rewrittenTrace *map[helper.ResultType][]string: set of bugs that have been already rewritten
//
// Returns:
//   - bool: true, if a rewrite was necessary, false if not (e.g. actual bug, warning)
//   - error: An error if the trace file could not be created
func rewriteTrace(outMachine string, newTrace string, resultIndex int,
	rewrittenTrace *map[helper.ResultType][]string) (bool, error) {
	timer.Start(timer.Rewrite)
	defer timer.Stop(timer.Rewrite)

	actual, bug, err := io.ReadAnalysisResults(outMachine, resultIndex)
	if err != nil {
		return false, err
	}

	if actual {
		return false, nil
	}

	// the same bug was found and confirmed by replay in an earlier run,
	// either in fuzzing or in another test
	// It is therefore not needed to rewrite it again
	if !flags.NoSkipRewrite && results.WasAlreadyConfirmed(bug.GetBugString()) {
		return false, nil
	}

	constr, code, err := f_active.RewriteTrace(bug, *rewrittenTrace)

	if err != nil {
		return constr != nil, err
	}

	closure := f_roc.CalcuateStableClosure(constr)

	err = f_base.WriteConstraint(newTrace, &a_base.MainTrace, constr, closure)
	if err != nil {
		return constr != nil, err
	}

	err = io.WriteRewriteInfoFile(newTrace, bug, code, resultIndex)
	if err != nil {
		return constr != nil, err
	}

	return constr != nil, nil
}
