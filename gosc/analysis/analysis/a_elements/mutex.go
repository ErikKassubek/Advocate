// gosc/analysis/analysis/elements/mutex.go

// Copyright (c) 2024 
//
// File: hbMutex.go
// Brief: Update functions for happens before info for mutex operation
//        Some of the functions start analysis functions
//
//
// License: BSD-3-Clause

package a_elements

import (
	"gosc/analysis/a_base"
	"gosc/analysis/analysis/a_scenarios"
	"gosc/analysis/hb/a_hbcalc"
	"gosc/analysis/hb/a_vc"
	"gosc/trace"
	"gosc/utils/flags"
	"gosc/utils/log"
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

		if a_base.AnalysisCasesMap[flags.UnlockBeforeLock] {
			a_scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	// --------- READ LOCK (RWMutex RLock) ---------
	case trace.MutexRLock:
		a_base.CurrentlyHoldLock[id] = mu
		a_scenarios.IncFuzzingCounter(mu)

		if a_base.AnalysisCasesMap[flags.UnlockBeforeLock] {
			a_scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	// --------- TRY LOCK (write) ---------
	case trace.MutexTryLock:
		if mu.IsSuc() {
			a_base.CurrentlyHoldLock[id] = mu
			a_scenarios.IncFuzzingCounter(mu)

			if a_base.AnalysisCasesMap[flags.UnlockBeforeLock] {
				a_scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}

	// --------- TRY RLOCK (read) ---------
	case trace.MutexTryRLock:
		if mu.IsSuc() {
			a_base.CurrentlyHoldLock[id] = mu
			a_scenarios.IncFuzzingCounter(mu)

			if a_base.AnalysisCasesMap[flags.UnlockBeforeLock] {
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

		if a_base.AnalysisCasesMap[flags.UnlockBeforeLock] {
			a_scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}

	// --------- RUNLOCK (read) ---------
	case trace.MutexRUnlock:
		a_base.RelR[id] = &a_base.ElemWithVc{
			Elem: mu,
			Vc:   a_vc.CurrentVC[routine].Copy(),
		}

		a_base.CurrentlyHoldLock[id] = nil

		if a_base.AnalysisCasesMap[flags.UnlockBeforeLock] {
			a_scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}

	default:
		log.Error("Unknown mutex operation: " + mu.String())
	}
}
