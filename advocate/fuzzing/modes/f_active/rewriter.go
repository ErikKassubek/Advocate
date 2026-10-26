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
	"advocate/fuzzing/f_roc"
	"advocate/utils/helper"
	"advocate/utils/results/bugs"
	"errors"
)

// RewriteTrace creates a new trace from the given bug
//
// Parameter:
//   - tr *trace.Trace: Pointer to the trace to rewrite
//   - bug Bug: The bug to create a trace for
//   - rewrittenBugs *map[bugs.ResultType][]string: map of already rewritten bugs
//
// Returns:
//   - bool: true if rewrite was needed, false otherwise (e.g. actual bug, warning)
//   - code: expected exit code
//   - error: An error if the trace could not be created
func RewriteTrace(bug bugs.Bug, rewrittenBugs map[helper.ResultType][]string) (constr *f_roc.Constraint, code int, err error) {
	code = helper.ExitCodeNone

	constr = nil

	switch bug.Type {

	// ACTUAL BUGS
	case helper.ASendOnClosed:
		err = errors.New("Actual send on closed. Therefore no rewrite is needed")
	case helper.ARecvOnClosed:
		err = errors.New("Actual receive on closed in trace. Therefore no rewrite is needed")
	case helper.ACloseOnClosed:
		err = errors.New("Actual close on close detected. Therefore no rewrite is needed")
	case helper.ACloseOnNilChannel:
		err = errors.New("Actual close on nil detected. Therefore no rewrite is needed")
	case helper.ANegWG:
		err = errors.New("Actual negative Wait Group. Therefore no rewrite is needed")
	case helper.AUnlockOfNotLockedMutex:
		err = errors.New("Actual unlock of not locked mutex. Therefore no rewrite is needed")
	case helper.ABlocking:
		err = errors.New("Actual blocking routine. Therefore no rewrite is needed")
	case helper.ADeadlock:
		err = errors.New("Actual deadlock. Therefore no rewrite is needed")
	case helper.AConcurrentRecv:
		err = errors.New("Rewriting trace for concurrent receive is not possible")
	case helper.PSendOnClosed:
		code = helper.ExitCodeSendClose
		constr, err = rewriteClosedChannel(bug) // TODP: rewrite
	case helper.PNegWG:
		code = helper.ExitCodeNegativeWG
		// err = rewriteGraph(tr, bug, code)  // TODO: rewrite
	case helper.PUnlockBeforeLock:
		code = helper.ExitCodeUnlockBeforeLock
		// err = rewriteGraph(tr, bug, code)  // TODO: rewrite
	case helper.PCyclicDeadlock:
		constr, err = rewriteCyclicDeadlock(bug) // TODO: rewrite
	case helper.PMixedDeadlock:
		code = helper.ExitCodeMixedDeadlock
		// err = rewriteMixedDeadlock(tr, bug, code)  // TODO: rewrite

	// LEAKS
	case helper.LUnknown, helper.LChan, helper.LSelect, helper.LNilChan, helper.LMutex, helper.LWaitGroup, helper.LCond:
		err = errors.New("Leak. No rewrite")
	case helper.RUnknownPanic:
		err = errors.New("Unknown panic. No rewrite possible")
	case helper.RTimeout:
		err = errors.New("Timeout. No rewrite possible")
	default:
		err = errors.New("For the given bug type no trace rewriting is implemented")
	}

	return constr, code, err
}
