// Copyright (c) 2025 Erik Kassubek
//
// File: getHb.go
// Brief: Get the HB info between two elements
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_hbcalc

import (
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/analysis/hb/a_cssts"
	"advocate/analysis/hb/a_pog"
	"advocate/analysis/hb/a_vc"
	"advocate/trace"
)

// GetHappensBefore returns the happens before relation between two operations
//
// Parameter:
//   - t1 trace.Element: the trace element
//   - t2 trace.Element: the second element
//   - vcType a_clock.VcType: weak or strong
//
// Returns:
//   - happensBefore: The happens before relation between the elements
func GetHappensBefore(e1, e2 trace.Element, vcType a_clock.VcType) a_hb.HappensBefore {
	if CalcVC {
		return a_vc.GetHappensBefore(e1, e2, vcType)
	}

	if CalcPog {
		return a_pog.GetHappensBefore(e1, e2, vcType)
	}

	if CalcCssts {
		return a_cssts.GetHappensBefore(e1, e2, vcType)
	}

	return a_hb.None
}
