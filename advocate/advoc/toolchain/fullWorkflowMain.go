// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: runFullWorkflowMain.go
// Brief: Function to run the whole ADVOCATE workflow, including running,
//    analysis and replay on a program with a main function
//
// Author: Erik Kassubek, Mario Occhinegro
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/utils/command"
	"advocate/utils/consts"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/complete"
	"advocate/utils/results/stats"
	"advocate/utils/timer"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Run ADVOCATE on a program with a main function
//
// Parameter:
//   - runRecord bool: run the recording. If set to false, but runAnalysis or runReplay is
//     set the trace at tracePath is used
//   - runAnalysis bool: run the analysis on a path
//   - runReplay bool: run replay, if runAnalysis is true, those replays are used
//     otherwise the trace at tracePath is replayed
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//   - onlyRecord bool: if true, only record th trace, but do not run any analysis
//
// Returns:
//   - int: TraceID
//   - int: number results
//   - error
func runWorkflowMain(
	runRecord, runAnalysis, runReplay bool,
	fuzzing int, fuzzingTrace string,
	firstRun bool) (int, int, error) {
	if _, err := os.Stat(paths.Prog); os.IsNotExist(err) {
		return 0, 0, fmt.Errorf("file %s does not exist", paths.Prog)
	}

	log.Info("Run main")

	// Change to the directory of the main file
	if err := os.Chdir(paths.ProgDir); err != nil {
		return 0, 0, fmt.Errorf("Failed to change directory: %v", err)
	}

	if firstRun {
		os.RemoveAll("advocateResult")
		if err := os.MkdirAll("advocateResult", os.ModePerm); err != nil {
			return 0, 0, fmt.Errorf("Failed to create advocateResult directory: %v", err)
		}

		// Remove possibly leftover traces from unexpected aborts that could interfere with replay
		RemoveTraces(paths.ProgDir)
		removeLogs(paths.ProgDir)
	}

	outFile, err := os.OpenFile(paths.NameOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to open log file: %v", err)
	}
	defer outFile.Close()

	origStdout := os.Stdout
	origStderr := os.Stderr

	os.Stdout = outFile
	os.Stderr = outFile

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Set GOROOT environment variable
	if err := os.Setenv("GOROOT", paths.GoPatch); err != nil {
		return 0, 0, fmt.Errorf("Failed to set GOROOT: %v", err)
	}
	// Unset GOROOT
	defer os.Unsetenv("GOROOT")
	if runRecord {
		// build the program
		if flags.MeasureTime && fuzzing < 1 {
			log.Info("Build Program")
			fmt.Printf("%s build\n", paths.Go)
			if err := command.RunCommand(origStdout, origStderr, command.NoTimeout, paths.Go, "build"); err != nil {
				log.Error("Error in building program, removing header and stopping workflow")
				return 0, 0, err
			}

			// run the program
			log.Info("Execute Program")
			timer.Start(timer.Run)
			execPath := paths.MakePathLocal(flags.ExecName)
			if err := command.RunCommand(origStdout, origStderr, command.NoTimeout, execPath); err != nil {
				log.Error("Error in Executing Program: ", err.Error())
			}
			timer.Stop(timer.Run)
		}

		// Build args
		buildFlags, err := buildArgsMain(paths.Prog, false, "1", false, fuzzing, fuzzingTrace, false)
		if err != nil {
			return 0, 0, fmt.Errorf("Error in adding header: %v", err)
		}

		// build the program
		log.Info("Build program for execution")
		if err := command.RunCommand(origStdout, origStderr, command.NoTimeout, paths.Go, "build", buildFlags); err != nil {
			log.Error("Error in building program, removing header and stopping workflow")
			return 0, 0, err
		}
		// run the recording
		log.Info("Run program for execution")
		timer.Start(timer.Recording)
		execPath := paths.MakePathLocal(flags.ExecName)
		command.RunCommand(origStdout, origStderr, command.NoTimeout, execPath)

		timer.Stop(timer.Recording)
	}

	// Apply analyzer
	if runAnalysis {
		analyzerOutput := filepath.Join(paths.ProgDir, paths.NameTrace)

		err = runAnalyzer(analyzerOutput, paths.NameResultReadable, paths.NameResultMachine,
			consts.RewrittenConsts, fuzzing, "", "main", false)

		if err != nil {
			return 0, 0, err
		}
	}

	rewrittenTraces := make([]string, 0)
	total := fuzzing != -1
	if runReplay {
		log.Info("Run replay")
		// Find rewrittenTrace directories
		if runAnalysis {
			rewrittenTraces, err = filepath.Glob(filepath.Join(paths.ProgDir, consts.RewrittenConsts+"*"))
			if err != nil {
				return 0, 0, fmt.Errorf("Error finding rewritten traces: %v", err)
			}
		} else {
			if flags.TracePath != "" {
				rewrittenTraces = append(rewrittenTraces, flags.TracePath)
			}
		}

		collect(paths.ProgDir, paths.ProgDir, paths.Result, total, false)

		timer.Start(timer.Replay)
		for _, trace := range rewrittenTraces {
			traceNum := extractTraceNum(trace)
			buildFlags, err := buildArgsMain(paths.Prog, true, traceNum, true, fuzzing, fuzzingTrace, false)
			if err != nil {
				return 0, 0, err
			}

			// build the program
			log.Info("Build program for replay")
			if err := command.RunCommand(origStdout, origStderr, command.NoTimeout, paths.Go, "build", buildFlags); err != nil {
				log.Error("Error in building program, removing header and stopping workflow")
				continue
			}

			// run the program
			log.Info("Run program for replay")
			execPath := paths.MakePathLocal(flags.ExecName)
			command.RunCommand(origStdout, origStderr, command.NoTimeout, execPath)

			analyzerOutput := filepath.Join(paths.ProgDir, paths.NameTrace)

			err = runAnalyzer(analyzerOutput, paths.NameResultReadable, paths.NameResultMachine,
				"", fuzzing, "", "main", true)

			if err != nil {
				return 0, 0, err
			}
		}
		timer.Stop(timer.Replay)
	}

	if flags.DeleteTraces && !flags.CreateStatistics {
		RemoveTraces(paths.Prog)
	}

	collect(paths.ProgDir, paths.ProgDir, paths.Result, total, true)

	// Generate Bug Reports
	var numberResults int
	if runAnalysis {
		log.Info("Generate Bug Reports")
		numberResults = generateBugReports(movedTraces, fuzzing)

		timer.UpdateTimeFileDetail("Main", len(rewrittenTraces))
	}

	if flags.NotExecuted {
		complete.Check(paths.Result, paths.Prog)
	}

	if flags.CreateStatistics {
		stats.CreateStats("", movedTraces, fuzzing)
	}

	if total {
		removeLogs(paths.Prog)
	}

	return movedTraces, numberResults, nil
}

// Given a path to a trace file, return the trace number
//
// Parameter:
//   - tracePath: path to the file
//
// Returns:
//   - string: trace number
func extractTraceNum(tracePath string) string {
	re := regexp.MustCompile(`[0-9]+$`)
	return re.FindString(tracePath)
}
