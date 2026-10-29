// Copyright (c) 2024 Erik Kassubek
//
// File: closed.go
// Brief: Rewrite traces for send and receive on closed channel
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_active

import (
	"advocate/fuzzing/f_constraints"
	"advocate/utils/log"
	"advocate/utils/results/bugs"
	"errors"
)

// Given a send/recv on a closed channel, rewrite the trace to make the bug
// occur.
//
// Parameter:
//   - trace *trace.Trace: Pointer to the trace to rewrite
//   - bug Bug: The bug to create a trace for
//   - exitCode int: The exit code to use for the stop marker
//
// Returns:
//   - f_constraints.Constraint: the constraint
//   - error: An error if the trace could not be created
func rewriteClosedChannel(bug bugs.Bug) (*f_constraints.Constraint, error) {
	log.Info("Start rewrite for send on closed channel...")

	constr := f_constraints.NewConstraint()

	if len(bug.TraceElement1) == 0 || bug.TraceElement1[0] == nil { // send
		return &constr, errors.New("Send is nil") // send/recv
	}
	if len(bug.TraceElement2) == 0 || bug.TraceElement2[0] == nil { // close
		return &constr, errors.New("Close is nil")
	}

	constr.Add(bug.TraceElement2[0])
	constr.Add(bug.TraceElement1[0])

	return &constr, nil
}
