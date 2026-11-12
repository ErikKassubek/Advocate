// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the vc for wait group
//
//
// License: BSD-3-Clause

package a_vc

import (
	"gosc/analysis/a_base"
	"gosc/analysis/hb/a_clock"
	"gosc/trace"
	"gosc/utils/log"
)

// UpdateHBWait update the vector clocks for a wait group operation
// Parameter:
//   - wa *trace.TraceElementWait: the wait group operation
func UpdateHBWait(wa *trace.ElementWait) {
	routine := wa.RoutineID()
	wa.SetVc(a_clock.Strong, CurrentVC[routine])
	wa.SetVc(a_clock.Weak, CurrentWVC[routine])

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

// Change calculate the new vector clock for a add or done operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Change(wa *trace.ElementWait) {
	id := wa.ResourceID()
	routine := wa.RoutineID()

	lw := a_base.LastChangeWG[id]
	if lw != nil {
		wa.VC(a_clock.Strong).Sync(lw.VC(a_clock.Strong))
	}
	a_base.LastChangeWG[id] = wa

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// Wait calculates the new vector clock for a wait operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Wait(wa *trace.ElementWait) {
	id := wa.ResourceID()
	routine := wa.RoutineID()

	if wa.Committed() {
		lc := a_base.LastChangeWG[id]
		if lc != nil {
			CurrentVC[routine].Sync(lc.VC(a_clock.Strong))
		}
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}
