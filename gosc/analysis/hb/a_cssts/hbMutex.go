// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the cssts for mutex operations
//
//
// License: BSD-3-Clause

package a_cssts

import (
	"gosc/analysis/a_base"
	"gosc/analysis/hb/a_clock"
	"gosc/trace"
	"gosc/utils/log"
)

// UpdateHBMutex updates the cssts of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateHBMutex(mu *trace.ElementMutex) {
	switch mu.Type(true) {
	case trace.MutexLock:
		Lock(mu)
	case trace.MutexRLock:
		RLock(mu)
	case trace.MutexTryLock:
		if mu.IsSuc() {
			Lock(mu)
		}
	case trace.MutexTryRLock:
		if mu.IsSuc() {
			RLock(mu)
		}
	case trace.MutexUnlock:
	case trace.MutexRUnlock:
		RUnlock(mu)
	default:
		err := "Unknown mutex operation: " + mu.String()
		log.Error(err)
	}
}

// Lock updates the cssts given a lock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func Lock(mu *trace.ElementMutex) {
	id := mu.ResourceID()

	if !mu.Committed() {
		return
	}

	if e, ok := a_base.RelW[id]; ok {
		AddEdge(e.Elem, mu, false)
	}
	if e, ok := a_base.RelR[id]; ok {
		AddEdge(e.Elem, mu, false)
	}
}

// RLock updates the cssts given a rlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//
// Returns:
//   - *VectorClock: The new vector clock
func RLock(mu *trace.ElementMutex) {
	id := mu.ResourceID()

	if !mu.Committed() {
		return
	}

	if e, ok := a_base.RelW[id]; ok {
		AddEdge(e.Elem, mu, false)
	}
}

// RUnlock updates the cssts given a runlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func RUnlock(mu *trace.ElementMutex) {
	id := mu.ResourceID()

	if !mu.Committed() {
		return
	}

	if _, ok := a_base.RelR[id]; !ok {
		a_base.RelR[id] = &a_base.ElemWithVc{
			Vc:   a_clock.NewVectorClock(a_base.GetNoRoutines()),
			Elem: nil,
		}
	} else {
		AddEdge(mu, a_base.RelR[id].Elem, false)
	}
}
