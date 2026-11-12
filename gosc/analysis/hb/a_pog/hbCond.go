// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the cssts for conds
//
//
// License: BSD-3-Clause

package a_pog

import (
	"gosc/trace"
	"gosc/utils/types"
)

// UpdateHBCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(graph *PoGraph, co *trace.ElementCond) {
	gr := graph
	if graph == nil {
		gr = &po
	}

	objId := co.ResourceID()

	switch co.Type(true) {
	case trace.CondWait:
		if _, ok := gr.curWaitingCond[objId]; !ok {
			gr.curWaitingCond[objId] = types.NewQueue[*trace.ElementCond]()
		}
		gr.curWaitingCond[objId].Push(co)
	case trace.CondSignal:
		CondSignal(graph, co)
	case trace.CondBroadcast:
		CondBroadcast(graph, co)
	}
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - co *TraceElementCond: The trace element
func CondSignal(graph *PoGraph, co *trace.ElementCond) {
	id := co.ResourceID()

	gr := graph
	if graph == nil {
		gr = &po
	}

	if !gr.curWaitingCond[id].IsEmpty() {
		tWait := gr.curWaitingCond[id].Pop()
		if graph != nil {
			graph.AddEdge(co, tWait)
		} else {
			AddEdge(co, tWait, false)
		}
	}
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - co *TraceElementCond: The trace element
func CondBroadcast(graph *PoGraph, co *trace.ElementCond) {
	id := co.ResourceID()

	gr := graph
	if graph == nil {
		gr = &po
	}

	for !gr.curWaitingCond[id].IsEmpty() {
		wait := gr.curWaitingCond[id].Pop()

		if graph != nil {
			graph.AddEdge(co, wait)
		} else {
			AddEdge(co, wait, false)
		}
	}
}
