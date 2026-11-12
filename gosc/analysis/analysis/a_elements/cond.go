// Copyright (c) 2024 
//
// File: hbCond.go
// Brief: Update functions for happens before info for conditional variables operations
//
//
// License: BSD-3-Clause

package a_elements

import (
	"gosc/analysis/a_base"
	"gosc/analysis/hb/a_hbcalc"
	"gosc/trace"
)

// AnalyzeCond calculates the hb info for cond events and updates relevant
// analysis data
//
// Parameter:
//   - co *trace.ElementCond: the element
func AnalyzeCond(co *trace.ElementCond) {
	a_hbcalc.UpdateHBCond(co)

	// update currently waiting elements
	id := co.ResourceID()
	if co.Committed() { // not leak
		switch co.Type(true) {
		case trace.CondWait:
			if _, ok := a_base.CurrentlyWaiting[id]; !ok {
				a_base.CurrentlyWaiting[id] = make([]*trace.ElementCond, 0)
			}
			a_base.CurrentlyWaiting[id] = append(a_base.CurrentlyWaiting[id], co)
		case trace.CondSignal:
			if len(a_base.CurrentlyWaiting[id]) != 0 {
				a_base.CurrentlyWaiting[id] = a_base.CurrentlyWaiting[id][1:]
			}
		case trace.CondBroadcast:
			a_base.CurrentlyWaiting[id] = make([]*trace.ElementCond, 0)
		}

	}
}
