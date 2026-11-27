// Copyright (c) 2024 Erik Kassubek
//
// File: vcWait.go
// Brief: Update functions for happens before info for wait group operations
//        Some function start analysis functions
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"advocate/analysis/a_base"
	"advocate/analysis/analysis/a_scenarios"
	"advocate/analysis/hb/a_hbcalc"
	"advocate/fuzzing/f_base"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
)

// AnalyzeWait updates and stores the vector clock of the element
// Parameter:
//   - wa *TraceElementWait: the wait trace element
func AnalyzeWait(wa *trace.ElementWait) {
	a_hbcalc.UpdateHBWait(wa)

	switch wa.Type(true) {
	case trace.WaitAdd, trace.WaitDone:
		a_base.LastChangeWG[wa.ResourceID()] = wa

		if flags.AnalysisCasesMap[flags.DoneBeforeAdd] || f_base.FuzzingModeGoCRHBPlus {
			a_scenarios.CheckForDoneBeforeAddChange(wa)
		}
	case trace.WaitWait:
	default:
		err := "Unknown operation on wait group: " + wa.String()
		log.Error(err)
	}
}
