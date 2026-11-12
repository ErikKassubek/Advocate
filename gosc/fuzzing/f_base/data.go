// Copyright (c) 2024 
//
// File: data.go
// Brief: File to define and contain the fuzzing data
//
//
// License: BSD-3-Clause

package f_base

import (
	"gosc/fuzzing/f_constraints"
	"gosc/utils/results/results"
	"gosc/utils/types"

	"time"
)

const factorCaseWithPartner = 3

// General data for fuzzing
var (
	MaxNumberRuns     = 100
	MaxTime           = 7 * time.Minute
	MaxTimeSet        = false
	NumberFuzzingRuns = 0
	MutationQueue     = types.NewQueue[Mutation]()

	// count how often a specific mutation has been in the queue
	AllMutations          = make(map[string]int)
	FuzzingModeGFuzz      = false
	FuzzingModeGoPie      = false
	FuzzingModeGoCRHBPlus = false
	FuzzingModeFlow       = false
	FuzzingModeGuided     = false
	FuzzingHbAnalysis     = true

	NumberOfPreviousRuns = 0

	UseHBInfoFuzzing = true

	FinishIfBugFound = false

	NumberWrittenMutations = 0
	NumberWrittenRewrites  = 0
	NumberWrittenFuzzing   = 0
	MaxRewriteExec         = -1

	// for each mutation file, store the file number and the chain
	ChainFiles = make(map[int]f_constraints.Constraint)

	TotalRuns = 0
	Equiv     = 0
)

func Clear() {
	MutationQueue = types.NewQueue[Mutation]()
	AllMutations = make(map[string]int)
	ChainFiles = make(map[int]f_constraints.Constraint)
}

// ClearDataFull resets the fuzzing data that is unique for each test but used for each fuzzing
// run of a test
func ClearDataFull() {
	results.Reset()

	NumberOfPreviousRuns = 0

	NumberFuzzingRuns = 0
	NumberWrittenFuzzing = 0
	NumberWrittenMutations = 0
	NumberWrittenRewrites = 0
}
