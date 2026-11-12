// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the cssts for wait group
//
//
// License: BSD-3-Clause

package a_cssts

import (
	"gosc/analysis/a_base"
	"gosc/trace"
	"gosc/utils/log"
)

// UpdateHBWait update the cssts for a wait group operation
// Parameter:
//   - wa *trace.TraceElementWait: the wait group operation
func UpdateHBWait(wa *trace.ElementWait) {
	switch wa.Type(true) {
	case trace.WaitAdd, trace.WaitDone:
		Change(wa)
	case trace.WaitWait:
		Wait(wa)
	default:
		err := "Unknown operation on wait group: " + wa.String()
		log.Error(err)
	}
}

// Change updates the cssts for an add or done operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Change(wa *trace.ElementWait) {
	id := wa.ResourceID()

	lw := a_base.LastChangeWG[id]
	if lw != nil {
		AddEdge(lw, wa, false)
	}
	a_base.LastChangeWG[id] = wa
}

// Wait updates the pog for a wait operation
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Wait(wa *trace.ElementWait) {
	id := wa.ResourceID()

	if wa.Committed() {
		lc := a_base.LastChangeWG[id]
		if lc != nil {
			AddEdge(lc, wa, false)
		}
	}
}
