// Copyright (c) 2025 Erik Kassubek
//
// File: data.go
// Brief: Data for GoPie
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gopie

import (
	"advocate/fuzzing/f_constraints"
	"advocate/trace"
	"advocate/utils/settings"
	"advocate/utils/types"
)

// Data for goPie fuzzing
var (
	// store all created mutations to avoid doubling
	allGoPieMutations = make(map[string]struct{})

	maxGoPieScore = 0

	// number of different starting points for chains in GoPie (in the original: cfg.MaxWorker)
	maxSCStart = settings.GoPieSCStart

	SchedulingChains []f_constraints.Constraint
	CurrentChain     f_constraints.Constraint
	LastRoutine      = -1

	// GoPie relations
	counterCPOP1 = 0
	counterCPOP2 = 0
	rel1         = make(map[trace.Element]map[trace.Element]struct{})
	rel2         = make(map[trace.Element]map[trace.Element]struct{})

	ElemsByID = make(map[int][]trace.Element) // id -> chan/sel/mutex elem

	usedStartPos = make([]types.Pair[trace.Element, trace.Element], 0)

	NumberInvalidMuts = 0
	NumberDoubleMuts  = 0
	NumberTotalMuts   = 0
)

// ClearDataRun the data for one run in GoPie
func ClearDataRun() {
	rel1 = make(map[trace.Element]map[trace.Element]struct{})
	rel2 = make(map[trace.Element]map[trace.Element]struct{})
	counterCPOP1 = 0
	counterCPOP2 = 0
	ElemsByID = make(map[int][]trace.Element)
}

// ClearData deletes all the GoPie data
func ClearData() {
	ClearDataRun()
	maxGoPieScore = 0
	usedStartPos = make([]types.Pair[trace.Element, trace.Element], 0)

	NumberInvalidMuts = 0
	NumberTotalMuts = 0
	NumberDoubleMuts = 0
}
