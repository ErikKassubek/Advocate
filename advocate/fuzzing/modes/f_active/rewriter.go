// Copyright (c) 2024 Erik Kassubek
//
// File: rewriter.go
// Brief: Main functions to rewrite the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

// Package rewriter provides functions for rewriting traces.
package f_active

import (
	"advocate/fuzzing/f_constraints"
	"advocate/utils/results/buginfo"
	"advocate/utils/results/bugs"
	"errors"
)

// RewriteTrace creates a new trace from the given bug
//
// Parameter:
//   - tr *trace.Trace: Pointer to the trace to rewrite
//   - bug Bug: The bug to create a trace for
//   - rewrittenBugs *map[buginfo.ResultType][]string: map of already rewritten bugs
//
// Returns:
//   - bool: true if rewrite was needed, false otherwise (e.g. actual bug, warning)
//   - code: expected exit code
//   - error: An error if the trace could not be created
func RewriteTrace(bug bugs.Bug, rewrittenBugs map[buginfo.ResultType][]string) (constr *f_constraints.Constraint, code int, err error) {
	code = buginfo.ExitCodeNone

	constr = nil

	switch bug.Type {

	// ACTUAL BUGS
	case buginfo.ASendOnClosed:
		err = errors.New("Actual send on closed. Therefore no rewrite is needed")
	case buginfo.ACloseOnClosed:
		err = errors.New("Actual close on close detected. Therefore no rewrite is needed")
	case buginfo.ACloseOnNilChannel:
		err = errors.New("Actual close on nil detected. Therefore no rewrite is needed")
	case buginfo.ANegWG:
		err = errors.New("Actual negative Wait Group. Therefore no rewrite is needed")
	case buginfo.AUnlockOfNotLockedMutex:
		err = errors.New("Actual unlock of not locked mutex. Therefore no rewrite is needed")
	case buginfo.ABlocking:
		err = errors.New("Actual blocking routine. Therefore no rewrite is needed")
	case buginfo.ADeadlock:
		err = errors.New("Actual deadlock. Therefore no rewrite is needed")
	case buginfo.AConcurrentRecv:
		err = errors.New("Rewriting trace for concurrent receive is not possible")
	case buginfo.PSendOnClosed:
		code = buginfo.ExitCodeSendClose
		constr, err = rewriteClosedChannel(bug)
	case buginfo.PNegWG:
		code = buginfo.ExitCodeNegativeWG
		constr, err = rewriteGraph(bug)
	case buginfo.PUnlockBeforeLock:
		code = buginfo.ExitCodeUnlockBeforeLock
		constr, err = rewriteGraph(bug)
	case buginfo.PCyclicDeadlock:
		constr, err = rewriteCyclicDeadlock(bug)
	case buginfo.PMixedDeadlock:
		constr, err = rewriteMixedDeadlock(bug)

	// LEAKS
	case buginfo.LUnknown, buginfo.LChan, buginfo.LSelect, buginfo.LNilChan, buginfo.LMutex, buginfo.LWaitGroup, buginfo.LCond:
		err = errors.New("Leak. No rewrite")
	case buginfo.RUnknownPanic:
		err = errors.New("Unknown panic. No rewrite possible")
	case buginfo.RTimeout:
		err = errors.New("Timeout. No rewrite possible")
	case buginfo.RReplayFail:
		err = errors.New("RReplayFail. No rewrite possible")
	default:
		err = errors.New("For the given bug type no trace rewriting is implemented")
	}

	return constr, code, err
}
