// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: headerUnitTests.go
// Brief: Functions to generate bug reports
//
// Author: Erik Kassubek, Mario Occhinegro
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/utils/log"
	"advocate/utils/results/explanation"
)

var movedTraces int = 0

// Generate the bug reports
//
// Parameter:
//   - traceID int: id of the trace
//   - fuzzingRun int: number of fuzzing run, -1 for not fuzzing
//
// Returns:
//   - numberResults int: numberResults
func generateBugReports(traceID, fuzzing int) int {
	numberResults, err := explanation.CreateOverview(true, traceID, fuzzing)
	if err != nil {
		log.Error("Error creating explanation: ", err.Error())
	}
	return numberResults
}
