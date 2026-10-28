// Copyright (c) 2025 Erik Kassubek
//
// File: funcs.go
// Brief: Function on data
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_base

import (
	"advocate/trace"
	"strings"
)

// AddMutToQueue adds a mutation to the queue. If a maximum number of mutation runs in set,
// only add the mutation if it does not exceed this max number
//
// Parameter:
//   - mut mutation: the mutation to add
//   - force bool: if true, the mutation is always added, otherwise only if it does not exceed the max number of runs
//
// Returns:
//   - bool: true, if the mutation was added, false otherwise
func AddMutToQueue(mut Mutation, force bool) bool {
	if force || MaxNumberRuns == -1 || NumberFuzzingRuns+MutationQueue.Size() <= MaxNumberRuns {
		MutationQueue.Push(mut)
		return true
	}
	return false
}

// IgnoreFuzzing checks if an element should be ignored for fuzzing
// For the creation of mutations we ignore all elements that do not directly
// correspond to relevant operations. Those are new, replay, routineEnd
//
// Parameter:
//   - elem *trace.TraceElementFork: The element to check
//   - ignoreNew bool: if true, new elem is ignored elem, otherwise not
//
// Returns:
//   - True if the element is of one of those types, false otherwise
func IgnoreFuzzing(elem trace.Element, ignoreNew bool) bool {
	t := elem.Type(false)
	return (ignoreNew && t == trace.New) || t == trace.Replay || elem.Type(true) == trace.RoutEnd || strings.Contains(elem.File(), "goPatch/src/")
}
