// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the pog for wait group
//
//
// License: BSD-3-Clause

package a_pog

import (
	"gosc/trace"
	"gosc/utils/log"
)

// UpdateHBWait update the pog for a wait group operation
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - wa *trace.TraceElementWait: the wait group operation
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func UpdateHBWait(graph *PoGraph, wa *trace.ElementWait, recorded bool) {
	switch wa.Type(true) {
	case trace.WaitAdd, trace.WaitDone:
		Change(graph, wa)
	case trace.WaitWait:
		Wait(graph, wa, recorded)
	default:
		err := "Unknown operation on wait group: " + wa.String()
		log.Error(err)
	}
}

// Change updates the pog for an add or done operation and update cv
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - wa *TraceElementWait: The trace element
func Change(graph *PoGraph, wa *trace.ElementWait) {
	id := wa.ResourceID()

	gr := graph
	if graph == nil {
		gr = &po
	}

	lw := gr.lastChangeWg[id]
	if lw != nil {
		if graph != nil {
			graph.AddEdge(lw, wa)
		} else {
			AddEdge(lw, wa, false)
		}
	}
	gr.lastChangeWg[id] = wa
}

// Wait updates the pog for a wait operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - wa *TraceElementWait: The trace element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func Wait(graph *PoGraph, wa *trace.ElementWait, recorded bool) {
	if recorded && !wa.Committed() {
		return
	}

	id := wa.ResourceID()
	gr := graph
	if graph == nil {
		gr = &po
	}

	lc := gr.lastChangeWg[id]
	if lc != nil {
		if graph != nil {
			graph.AddEdge(lc, wa)
		} else {
			AddEdge(lc, wa, false)
		}
	}
}
