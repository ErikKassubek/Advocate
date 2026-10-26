// Copyright (c) 2024 Erik Kassubek
//
// File: cyclicDeadlock.go
// Brief: Rewrite trace for cyclic deadlocks
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_active

import (
	"advocate/fuzzing/f_roc"
	"advocate/trace"
	"advocate/utils/results/bugs"
	"errors"
)

// TODO: rewrite with constraints and closure

// rewriteCyclicDeadlock rewrites the trace in such a way, that it should
// trigger the cyclic/resource deadlock described in the bug
//
// Parameter:
//   - bug bugs.Bug: the bug that should be triggered by the rewrite
//
// Returns:
//   - *f_roc.Constraint: the constraint containing the bug
//   - error
func rewriteCyclicDeadlock(bug bugs.Bug) (*f_roc.Constraint, error) {
	if len(bug.TraceElement2) == 0 {
		return nil, errors.New("no trace elements in bug")
	}

	if len(bug.TraceElement2) < 2 {
		return nil, errors.New("at least 2 trace elements are needed for a deadlock")
	}

	constr := f_roc.NewConstraint()

	for _, elem := range bug.TraceElement2 { // acquired
		constr.Add(elem)
	}

	for _, elem := range bug.TraceElement1 { // blocking
		constr.Add(elem)
	}

	return &constr, nil
}

// findLastTime returns the latest time stamp from the bug elements
//
// Parameters:
//   - bugElements []trace.TraceElement: the bug element to search through
//
// Returns:
//   - int: the highest tPost from the bug elements
func findLastTime(bugElements []trace.Element) int {
	lastTime := -1

	for _, e := range bugElements {
		if lastTime == -1 || e.T(trace.Sorting) > lastTime {
			lastTime = e.T(trace.Sorting)
		}
	}
	return lastTime
}
