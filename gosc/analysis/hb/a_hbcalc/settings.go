// Copyright (c) 2025 
//
// File: settings.go
// Brief: Settings for hb calculation
//
//
// License: BSD-3-Clause

package a_hbcalc

// Settings for which hb structures should be calculated
var (
	CalcVC    = false
	CalcPog   = false
	CalcCssts = false
)

// SetHbSettings sets which hb structure should be calculated
//
// Parameter:
//   - vc bool: calculate vector clocks
//   - pog bool: calculate partial order graph
//   - cssts bool: calculate Collective Sparse Segment Trees
func SetHbSettings(vc, pog, cssts bool) {
	CalcVC = vc
	CalcPog = pog
	CalcCssts = cssts
}
