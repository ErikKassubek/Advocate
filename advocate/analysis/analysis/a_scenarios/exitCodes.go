// Copyright (c) 2025 Erik Kassubek
//
// File: exitCodes.go
// Brief: Check exit codes
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/results/results"
	"advocate/utils/timer"
)

// RunAnalysisOnExitCodes checks the exit codes for the recording for actual bugs
//
// Parameter:
//   - all bool: If true, check for all, else only check for the once, that are not detected by the full analysis
func RunAnalysisOnExitCodes(all bool) {
	timer.Start(timer.AnaExitCode)
	defer timer.Stop(timer.AnaExitCode)

	switch a_base.ExitCode {
	case helper.ExitCodeCloseClose: // close on closed
		file, line, err := trace.PosFromPosString(a_base.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}

		arg1 := results.TraceElementResult{
			RoutineID:  0,
			ResourceID: 0,
			TRequest:   0,
			ObjType:    "CC",
			File:       file,
			Line:       line,
		}
		results.Result(results.CRITICAL, helper.ACloseOnClosed,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case helper.ExitCodeCloseNil: // close on nil
		file, line, err := trace.PosFromPosString(a_base.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}
		arg1 := results.TraceElementResult{
			RoutineID:  0,
			ResourceID: 0,
			TRequest:   0,
			ObjType:    "CC",
			File:       file,
			Line:       line,
		}
		results.Result(results.CRITICAL, helper.ACloseOnNilChannel,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case helper.ExitCodeNegativeWG: // negative wg counter
		file, line, err := trace.PosFromPosString(a_base.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}
		arg1 := results.TraceElementResult{
			RoutineID:  0,
			ResourceID: 0,
			TRequest:   0,
			ObjType:    "WD",
			File:       file,
			Line:       line,
		}
		results.Result(results.CRITICAL, helper.ANegWG,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case helper.ExitCodeUnlockBeforeLock: // unlock of not locked mutex
		file, line, err := trace.PosFromPosString(a_base.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}
		arg1 := results.TraceElementResult{
			RoutineID:  0,
			ResourceID: 0,
			TRequest:   0,
			ObjType:    "ML",
			File:       file,
			Line:       line,
		}
		results.Result(results.CRITICAL, helper.AUnlockOfNotLockedMutex,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case helper.ExitCodePanic: // unknown panic
		file, line, err := trace.PosFromPosString(a_base.ExitPos)
		if err != nil {
			log.Error("Could not read exit pos: ", err)
		}
		arg1 := results.TraceElementResult{
			RoutineID:  0,
			ResourceID: 0,
			TRequest:   0,
			ObjType:    "XX",
			File:       file,
			Line:       line,
		}
		results.Result(results.CRITICAL, helper.RUnknownPanic,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case helper.ExitCodeTimeout: // timeout
		// results.Result(results.CRITICAL, helper.RTimeout,
		// 	[]results.ResultElem{}, nil)
	}
	if all {
		if a_base.ExitCode == helper.ExitCodeSendClose { // send on closed
			file, line, err := trace.PosFromPosString(a_base.ExitPos)
			if err != nil {
				log.Error("Could not read exit pos: ", err)
			}
			arg1 := results.TraceElementResult{ // send
				RoutineID:  0,
				ResourceID: 0,
				TRequest:   0,
				ObjType:    "CS",
				File:       file,
				Line:       line,
			}
			results.Result(results.CRITICAL, helper.ASendOnClosed,
				[]results.ResultElem{arg1}, nil)
			a_base.BugWasFound = true
		}
	}
}
