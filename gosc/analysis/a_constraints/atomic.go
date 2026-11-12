// Copyright (c) 2025 
//
// File: atomic.go
// Brief: Create constraint from atomics
//
//
// License: BSD-3-Clause

package a_constraints

import (
	"gosc/analysis/a_base"
	"gosc/trace"
)

// AddAtomic  add the element to be the last write for an atomic write
// For an atomic read, add a constraint between the last writer and the element
//
// Parameter:
//   - elem *trace.ElementAtomic: the atomic trace element
func AddAtomic(elem *trace.ElementAtomic) {
	// all operation other than store, read from the atomic variable
	// Set a constraint with the last writer. If there is non, the variable
	// tries to read a default value, which does not create an constraint
	if elem.Type(true) != trace.AtomicStore {
		if lw, ok := a_base.LastAtomicWriter[elem.ResourceID()]; ok {
			AddConstraint(true, lw, elem)
		}
	}

	// all operations other than load write to the atomic variable
	// set as last writer
	if elem.Type(true) != trace.AtomicLoad {
		a_base.LastAtomicWriter[elem.ResourceID()] = elem
	}
}
