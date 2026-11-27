// advocate/analysis/analysis/elements/mutex.go

// Copyright (c) 2024 Erik Kassubek
//
// File: hbMutex.go
// Brief: Update functions for happens before info for mutex operation
//        Some of the functions start analysis functions
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"advocate/analysis/a_base"
	"advocate/analysis/analysis/a_scenarios"
	"advocate/analysis/hb/a_hbcalc"
	"advocate/analysis/hb/a_vc"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
)

// UpdateMutex store and update the vector clock of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
//   - alt bool: if IgnoreCriticalSections is set
func UpdateMutex(mu *trace.ElementMutex, alt bool) {
	a_hbcalc.UpdateHBMutex(mu, alt)

	routine := mu.RoutineID()
	id := mu.ResourceID()

	switch mu.Type(true) {

	// --------- WRITE LOCK ---------
	case trace.MutexLock:
		a_base.CurrentlyHoldLock[id] = mu
		a_scenarios.IncFuzzingCounter(mu)

		if flags.AnalysisCasesMap[flags.UnlockBeforeLock] {
			a_scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	// --------- READ LOCK (RWMutex RLock) ---------
	case trace.MutexRLock:
		a_base.CurrentlyHoldLock[id] = mu
		a_scenarios.IncFuzzingCounter(mu)

		if flags.AnalysisCasesMap[flags.UnlockBeforeLock] {
			a_scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	// --------- TRY LOCK (write) ---------
	case trace.MutexTryLock:
		if mu.IsSuc() {
			a_base.CurrentlyHoldLock[id] = mu
			a_scenarios.IncFuzzingCounter(mu)

			if flags.AnalysisCasesMap[flags.UnlockBeforeLock] {
				a_scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}

	// --------- TRY RLOCK (read) ---------
	case trace.MutexTryRLock:
		if mu.IsSuc() {
			a_base.CurrentlyHoldLock[id] = mu
			a_scenarios.IncFuzzingCounter(mu)

			if flags.AnalysisCasesMap[flags.UnlockBeforeLock] {
				a_scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}

	// --------- UNLOCK (write) ---------
	case trace.MutexUnlock:
		a_base.RelW[id] = &a_base.ElemWithVc{
			Elem: mu,
			Vc:   a_vc.CurrentVC[routine].Copy(),
		}

		a_base.RelR[id] = &a_base.ElemWithVc{
			Elem: mu,
			Vc:   a_vc.CurrentVC[routine].Copy(),
		}

		a_base.CurrentlyHoldLock[id] = nil

		if flags.AnalysisCasesMap[flags.UnlockBeforeLock] {
			a_scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}

	// --------- RUNLOCK (read) ---------
	case trace.MutexRUnlock:
		a_base.RelR[id] = &a_base.ElemWithVc{
			Elem: mu,
			Vc:   a_vc.CurrentVC[routine].Copy(),
		}

		a_base.CurrentlyHoldLock[id] = nil

		if flags.AnalysisCasesMap[flags.UnlockBeforeLock] {
			a_scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}

	default:
		log.Error("Unknown mutex operation: " + mu.String())
	}
}
