// Copyright (c) 2025 
//
// File: pos.go
// Brief: Positive constraints
//
//
// License: BSD-3-Clause

package a_constraints

import "gosc/trace"

// A positive constraint consisting of two elements
// The positive constraints signals, that the first element influences the
// second, e.g. for an atomic operation, the second element reads from the first
// or the recv in the second element receives from the send in the first element
type posConstraint2 struct {
	first  trace.Element
	second trace.Element
}

// Return if the constraint is positive.
// Always return true
// Required to implement constraint interface
func (this posConstraint2) isPos() bool {
	return true
}
