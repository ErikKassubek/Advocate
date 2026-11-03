// Copyright (c) 2024 Erik Kassubek
//
// File: analysisClose.go
// Brief: Trace analysis for send, receive and close on closed channel
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/results/buginfo"
	"advocate/utils/results/results"
	"advocate/utils/timer"
)

// CheckForCommunicationOnClosedChannel checks if a send or receive on a
// closed channel is possible.
// It it is possible, print a warning or error.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func CheckForCommunicationOnClosedChannel(ch *trace.ElementChannel) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	id := ch.ResourceID()

	// check if there is an earlier send, that could happen concurrently to close
	if a_base.AnalysisCasesMap[flags.SendOnClosed] && a_base.HasSend[id] {
		for routine, mrs := range a_base.MostRecentSend {
			happensBefore := a_clock.GetHappensBefore(mrs[id].Vc, a_base.CloseData[id].GetVC(a_clock.Strong))

			elem := mrs[id].Elem

			if elem != nil && happensBefore != a_hb.Before {

				arg1 := results.TraceElementResult{ // send
					RoutineID:  routine,
					ResourceID: id,
					TRequest:   elem.T(trace.Request),
					ObjType:    "CS",
					File:       elem.File(),
					Line:       elem.Line(),
				}

				arg2 := results.TraceElementResult{ // close
					RoutineID:  a_base.CloseData[ch.ResourceID()].RoutineID(),
					ResourceID: id,
					TRequest:   ch.T(trace.Request),
					ObjType:    "CC",
					File:       ch.File(),
					Line:       ch.Line(),
				}

				results.Result(results.CRITICAL, buginfo.PSendOnClosed, []results.ResultElem{arg1}, []results.ResultElem{arg2})
			}
		}
	}

}

// FoundSendOnClosedChannel is called, id an actual send on closed was found.
//
// Parameter:
//   - elem TraceElement: the send/select elem
//   - id int: id of the channel
//   - actual bool: set actual to true it the panic occurred, set to false if it is in an not triggered select case
func FoundSendOnClosedChannel(elem trace.Element, actual bool) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	id := elem.ResourceID()

	if _, ok := a_base.CloseData[id]; !ok {
		return
	}

	closeElem := a_base.CloseData[id]
	fileSend := elem.File()

	if fileSend == "" || fileSend == "\n" {
		return
	}

	arg1 := results.TraceElementResult{ // send
		RoutineID:  elem.RoutineID(),
		ResourceID: id,
		TRequest:   elem.T(trace.Request),
		ObjType:    "CS",
		File:       fileSend,
		Line:       elem.Line(),
	}

	arg2 := results.TraceElementResult{ // close
		RoutineID:  a_base.CloseData[id].RoutineID(),
		ResourceID: id,
		TRequest:   closeElem.T(trace.Request),
		ObjType:    "CC",
		File:       closeElem.File(),
		Line:       closeElem.Line(),
	}

	if actual {
		results.Result(results.CRITICAL, buginfo.ASendOnClosed,
			[]results.ResultElem{arg1}, []results.ResultElem{arg2})
	} else {
		results.Result(results.CRITICAL, buginfo.PSendOnClosed,
			[]results.ResultElem{arg1}, []results.ResultElem{arg2})
	}

}

// CheckForClosedOnClosed checks for a close on a closed channel.
// Must be called, before the current close operation is added to closePos
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func CheckForClosedOnClosed(ch *trace.ElementChannel) {
	timer.Start(timer.AnaClose)
	defer timer.Stop(timer.AnaClose)

	id := ch.ResourceID()

	if oldClose, ok := a_base.CloseData[id]; ok {
		if oldClose.ID() == 0 || ch.ID() == 0 {
			return
		}

		arg1 := results.TraceElementResult{
			RoutineID:  ch.RoutineID(),
			ResourceID: id,
			TRequest:   oldClose.T(trace.Request),
			ObjType:    "CC",
			File:       oldClose.File(),
			Line:       oldClose.Line(),
		}

		arg2 := results.TraceElementResult{
			RoutineID:  ch.RoutineID(),
			ResourceID: id,
			TRequest:   ch.T(trace.Request),
			ObjType:    "CC",
			File:       ch.File(),
			Line:       ch.Line(),
		}

		results.Result(results.CRITICAL, buginfo.ACloseOnClosed,
			[]results.ResultElem{arg1}, []results.ResultElem{arg2})
	}
}
