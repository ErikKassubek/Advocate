// Copyright (c) 2025 Erik Kassubek
//
// File: hb.go
// Brief: Get the happens before info
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_vc

import (
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
)

// GetHappensBefore returns the happens before relation between two operations given there
// vc
//
// Parameter:
//   - t1 trace.Element: the trace element
//   - t2 trace.Element: the second element
//   - vcType a_clock.VcType: weak or strong
//
// Returns:
//   - happensBefore: The happens before relation between the elements
func GetHappensBefore(t1, t2 trace.Element, vcType a_clock.VcType) a_hb.HappensBefore {
	return a_clock.GetHappensBefore(t1.VC(vcType), t2.VC(vcType))
}
