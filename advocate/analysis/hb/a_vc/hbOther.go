// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Store the vc for new and routine element
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_vc

import (
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
)

// UpdateHBNew store the vector clock of the element
// Parameter:
//   - n *trace.TraceElementAlloc: the new trace element
func UpdateHBNew(n *trace.ElementAlloc) {
	routine := n.RoutineID()
	n.SetVc(a_clock.Strong, CurrentVC[routine])
	n.SetVc(a_clock.Weak, CurrentWVC[routine])
}

// UpdateHBRoutineEnd store the vector clock of the element
// Parameter:
//   - re *trace.TraceElementRoutineEnd: the new trace element
func UpdateHBRoutineEnd(re *trace.ElementRoutineEnd) {
	routine := re.RoutineID()
	re.SetVc(a_clock.Strong, CurrentVC[routine])
	re.SetVc(a_clock.Weak, CurrentWVC[routine])
}
