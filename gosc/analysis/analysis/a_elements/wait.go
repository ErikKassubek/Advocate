// Copyright (c) 2024 
//
// File: vcWait.go
// Brief: Update functions for happens before info for wait group operations
//        Some function start analysis functions
//
//
// License: BSD-3-Clause

package a_elements

import (
	"gosc/analysis/a_base"
	"gosc/analysis/analysis/a_scenarios"
	"gosc/analysis/hb/a_hbcalc"
	"gosc/fuzzing/f_base"
	"gosc/trace"
	"gosc/utils/flags"
	"gosc/utils/log"
)

// AnalyzeWait updates and stores the vector clock of the element
// Parameter:
//   - wa *TraceElementWait: the wait trace element
func AnalyzeWait(wa *trace.ElementWait) {
	a_hbcalc.UpdateHBWait(wa)

	switch wa.Type(true) {
	case trace.WaitAdd, trace.WaitDone:
		a_base.LastChangeWG[wa.ResourceID()] = wa

		if a_base.AnalysisCasesMap[flags.DoneBeforeAdd] || f_base.FuzzingModeGoCRHBPlus {
			a_scenarios.CheckForDoneBeforeAddChange(wa)
		}
	case trace.WaitWait:
	default:
		err := "Unknown operation on wait group: " + wa.String()
		log.Error(err)
	}
}
