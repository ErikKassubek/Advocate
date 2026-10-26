// Copyright (c) 2026 Erik Kassubek
//
// File: modes.go
// Brief: controll the different modes
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package advoc

import (
	"advocate/advoc/toolchain"
	"advocate/fuzzing/f_fuzzing"
	"advocate/utils/flags"
	"advocate/utils/io"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/stats"
)

// modeFuzzing starts the fuzzing
func modeFuzzing() error {
	if flags.ProgName == "" {
		flags.ProgName = paths.GetProgName(flags.ProgPath)
	}

	err := f_fuzzing.Fuzzing()
	if err != nil {
		log.Error("Fuzzing Failed: ", err.Error())
		return err
	}

	return nil
}

// Start point for the toolchain
// This will run, analyze and replay a given program or test
//
// Parameter:
//   - mode string: main for main function, test for test function
//   - record bool: if true, the toolchain will run the recording
//   - analysis bool: if true, the toolchain will run analysis
//   - replay bool: if true, the toolchain will run replays
//
// Note:
//   - If recording is false, but analysis or replay is set, -trace must be set
func modeToolchain(mode string, record bool, analysis bool, replay bool) (err error) {
	if !record && (analysis || replay) {
		flags.TracePath, err = paths.CheckPath(flags.TracePath)
		if err != nil {
			log.Error("Error on checking trace path: ", err)
			return err
		}
	}

	if mode == "test" && !record && replay && flags.ExecName == "" {
		log.Error("When running replay of test without recording, -exec [TestName] must be set")
		panic("When running replay of test without recording, -exec [TestName] must be set")
	}

	firstRun := true
	fileNumber, testNumber := 1, 0
	_, _, err = toolchain.Run(mode, "", record, analysis,
		replay, -1, "", firstRun, fileNumber, testNumber)
	if err != nil {
		return err
	}

	if flags.CreateStatistics {
		err = stats.CreateStatsTotal(flags.ProgPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// Mode to read a trace and show it in a tabular format
func modeShowTrace() error {
	tr, _, _, err := io.CreateTraceFromFiles(flags.TracePath, true)
	if err != nil {
		return err
	}

	tr.PrintAsTab()

	return nil
}
