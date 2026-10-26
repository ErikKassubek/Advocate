// advocate/analysis/analysis/analysis.go

// Copyright (c) 2024 Erik Kassubek
//
// File: analysis.go
// Brief: analysis of traces if performed from here
//
// Author: Erik Kassubek, Sebastian Pohsner
//
// License: BSD-3-Clause

package a_analysis

import (
	"advocate/analysis/a_base"
	"advocate/analysis/analysis/a_elements"
	"advocate/analysis/analysis/a_scenarios"
	"advocate/analysis/hb/a_cssts"
	"advocate/analysis/hb/a_hbcalc"
	hb "advocate/analysis/hb/a_hbcalc"
	"advocate/analysis/hb/a_pog"
	"advocate/analysis/hb/a_vc"
	"advocate/fuzzing/f_base"
	"advocate/trace"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// RunAnalysis starts the analysis of the main trace
//
// Parameter:
//   - fuzzing bool: true if run with fuzzing
func RunAnalysis(fuzzing bool) {
	// catch panics in analysis.
	// Prevents the whole toolchain to panic if one analysis panics
	if log.IsPanicPrevent() {
		defer func() {
			if r := recover(); r != nil {
				control.Cancel()
				log.Error(r)
			}
		}()
	}

	a_base.AnalysisFuzzingFlow = fuzzing

	timer.Start(timer.Analysis)
	defer timer.Stop(timer.Analysis)

	a_scenarios.RunAnalysisOnExitCodes(flags.OnlyAPanicAndLeak)

	if !fuzzing || f_base.UseHBInfoFuzzing {
		RunHBAnalysis(fuzzing)
	}

	if a_base.AnalysisCasesMap[flags.Leak] || flags.OnlyAPanicAndLeak {
		err := a_scenarios.Blocked()
		if err != nil {
			log.Error("Failed to read block info: ", err.Error())
		}
	}

}

// RunHBAnalysis runs the full analysis happens before based analysis
//
// Parameter:
//   - fuzzing bool: true if run with fuzzing
//   - runAna bool: true to run the predictive analysis
//
// Returns:
//   - bool: true if something has been found
func RunHBAnalysis(fuzzing bool) {
	log.Info("Start Analysis")

	a_base.ModeIsFuzzing = fuzzing

	// set which hb structures should be calculated
	// NOTE: Do not use predictive analysis if the first parameter is false
	a_hbcalc.SetHbSettings(true, false, false)
	if flags.OnlyAPanicAndLeak || !a_hbcalc.CalcVC {
		for key := range a_base.AnalysisCasesMap {
			a_base.AnalysisCasesMap[key] = false
		}
	}

	if hb.CalcVC {
		a_vc.InitVC()
	}

	if hb.CalcPog {
		a_pog.InitPOG()
	}

	if hb.CalcCssts {
		a_cssts.InitCSSTs(a_base.GetTraceLengths())
	}

	if a_base.AnalysisCasesMap[flags.ResourceDeadlock] {
		a_scenarios.ResetState()
	}

	if a_base.AnalysisCasesMap[flags.MixedDeadlock] {
		a_scenarios.ResetMixedDeadlockState()
	}

	if hb.CalcVC {
		a_vc.CurrentVC[1].Inc(1)
		a_vc.CurrentWVC[1].Inc(1)
	}

	traceIter := a_base.MainTrace.AsIterator()
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {

		// not enough memory
		if control.IsCanceledRAM.Load() {
			return
		}

		// add edge between element of same routine to partial order trace
		if hb.CalcPog {
			a_pog.AddEdgeSameRoutineAndFork(nil, elem)
		}

		// count how many operations where executed on the underlying structure
		// do not count for operations that do not have an underlying structure
		switch e := elem.(type) {
		case *trace.ElementFork, *trace.ElementAlloc, *trace.ElementRoutineEnd:
		default:
			a_base.AddOpsPerID(e.ResourceID())
		}

		switch e := elem.(type) {
		case *trace.ElementAtomic:
			a_elements.AnalyzeAtomic(e)
		case *trace.ElementChannel:
			a_elements.UpdateChannel(e)
		case *trace.ElementMutex:
			if flags.IgnoreCriticalSection {
				a_elements.UpdateMutex(e, true)
			} else {
				a_elements.UpdateMutex(e, false)
			}
			if a_base.AnalysisFuzzingFlow {
				a_scenarios.GetConcurrentMutexForFuzzing(e)
			}
		case *trace.ElementFork:
			a_elements.AnalyzeFork(e)
		case *trace.ElementSelect:
			cases := e.GetCases()
			ids := make([]int, 0)
			opTypes := make([]int, 0)
			for _, c := range cases {
				switch c.Type(true) {
				case trace.ChannelSend:
					ids = append(ids, c.ResourceID())
					opTypes = append(opTypes, 0)
				case trace.ChannelRecv:
					ids = append(ids, c.ResourceID())
					opTypes = append(opTypes, 1)
				}
			}
			a_elements.UpdateSelect(e)
		case *trace.ElementWait:
			a_elements.AnalyzeWait(e)
		case *trace.ElementCond:
			a_elements.AnalyzeCond(e)
		case *trace.ElementOnce:
			a_elements.AnalyzeOnce(e)
			if a_base.AnalysisFuzzingFlow {
				a_scenarios.GetConcurrentOnceForFuzzing(e)
			}
		case *trace.ElementRoutineEnd:
			a_elements.AnalyzeRoutineEnd(e)
		case *trace.ElementAlloc:
			a_elements.AnalyzeNew(e)
		}

		if a_base.AnalysisCasesMap[flags.ResourceDeadlock] {
			switch e := elem.(type) {
			case *trace.ElementMutex:
				a_scenarios.HandleMutexEventForRessourceDeadlock(*e)
			}
		}

		if a_base.AnalysisCasesMap[flags.MixedDeadlock] {
			switch e := elem.(type) {
			case *trace.ElementMutex:
				a_scenarios.HandleMutexEventForMixedDeadlock(e)
			case *trace.ElementChannel:
				a_scenarios.HandleChannelEventForMixedDeadlock(e)
			}
		}

		if control.WasCanceled() {
			return
		}
	}

	a_base.MainTrace.SetHBWasCalc(hb.CalcVC)

	log.Info("Finished HB analysis")

	if f_base.FuzzingModeGFuzz {
		a_scenarios.RerunCheckForSelectCaseWithPartnerChannel()
		a_scenarios.CheckForSelectCaseWithPartner()
	}

	if control.WasCanceled() {
		return
	}

	if a_base.AnalysisCasesMap[flags.DoneBeforeAdd] {
		a_scenarios.CheckForDoneBeforeAdd()
	}

	if control.WasCanceled() {
		return
	}

	if a_base.AnalysisCasesMap[flags.ResourceDeadlock] {
		a_scenarios.CheckForResourceDeadlock()
	}

	if control.WasCanceled() {
		return
	}

	if a_base.AnalysisCasesMap[flags.MixedDeadlock] {
		a_scenarios.CheckForMixedDeadlock()
	}

	if control.WasCanceled() {
		return
	}

	if a_base.AnalysisCasesMap[flags.UnlockBeforeLock] {
		a_scenarios.CheckForUnlockBeforeLock()
	}
}
