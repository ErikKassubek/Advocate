// Copyright (c) 2024 
//
// File: vcOnce.go
// Brief: Update functions for happens before info for once operations
//
//
// License: BSD-3-Clause

package a_elements

import (
	"gosc/analysis/a_base"
	"gosc/analysis/hb/a_hbcalc"
	"gosc/trace"
)

// AnalyzeOnce update the hb info of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func AnalyzeOnce(on *trace.ElementOnce) {
	a_hbcalc.UpdateHBOnce(on)

	if on.GetSuc() {
		id := on.ResourceID()
		a_base.OSuc[id] = on
	}
}
