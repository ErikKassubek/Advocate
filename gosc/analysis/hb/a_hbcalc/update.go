// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update hb info for the different types
//
//
// License: BSD-3-Clause

package a_hbcalc

import (
	"gosc/analysis/hb/a_cssts"
	"gosc/analysis/hb/a_pog"
	"gosc/analysis/hb/a_vc"
	"gosc/trace"
	"gosc/utils/timer"
)

// UpdateHBAtomic update the vector clocks for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateHBAtomic(at *trace.ElementAtomic) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		a_vc.UpdateHBAtomic(at)
	}

	if CalcPog {
		a_pog.UpdateHBAtomic(nil, at)
	}

	if CalcCssts {
		a_cssts.UpdateHBAtomic(at)
	}
}

// UpdateHBChannel updates the hb info of the trace for a channel operation
//
// Parameter
//   - ch *trace.TraceElementChannel: the channel trace operation
func UpdateHBChannel(ch *trace.ElementChannel) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		a_vc.UpdateHBChannel(ch)
	}

	if CalcPog {
		a_pog.UpdateHBChannel(nil, ch, true)
	}

	if CalcCssts {
		a_cssts.UpdateHBChannel(ch)
	}
}

// UpdateHBSelect updates the hb info of the trace for a select
//
// Parameter
//   - ch *trace.TraceElementChannel: the channel trace operation
func UpdateHBSelect(se *trace.ElementSelect) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		a_vc.UpdateHBSelect(se)
	}

	if CalcPog {
		a_pog.UpdateHBSelect(nil, se, true)
	}

	if CalcCssts {
		a_cssts.UpdateHBSelect(se)
	}
}

// UpdateHBCond updates the hb info of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		a_vc.UpdateHBCond(co)
	}

	if CalcPog {
		a_pog.UpdateHBCond(nil, co)
	}

	if CalcCssts {
		a_cssts.UpdateHBCond(co)
	}
}

// UpdateHBFork updates the hb info of the trace for a fork
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBFork(fo *trace.ElementFork) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	// Note: the update for the pog is done in AddEdgeSameRoutineAndFork

	if CalcVC {
		a_vc.UpdateHBFork(fo)
	}

	if CalcCssts {
		a_cssts.UpdateHBFork(fo)
	}
}

// UpdateHBMutex updates the hb info of the trace for a mutex operation
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
//   - alt bool: if IgnoreCriticalSections is set
func UpdateHBMutex(mu *trace.ElementMutex, alt bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		a_vc.UpdateHBMutex(mu, alt)
	}

	if CalcPog {
		a_pog.UpdateHBMutex(nil, mu, true)
	}

	if CalcCssts {
		a_cssts.UpdateHBMutex(mu)
	}
}

// UpdateHBNew stores the hb info of the trace for a new element
//
// Parameter
//   - n *trace.TraceElementAlloc: the new trace operation
func UpdateHBNew(n *trace.ElementAlloc) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	// For new and routine end elements, we only store the current vc
	// Therefore, the graph based methods do not do anything

	if CalcVC {
		a_vc.UpdateHBNew(n)
	}
}

// UpdateHBOnce updates the hb info of the trace for a once
//
// Parameter
//   - on *trace.TraceElementOnce: the once trace operation
func UpdateHBOnce(on *trace.ElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		a_vc.UpdateHBOnce(on)
	}

	if CalcPog {
		a_pog.UpdateHBOnce(nil, on)
	}

	if CalcCssts {
		a_cssts.UpdateHBOnce(on)
	}
}

// UpdateHBRoutineEnd stores the hb info of the trace for a routine end element
//
// Parameter
//   - n *trace.TraceElementAlloc: the new trace operation
func UpdateHBRoutineEnd(n *trace.ElementRoutineEnd) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	// For new and routine end elements, we only store the current vc
	// Therefore, the graph based methods do not do anything

	if CalcVC {
		a_vc.UpdateHBRoutineEnd(n)
	}
}

// UpdateHBWait updates the hb info of the trace for a fait group
//
// Parameter
//   - wa *trace.TraceElementWait: the wait group trace operation
func UpdateHBWait(wa *trace.ElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		a_vc.UpdateHBWait(wa)
	}

	if CalcPog {
		a_pog.UpdateHBWait(nil, wa, true)
	}

	if CalcCssts {
		a_cssts.UpdateHBWait(wa)
	}
}
