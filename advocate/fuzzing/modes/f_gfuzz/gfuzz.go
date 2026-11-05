// Copyright (c) 2025 Erik Kassubek
//
// File: gfuzz.go
// Brief: Main function to run gfuzz
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gfuzz

import (
	"advocate/utils/log"
	"math"
)

const maxNumberMutsIfGuided = 4

// CreateMutations creates a new mutations for GFuzz if the previous run was interesting
func CreateMutations(guided bool) {
	// add new mutations based on GFuzz select
	if isInterestingSelect() {
		numberMut := numberMutations()
		if guided {
			numberMut = min(numberMut, maxNumberMutsIfGuided)
		}
		flipProb := getFlipProbability()
		numMutAdd := createMutationsGFuzz(numberMut, flipProb)
		log.Infof("Add %d select mutations to queue", numMutAdd)
	} else {
		log.Info("Add 0 select mutations to queue")
	}
}

// Get the probability that a select changes its preferred case
// It is selected in such a way, that at least one of the selects if flipped
// with a probability of at least 99%.
// Additionally the flip probability is at least 10% for each select.
func getFlipProbability() float64 {
	p := 0.99   // min prob that at least one case is flipped
	pMin := 0.1 // min prob that a select gets flipped

	return max(pMin, 1-math.Pow(1-p, 1/float64(NumberSelects)))
}
