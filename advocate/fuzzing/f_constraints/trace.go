// Copyright (c) 2026 Erik Kassubek
//
// File: trace.go
// Brief: Function to turn a trace into a full constratin
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_constraints

import (
	"advocate/trace"
)

func TraceToConstraint(tr *trace.Trace) *Constraint {
	trIter := tr.AsIterator()

	res := NewConstraint()

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		if !elem.Replayable() {
			continue
		}

		res.Add(elem)
	}

	return &res
}
