// Copyright (c) 2026 Erik Kassubek
//
// File: run.go
// Brief: start/controll the different modes
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package advoc

import (
	"advocate/analysis/a_base"
	"advocate/fuzzing/f_base"
	"advocate/fuzzing/f_fuzzing"
	"advocate/utils/command"
	"advocate/utils/consts"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/settings"
	"advocate/utils/timer"
	"fmt"
	"os"
)

const (
	record   = true
	analysis = true
	replay   = true
)

// Run starts the execution of advocate
func Run() error {

	// If -main is set, the path needs to be the path to the main file
	// If the given path is to a folder, check if a main.go file exists in this folder
	// If so, fix the path. Otherwise return error and finish
	var err error
	if flags.ModeMain {
		flags.ProgPath, err = paths.GetMainPath(flags.ProgPath)
		if err != nil {
			log.Error("Could not find main file. If -main is set, -path should point to the main file.")
			log.Error(err)

			return err
		}
	}

	settings.SetSettings()
	paths.BuildPaths(flags.ModeMain)

	err = CheckBin()
	if err != nil {
		return err
	}

	timer.Start(timer.Total)
	defer timer.Stop(timer.Total)

	control.SetMaxNumberElem()
	if !flags.NoMemorySupervisor {
		go control.Supervisor(a_base.ClearTrace, a_base.ClearData, f_fuzzing.ResetFuzzing) // cancel analysis if not enough ram
	}

	// don't run any HB Analysis for direct GFuzz, GoPie and GoCR
	if flags.Mode == consts.ModeFuzzing && (flags.FuzzingMode == f_base.GFuzz ||
		flags.FuzzingMode == f_base.GoPie || flags.FuzzingMode == f_base.GoCR) {
		flags.Scenarios = "-"
		flags.OnlyAPanicAndLeak = true
	}

	if flags.FuzzingMode == "" {
		flags.FuzzingMode = f_base.Guided
	}

	a_base.AnalysisCasesMap, err = flags.ParseAnalysisCases()
	if err != nil {
		log.Error("Could not read analysis cases: ", err)
		return err
	}

	if flags.Mode != "show" {
		CheckProg()
		command.RunGoModTidy()
	}

	if flags.ModeMain && flags.ExecName == "" {
		log.Error("Could not determine executable name from go.mod. Provide with -exec [ExecutableName]")
		return fmt.Errorf("Could not determine executable name")
	}

	switch flags.Mode {
	case "analysis":
		flags.Mode = consts.ModeAnalysis
		err = modeToolchain(record, analysis, replay)
	case "fuzzing":
		flags.Mode = consts.ModeFuzzing
		err = modeFuzzing()
	case "record", "recording", "trace":
		flags.Mode = consts.ModeRecord
		flags.DeleteTraces = false
		err = modeToolchain(record, !analysis, !replay)
	case "replay":
		flags.Mode = consts.ModeReplay
		err = modeToolchain(!record, !analysis, replay)
	case "experiment", "experiments":
		flags.Mode = consts.ModeExperiment
		err = modeExperiments()
	case "show":
		flags.Mode = consts.ModeShowTrace
		err = modeShowTrace()
		return nil
		// case "static": // TODO: only for testing, when working move into analysis
		// flags.Mode = consts.ModeReplay
	// 	flags.DeleteTraces = false
	// 	err = modeToolchain(modeMainTest, record, analysis, !replay)
	// 	err = s_blocking.BuildStaticBlockingAnalysis()

	default:
		log.Errorf("Unknown mode %s\n", os.Args[1])
		log.Error("Select one mode from  'analysis', 'fuzzing', 'replay', 'record', experiments or 'gui' (only if -tags gui is set)")
		err = fmt.Errorf("Unknown mode %s", os.Args[1])
		helper.PrintHelp()
	}

	if err != nil {
		return err
	}

	numberBugs, _, numberTestWithRes, numberErr, numberTimeout := log.GetLoggingNumbers()
	if numberErr == 0 {
		log.Info("Finished with 0 errors")
	} else {
		log.Errorf("Finished with %d errors", numberErr)
	}
	if numberTimeout == 0 {
		log.Info("No internal replay timeouts occurred")
	} else {
		log.Errorf("%d internal replay timeouts occurred", numberTimeout)
	}
	if flags.Mode == consts.ModeAnalysis || flags.Mode == consts.ModeFuzzing {
		if numberTestWithRes == 0 {
			log.Info("No bugs have been found/indicated")
		} else {
			if !flags.ModeMain {
				log.Resultf(false, false, "", "Tests with indicated bugs: %d", numberTestWithRes)
			}
			log.Resultf(false, false, "", "Number indicated bugs:  %d", numberBugs)
		}
	}
	timer.UpdateTimeFileOverview("*Total*")

	return nil
}

func Initialize() {
	progPathDir := paths.GetDirectory(flags.ProgPath)
	timer.Init(progPathDir)
}
