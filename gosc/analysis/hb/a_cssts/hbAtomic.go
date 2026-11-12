// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the cssts for atomics
//
//
// License: BSD-3-Clause

package a_cssts

import (
	"gosc/analysis/a_base"
	"gosc/trace"
	"gosc/utils/log"
)

// UpdateHBAtomic update the cssts for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateHBAtomic(at *trace.ElementAtomic) {
	switch at.Type(true) {
	case trace.AtomicLoad, trace.AtomicSwap, trace.AtomicCompAndSwap:
		Read(at, true)
	case trace.AtomicStore, trace.AtomicAdd, trace.AtomicAnd, trace.AtomicOr:
		// csst does not add an edge for write
	default:
		err := "Unknown operation: " + at.String()
		log.Error(err)
	}
}

// Read calculates the new vector clock for a read operation and update cv
//
// Parameter:
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
//   - sync bool: sync reader with last writer
func Read(at *trace.ElementAtomic, sync bool) {
	id := at.ResourceID()

	if sync && a_base.LastAtomicWriter[id] != nil {
		AddEdge(at, a_base.LastAtomicWriter[id], false)
	}
}
