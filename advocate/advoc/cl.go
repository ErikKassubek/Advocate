// Copyright (c) 2024 Erik Kassubek
//
// File: cl.go
// Brief: command line entries
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package advoc

import (
	"advocate/fuzzing/f_base"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"flag"
	"os"
	"strings"
)

var (
	help bool
)

// CommandLine reads and processes the command line arguments
//
// Returns:
//   - bool: true if execution can continue, false if it should stop, e.g. because help was selected/shown
func CommandLine() bool {
	flag.BoolVar(&help, "h", false, "Print help")
	flag.BoolVar(&help, "help", false, "Print help")

	flag.StringVar(&flags.ProgPath, "path", "", "Path to the program folder, for main: path to main file, for test: path to test folder")
	flag.StringVar(&flags.RootPath, "root", "", "Path to the root of the program folder. Must only be set if different from path")

	flag.StringVar(&flags.ProgName, "prog", "", "Name of the program")
	flag.StringVar(&flags.ExecName, "exec", "", "Name of the executable or test")

	flag.StringVar(&flags.TracePath, "trace", "", "Path to the trace folder to replay")

	flag.IntVar(&flags.Timeout, "timeoutRun", 180, "Set the timeout in seconds for one recording/replay. Default: 180s. To disable set to -1")
	flag.IntVar(&flags.TimeoutFuzzing, "timeoutFuz", 420, "Timeout of fuzzing per test/program in seconds. Default: 420s. To Disable, set to -1")
	flag.IntVar(&flags.MaxFuzzingRun, "maxFuzzingRuns", -1, "Maximum number of fuzzing runs per test/prog. Default: -1. To Disable, set to -1")
	flag.IntVar(&flags.MaxNumberElements, "maxNumberElements", 10000000, "Set the maximum number of elements in a trace. Traces with more elements will be skipped. To disable set -1. Default: 10000000")

	flag.BoolVar(&flags.RecordPlus, "recordPlus", false, "Record functions and flow controll operations.")

	flag.BoolVar(&flags.MeasureTime, "time", false, "measure the runtime")
	flag.BoolVar(&flags.CreateStatistics, "stats", false, "Create statistics.")
	flag.BoolVar(&flags.NotExecuted, "notExec", false, "Find never executed operations")

	flag.BoolVar(&flags.IgnoreCriticalSection, "ignoreCritSec", false, "Ignore happens before relations of critical sections (default false)")
	flag.BoolVar(&flags.IgnoreAtomics, "ignoreAtomics", false, "Ignore atomic operations (default false). Use to reduce memory header for large traces.")
	flag.BoolVar(&flags.OnlyAPanicAndLeak, "onlyActual", false, "only test for actual bugs leading to panic and actual leaks. This will overwrite `scen`")

	flag.BoolVar(&flags.NoSkipRewrite, "replayAll", false, "Replay a bug even if it has already been confirmed")
	flag.BoolVar(&flags.NoRewrite, "noRewrite", false, "Do not rewrite the trace file (default false)")
	flag.BoolVar(&flags.DeleteTraces, "deleteTrace", false, "If set, the traces are deleted after analysis.")
	flag.BoolVar(&flags.SkipExisting, "skipExisting", false, "If set, all tests that already have a results folder will be skipped. Also skips failed tests.")

	flag.BoolVar(&flags.Continue, "cont", false, "Continue a partial analysis of tests")

	flag.BoolVar(&flags.NoWarning, "noWarning", false, "Only show critical bugs")
	flag.BoolVar(&flags.NoInfo, "noInfo", false, "Do not show infos in the terminal (will only show results, errors, important and progress)")
	flag.BoolVar(&flags.NoProgress, "noProgress", false, "Do not show progress info")
	flag.BoolVar(&flags.Output, "output", false, "Show the output of the executed programs in the terminal. Otherwise it is only in output.log file.")

	flag.BoolVar(&flags.AlwaysPanic, "panic", false, "Panic if the analysis panics")
	flag.BoolVar(&flags.NoMemorySupervisor, "noMemorySupervisor", false, "Disable the memory supervisor")

	flag.BoolVar(&flags.HideFile, "hideFile", false, "Do not show the file name. Only used for 'show' mode")

	flag.StringVar(&flags.Scenarios, "scen", "", "Select which analysis scenario to run, e.g. -scen srd for the option s, r and d."+
		"If not set, all scenarios are run.\n"+
		"Options:\n"+
		"\ts: Send on closed channel\n"+
		// "\tr: Receive on closed channel\n"+
		"\tw: Done before add on waitGroup\n"+
		"\tn: Close of closed channel\n"+
		// "\tb: Concurrent receive on channel\n"+
		"\tl: Leaking routine\n"+
		"\tu: Unlock of unlocked mutex\n"+
		"\tc: Cyclic deadlock\n"+
		"\tm: Mixed deadlock\n",
	)

	flag.StringVar(&flags.FuzzingMode, "mode", "",
		"Mode for fuzzing. Possible values are:\n\tGuided\n\tGFuzz\n\tGFuzzHB\n\tGFuzzHBFlow\n\tFlow\n\tGoPie\n\tGoCR\n\tGoCRHB\n\tDefault: Guided")

	flag.BoolVar(&flags.ModeMain, "main", false, "set to run on main function")

	flag.StringVar(&flags.Settings, "settings", "", "Set some internal settings. For more info, see ../doc/usage.md")

	flag.BoolVar(&flags.CheckBenign, "benign", false, "Check if blocking bug is benign")

	// for experiments
	flag.BoolVar(&f_base.FinishIfBugFound, "finishIfBugFound", false, "Finish fuzzing as soon as a bug was found")
	flag.BoolVar(&flags.CancelTestIfBugFound, "cancelTestIfBugFound", false, "Skip further fuzzing runs of a test if one bug has been found")

	flag.Parse()

	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		flags.Mode = os.Args[1]
		flag.CommandLine.Parse(os.Args[2:])
		if help {
			helper.PrintHelpMode(flags.Mode)
			return false
		}
	} else {
		if help {
			helper.PrintHelp()
			return false
		}
		helper.PrintHelp()
		return false
	}

	return true

}
