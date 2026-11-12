// Copyright (c) 2025 
//
// File: constraints.go
// Brief: Main file for fuzzing using constraints
//
//
// License: BSD-3-Clause

package a_constraints

import (
	"gosc/trace"
	"fmt"
)

var (
	allConstraints = make([]constraint, 0)
)

// A constraint consisting of up to two elements
//
// The positive constraints signals, that the first element influences the
// second, e.g. for an atomic operation, the second element reads from the first
// or the recv in the second element receives from the send in the first element
//
// The negative constraints signals, that the first element does not influences the
// second, e.g. for an atomic operation, the second element does not reads from the first
// or the recv in the second element does not receives from the send in the first element
//
// Fields:
//   - first trace.Element: the first element in the constraint
//   - second trace.Element: the second element in the constraint
//   - pos bool: true if it is a positive constraint, false if it is a negative constraint
//   - twoElem bool: if true, it is a constraint of two elements, otherwise of one
type constraint struct {
	first   trace.Element
	second  trace.Element
	pos     bool
	twoElem bool
}

// AddConstraint adds a new constraint to the set of constraints
//
// Parameter:
//   - pos bool: true for a positive constraint, false for a negative
//   - first trace.Element: the first element of the constraint
//   - second trace.Element: the second element of the constraint of nil
func AddConstraint(pos bool, first, second trace.Element) {
	twoElem := true
	if second == nil {
		twoElem = false
	}

	c := constraint{
		first:   first,
		second:  second,
		pos:     pos,
		twoElem: twoElem,
	}

	allConstraints = append(allConstraints, c)
}

// Return a string representation of the constraint
//
// Returns:
//   - string: a string representation of the form:
//     [type];[obj];[elem1];[elem2];
//     where
//     [type] is + if it is a pos constraint or - if it is a negative
//     [obj] is the type of object the constaint is build on, e.g. C for channel, M for mutex, A for atomic
//     [elem1] is a representation of the first element in the form [routine],[file]:line\
//     [elem2] is a representation of the second element in the same form or empty if twoElem is false
func (this *constraint) toString() string {
	res := ""
	if this.pos {
		res += "+;"
	} else {
		res += "-;"
	}

	res += fmt.Sprintf("%s;", this.first.Type(false))

	res += fmt.Sprintf("%d,%s;", this.first.RoutineID(), this.first.Pos())
	if this.twoElem {
		res += fmt.Sprintf("%d,%s", this.second.RoutineID(), this.second.Pos())
	}

	return res
}

// flip turns a positive constraint into a negative one and a negative constraint
// into a positive
func (this *constraint) flip() {
	this.pos = !this.pos
}
