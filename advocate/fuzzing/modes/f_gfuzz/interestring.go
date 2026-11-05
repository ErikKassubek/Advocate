// Copyright (c) 2024 Erik Kassubek
//
// File: interestring.go
// Brief: Functions to determine whether a run was interesting
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gfuzz

import (
	"advocate/fuzzing/f_base"
	"advocate/utils/types"
	"math"
)

// A run is considered interesting, if at least one of the following conditions is met
//  1. The run contains a new pair of channel operations (new meaning it has not been seen in any of the previous runs)
//  2. An operation pair's execution counter changes significantly (by at least 50%) from previous avg over all runs.
//  3. A new channel operation is triggered, such as creating, closing or not closing a channel for the first time
//  4. A buffered channel gets a larger maximum fullness than in all previous executions (MaxChBufFull)
//  5. A select case is executed for the first time
func isInterestingSelect() bool {
	// 1. The run contains a new pair of channel operations (new meaning it has not been seen in any of the previous runs)
	for keyTrace, pit := range PairInfoTrace {
		pif, ok := PairInfoFile[keyTrace]
		if !ok {
			return true
		}

		// 2. An operation pair's execution counter changes significantly from previous order.
		change := math.Abs((pit.Com - pif.Com) / pif.Com)
		if change > 0.5 {
			return true
		}
	}

	for _, cit := range ChannelInfoTrace {
		fileData, ok := ChannelInfoFile[cit.GlobalID]

		// 3. A new channel operation is triggered, such as creating, closing or not closing a channel for the first time
		// never created before
		if !ok {
			return true
		}
		// first time closed
		if cit.CloseInfo == Always && fileData.CloseInfo == Never {
			return true
		}
		// first time not closed
		if cit.CloseInfo == Never && fileData.CloseInfo == Always {
			return true
		}

		// 4. A buffered channel gets a larger maximum fullness than in all previous executions (MaxChBufFull)
		if cit.MaxQCount > fileData.MaxQCount {
			return true
		}
	}

	if f_base.UseHBInfoFuzzing {
		// 5. A select choses a case it has never been selected before
		for id, sit := range SelectInfoTrace {
			alreadyExecCase, ok := SelectInfoFile[id]
			if !ok { // select has never been seen before
				return true
			}

			for _, sel := range sit { // case has been executed for the first time
				if !types.Contains(alreadyExecCase, sel.ChosenCase) {
					return true
				}
			}
		}
	}

	return false
}
