// Copyright (c) 2024 Erik Kassubek
//
// File: score.go
// Brief: Functions to compute the score for fuzzing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gfuzz

import (
	"advocate/fuzzing/f_base"
	"advocate/utils/settings"
	"math"
)

// Calculate how many gFuzz mutations should be created for a given
// trace
//
// Returns:
//   - int: the number of mutations
func numberMutations() int {
	score := calculateScore()
	maxScore = math.Max(score, maxScore)

	return int(math.Ceil(5.0 * score / maxScore))
}

// Calculate the score of the given run
func calculateScore() float64 {
	fact1 := settings.GFuzzW1
	fact2 := settings.GFuzzW2
	fact3 := settings.GFuzzW3
	fact4 := settings.GFuzzW4

	res := 0.0

	// number of communications per communication pair (countChOpPair)
	for _, pair := range PairInfoTrace {
		res += math.Log2(float64(pair.Com))
	}

	// number of channels created (createCh)
	res += fact1 * float64(len(ChannelInfoTrace))

	// number of close (closeCh)
	res += fact2 * float64(NumberClose)

	// maximum buffer size for each chan (maxChBufFull)
	bufFullSum := 0.0
	for _, ch := range ChannelInfoFile {
		bufFullSum += float64(ch.MaxQCount)
	}
	res += fact3 * bufFullSum

	if f_base.UseHBInfoFuzzing {
		// number of select cases with possible partner (both executed and not executed)
		res += fact4 * float64(NumberSelectCasesWithPartner)
	}

	return res
}
