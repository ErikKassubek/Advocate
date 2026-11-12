// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the pog for atomics
//
//
// License: BSD-3-Clause

package a_pog

import (
	"gosc/trace"
	"gosc/utils/log"
)

// UpdateHBAtomic update the pog for an atomic operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateHBAtomic(graph *PoGraph, at *trace.ElementAtomic) {
	switch at.Type(true) {
	case trace.AtomicLoad:
		Read(graph, at)
	case trace.AtomicSwap, trace.AtomicCompAndSwap:
		Read(graph, at)
		Write(graph, at)
	case trace.AtomicStore, trace.AtomicAdd, trace.AtomicAnd, trace.AtomicOr:
		Write(graph, at)
	default:
		err := "Unknown operation: " + at.String()
		log.Error(err)
	}

}

// Read calculates the new vector clock for a read operation and update cv
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - at *TraceElementAtomic: The trace element
//   - numberOfRoutines int: The number of routines in the trace
func Read(graph *PoGraph, at *trace.ElementAtomic) {
	id := at.ResourceID()

	if graph != nil {
		if graph.lastAtomicWriter[id] != nil {
			graph.AddEdge(at, graph.lastAtomicWriter[id])
		}
	} else {
		if po.lastAtomicWriter[id] != nil {
			AddEdge(at, po.lastAtomicWriter[id], false)
		}
	}
}

func Write(graph *PoGraph, at *trace.ElementAtomic) {
	id := at.ResourceID()

	gr := graph
	if graph == nil {
		gr = &po
	}

	gr.lastAtomicWriter[id] = at
}
