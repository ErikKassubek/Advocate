// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the cssts for conds
//
//
// License: BSD-3-Clause

package a_cssts

import (
	"gosc/analysis/a_base"
	"gosc/trace"
)

// UpdateHBCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	switch co.Type(true) {
	case trace.CondWait:
		// wait does not add any edge
	case trace.CondSignal:
		CondSignal(co)
	case trace.CondBroadcast:
		CondBroadcast(co)
	}
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondSignal(co *trace.ElementCond) {
	id := co.ResourceID()

	if len(a_base.CurrentlyWaiting[id]) != 0 {
		tWait := a_base.CurrentlyWaiting[id][0]
		AddEdge(co, tWait, false)
	}
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondBroadcast(co *trace.ElementCond) {
	id := co.ResourceID()
	for _, wait := range a_base.CurrentlyWaiting[id] {
		AddEdge(co, wait, false)
	}
}
