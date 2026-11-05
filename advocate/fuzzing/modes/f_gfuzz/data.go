// Copyright (c) 2025 Erik Kassubek
//
// File: data.go
// Brief: Data for gFuzz
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gfuzz

import (
	"advocate/fuzzing/f_base"
	"advocate/utils/types"
)

const (
	maxRunPerMut = 2
)

// Data for gFuzz
var (
	maxScore = 0.0

	// Info for the current trace
	ChannelInfoTrace = make(map[int]FuzzingChannel)            // localID -> fuzzingChannel
	PairInfoTrace    = make(map[string]FuzzingPair)            // posSend-posRecv -> fuzzing pair
	SelectInfoTrace  = make(map[string][]f_base.FuzzingSelect) // id -> []fuzzingSelects
	NumberSelects    = 0
	NumberClose      = 0

	ChannelInfoFile              = make(map[string]FuzzingChannel) // globalID -> fuzzingChannel
	PairInfoFile                 = make(map[string]FuzzingPair)    // posSend -> fuzzing pair
	SelectInfoFile               = make(map[string][]int)          // globalID -> executed casi
	NumberSelectCasesWithPartner = 0
)

// ClearDataFull deletes all the data of gFuzz
func ClearDataFull() {
	maxScore = 0.0
	ClearDataRun()
}

// ClearDataRun deletes all the data of gFuzz required for one run
func ClearDataRun() {
	ChannelInfoTrace = make(map[int]FuzzingChannel)
	PairInfoTrace = make(map[string]FuzzingPair)
	SelectInfoTrace = make(map[string][]f_base.FuzzingSelect)

	NumberSelects = 0
	NumberClose = 0

	// Info from the file/the previous runs
	ChannelInfoFile = make(map[string]FuzzingChannel) // globalID -> fuzzingChannel
	PairInfoFile = make(map[string]FuzzingPair)       // posSend -> fuzzing pair
	SelectInfoFile = make(map[string][]int)           // globalID -> executed casi
	NumberSelectCasesWithPartner = 0
}

// Merge the close information for a channel from a trace into the internal
//
// Parameter:
//   - trace closeInfo: info from the last recorded run
//   - file closeInfo: stored close info
//
// Returns:
//   - closeInfo: the new close info for the channel
func mergeCloseInfo(trace closeInfo, file closeInfo) closeInfo {
	if trace != file {
		return Sometimes
	}
	return file
}

// MergeTraceInfoIntoFileInfo merges the close info from the last run into the
// internal close info for all ever executed channel close for each element
func MergeTraceInfoIntoFileInfo() {
	// channel info
	for _, cit := range ChannelInfoTrace {
		if cif, ok := ChannelInfoFile[cit.GlobalID]; !ok {
			ChannelInfoFile[cit.GlobalID] = cit
		} else {
			ChannelInfoFile[cit.GlobalID] = FuzzingChannel{cit.GlobalID, 0,
				mergeCloseInfo(cif.CloseInfo, cit.CloseInfo),
				cit.QSize, max(cif.MaxQCount, cit.MaxQCount)}
		}
	}

	// pair info
	for id, pit := range PairInfoTrace {
		if pif, ok := PairInfoFile[id]; !ok {
			PairInfoFile[id] = pit
		} else {
			npr := float64(f_base.NumberOfPreviousRuns)
			pif.Com = (npr*pif.Com + pit.Com) / (npr + 1)
			PairInfoFile[id] = pif
		}
	}

	// select info
	for id, sits := range SelectInfoTrace {
		if _, ok := SelectInfoFile[id]; !ok {
			SelectInfoFile[id] = make([]int, 0)
		}

		for _, sit := range sits {
			SelectInfoFile[id] = types.AddIfNotContains(SelectInfoFile[id], sit.ChosenCase)
		}
	}
}
