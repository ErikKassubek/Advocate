// Copyright (c) 2025 
//
// File: atomic.go
// Brief: Update the data for an atomic element
//
//
// License: BSD-3-Clause

package a_elements

import (
	"gosc/analysis/hb/a_hbcalc"
	"gosc/trace"
)

// AnalyzeAtomic update the hb info for an atomic event
//
// Parameter:
//   - at *trace.ElementAtomic: the element
func AnalyzeAtomic(at *trace.ElementAtomic) {
	a_hbcalc.UpdateHBAtomic(at)
}
