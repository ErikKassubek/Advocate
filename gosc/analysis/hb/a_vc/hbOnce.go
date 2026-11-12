// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the vc for once
//
//
// License: BSD-3-Clause

package a_vc

import (
	"gosc/analysis/a_base"
	"gosc/analysis/hb/a_clock"
	"gosc/trace"
)

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(on *trace.ElementOnce) {
	routine := on.RoutineID()

	on.SetVc(a_clock.Strong, CurrentVC[routine])
	on.SetVc(a_clock.Weak, CurrentVC[routine])

	if on.GetSuc() {
		DoSuc(on)
	} else {
		DoFail(on)
	}
}

// DoSuc updates and calculates the vector clocks given a successful do operation
//
// Parameter:
//   - on *TraceElementOnce: The trace element
func DoSuc(on *trace.ElementOnce) {
	routine := on.RoutineID()

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// DoFail updates and calculates the vector clocks given a unsuccessful do operation
//
// Parameter:
//   - on *TraceElementOnce: The trace element
func DoFail(on *trace.ElementOnce) {
	id := on.ResourceID()
	routine := on.RoutineID()

	suc := a_base.OSuc[id]

	if suc != nil {
		CurrentVC[routine].Sync(suc.VC(a_clock.Strong))
	}
	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}
