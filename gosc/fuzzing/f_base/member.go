// Copyright (c) 2025 
//
// File: member.go
// Brief: Member function on the types in fuzzing/data/types.go
//
//
// License: BSD-3-Clause

package f_base

import "math/rand"

// GetCopyRandom returns a copy of fs with a randomly selected case id.
//
// Parameter:
//   - def bool: if true, default is a possible value, if false it is not
//   - flipChange bool: probability that a select case is chosen randomly. Otherwise the chosen case is kept
//
// Returns:
//   - int: the chosen case ID
func (this FuzzingSelect) GetCopyRandom(def bool, flipChance float64) FuzzingSelect {
	// do only flip with certain chance
	if rand.Float64() > flipChance {
		return FuzzingSelect{ID: this.ID, T: this.T, ChosenCase: this.ChosenCase, NumberCases: this.NumberCases, ContainsDefault: this.ContainsDefault}
	}

	// if at most one case and no default (should not happen), or only default select the same case again
	if (!def && this.NumberCases <= 1) || (def && this.NumberCases == 0) {
		return FuzzingSelect{ID: this.ID, T: this.T, ChosenCase: this.ChosenCase, NumberCases: this.NumberCases, ContainsDefault: this.ContainsDefault}
	}

	prefCase := this.chooseRandomCase(def)

	return FuzzingSelect{ID: this.ID, T: this.T, ChosenCase: prefCase, NumberCases: this.NumberCases, ContainsDefault: this.ContainsDefault}
}

// Randomly select a case.
// The case is between 0 and fs.numberCases if def is false and between -1 and fs.numberCases otherwise
// fs.chosenCase is never chosen
// The values in fs.casiWithPos have a higher likelihood to be chosen by a factor factorCaseWithPartner (defined in fuzzing/data.go)
//
// Parameter
//   - def bool: true if the select contains a bool
//
// Returns:
//   - the chosen case id
func (this FuzzingSelect) chooseRandomCase(def bool) int {
	// Determine the starting number based on includeZero
	start := 0
	if def {
		start = -1
	}

	// Create a weight map for the probabilities
	weights := make(map[int]int)

	// Assign weights to each number
	for i := start; i < this.NumberCases; i++ {
		weights[i] = 1 // Default weight
	}

	// Increase weights for numbers in fs.casiWithPos
	for _, num := range this.CasiWithPos {
		if num >= start && num < this.NumberCases && num != this.ChosenCase {
			weights[num] *= factorCaseWithPartner
		}
	}

	// Generate a cumulative weight array
	cumulativeWeights := []int{}
	numbers := []int{} // Keep track of the corresponding numbers
	totalWeight := 0

	for i := start; i < this.NumberCases; i++ {
		if weight, exists := weights[i]; exists && weight > 0 {
			totalWeight += weight
			cumulativeWeights = append(cumulativeWeights, totalWeight)
			numbers = append(numbers, i)
		}
	}

	// Handle edge case where no valid number can be chosen
	if totalWeight == 0 {
		return 0
	}

	r := rand.Intn(totalWeight)

	// Find the number corresponding to the random weight
	for i, cw := range cumulativeWeights {
		if r < cw {
			return numbers[i]
		}
	}

	// Fallback (should never reach here)
	return 0
}
