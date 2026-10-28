// Copyright (c) 2024 Erik Kassubek
//
// File: headerUnitTests.go
// Brief: Functions to add and remove the ADVOCATE header into file containing
//    unit tests
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/analysis/a_base"
	"advocate/utils/flags"
	"advocate/utils/paths"
	"fmt"
)

// Run is the main function for the toolchain
//
// Parameter:
//   - pathToTest string: specify specific test path, only used for fuzzing
//   - runRecord bool: run the recording. If set to false, but runAnalysis or runReplay is
//     set the trace at tracePath is used
//   - runAnalysis bool: run the analysis on a path
//   - runReplay bool: run replay, if runAnalysis is true, those replays are used
//     otherwise the trace at tracePath is replayed
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//   - replayAt bool: replay atomics
//   - stats bool: create statistics
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//
// Returns:
//   - int: TraceID
//   - int: number results
//   - error
func Run(pathToTest string,
	runRecord, runAnalysis, runReplay bool, fuzzing int, fuzzingTrace string,
	firstRun bool, fileNumber, testNumber int) (int, int, error) {

	a_base.Clear()

	switch flags.ModeMain {
	case true:
		if paths.Advocate == "" {
			return 0, 0, fmt.Errorf("Path to advocate required for mode main")
		}
		if paths.Prog == "" {
			return 0, 0, fmt.Errorf("Path to file required")
		}
		if flags.ExecName == "" {
			return 0, 0, fmt.Errorf("Name of the executable required")
		}
		if (flags.CreateStatistics || flags.MeasureTime) && flags.ProgName == "" {
			flags.ProgName = paths.GetProgName(flags.ProgPath)
		}
		return runWorkflowMain(runRecord, runAnalysis, runReplay,
			fuzzing, fuzzingTrace, firstRun)
	case false:
		if paths.Advocate == "" {
			return 0, 0, fmt.Errorf("Path to advocate required")
		}
		if paths.Prog == "" {
			return 0, 0, fmt.Errorf("Path to test folder required for mode main")
		}
		if (flags.CreateStatistics || flags.MeasureTime) && flags.ProgName == "" {
			flags.ProgName = paths.GetProgName(flags.ProgPath)
		}
		return runWorkflowUnit(runRecord, runAnalysis, runReplay,
			pathToTest, fuzzing, fuzzingTrace,
			firstRun, fileNumber, testNumber)
	default:
		return 0, 0, fmt.Errorf("Choose one mode from 'main' or 'test'")
	}
}
