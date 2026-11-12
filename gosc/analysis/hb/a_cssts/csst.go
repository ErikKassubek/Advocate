// Copyright (c) 2025 
//
// File: csst.go
// Brief: functions to use the csst
//
//
// License: BSD-3-Clause

package a_cssts

import (
	"gosc/analysis/a_base"
	"gosc/trace"
	"gosc/utils/types"
)

// Data for the csst
var (
	Csst         IncrementalCSST
	CsstInverted IncrementalCSST

	CsstWeak         IncrementalCSST
	CsstWeakInverted IncrementalCSST
)

// InitCSSTs initializes the cssts
//
// Parameter:
//   - lengths []int: the number of elements per routine
func InitCSSTs(lengths []int) {
	chanBuffer = make(map[int]([]a_base.BufferedVC))
	chanBufferSize = make(map[int]int)

	Csst = NewIncrementalCSST(lengths)
	CsstInverted = NewIncrementalCSST(lengths)

	CsstWeak = NewIncrementalCSST(lengths)
	CsstWeakInverted = NewIncrementalCSST(lengths)
}

// For a trace element, return the routine id and elem rout index used as identifier
// in the CSST
//
// Parameters:
//   - elem trace.Element: the element to find the index for
//
// Returns:
//   - types.Pair[int, int]: routine id of elem, routine local index of elem
func getIndicesFromTraceElem(elem trace.Element) types.Pair[int, int] {
	rout, index := elem.TraceIndex()
	return types.NewPair(rout, index)
}
