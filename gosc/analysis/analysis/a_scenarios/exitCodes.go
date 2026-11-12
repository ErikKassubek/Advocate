// Copyright (c) 2025 
//
// File: exitCodes.go
// Brief: Check exit codes
//
//
// License: BSD-3-Clause

package a_scenarios

import (
	"gosc/analysis/a_base"
	"gosc/trace"
	"gosc/utils/log"
	"gosc/utils/results/buginfo"
	"gosc/utils/results/results"
	"gosc/utils/timer"
)

// RunAnalysisOnExitCodes checks the exit codes for the recording for actual bugs
//
// Parameter:
//   - all bool: If true, check for all, else only check for the once, that are not detected by the full analysis
func RunAnalysisOnExitCodes(all bool) {
	timer.Start(timer.AnaExitCode)
	defer timer.Stop(timer.AnaExitCode)

	switch a_base.ExitCode {
	case buginfo.ExitCodeCloseNil: // close on nil
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
		results.Result(results.CRITICAL, buginfo.ACloseOnNilChannel,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case buginfo.ExitCodeNegativeWG: // negative wg counter
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
		results.Result(results.CRITICAL, buginfo.ANegWG,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case buginfo.ExitCodeUnlockBeforeLock: // unlock of not locked mutex
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
		results.Result(results.CRITICAL, buginfo.AUnlockOfNotLockedMutex,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case buginfo.ExitCodePanic: // unknown panic
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
		results.Result(results.CRITICAL, buginfo.RUnknownPanic,
			[]results.ResultElem{arg1}, nil)
		a_base.BugWasFound = true
	case buginfo.ExitCodeTimeout: // timeout
		// results.Result(results.CRITICAL, buginfo.RTimeout,
		// 	[]results.ResultElem{}, nil)
	}
	if all {
		switch a_base.ExitCode {
		case buginfo.ExitCodeCloseClose: // close on closed
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
			results.Result(results.CRITICAL, buginfo.ACloseOnClosed,
				[]results.ResultElem{arg1}, nil)
			a_base.BugWasFound = true
		case buginfo.ExitCodeSendClose: // send on closed
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
			results.Result(results.CRITICAL, buginfo.ASendOnClosed,
				[]results.ResultElem{arg1}, nil)
			a_base.BugWasFound = true
		}
	}
}
