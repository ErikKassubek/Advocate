// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for fork
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_vc

import (
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
)

// UpdateHBFork update and calculate happens before information for fork operations
// It only calculates the VC and csst, not the pog, which is included in the
// edge creation of elements in the same routine
//
// Parameter:
//   - fo *TraceElementFork: the fork element
func UpdateHBFork(fo *trace.ElementFork) {
	routine := fo.RoutineID()

	fo.SetVc(a_clock.Strong, CurrentVC[routine])
	fo.SetVc(a_clock.Weak, CurrentWVC[routine])

	oldRout := fo.RoutineID()
	newRout := fo.ResourceID()

	CurrentVC[newRout] = CurrentVC[oldRout].Copy()
	CurrentVC[oldRout].Inc(oldRout)
	CurrentVC[newRout].Inc(newRout)

	CurrentWVC[newRout] = CurrentWVC[oldRout].Copy()
	CurrentWVC[oldRout].Inc(oldRout)
	CurrentWVC[newRout].Inc(newRout)
}
