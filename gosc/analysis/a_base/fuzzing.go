// gosc/analysis/baseA/fuzzing.go

// Copyright (c) 2024 
//
// File: analysisConcurrentCommunication.go
// Brief: Data collected during analysis for use in fuzzing
//
//
// License: BSD-3-Clause

package a_base

// for fuzzing flow
var (
	FuzzingFlowOnce  = make([]ConcurrentEntry, 0)
	FuzzingFlowMutex = make([]ConcurrentEntry, 0)
	FuzzingFlowSend  = make([]ConcurrentEntry, 0)
	FuzzingFlowRecv  = make([]ConcurrentEntry, 0)

	FuzzingCounter = make(map[int]map[string]int) // id -> pos -> counter

	T1 = false
)
