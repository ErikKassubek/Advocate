// Copyright (c) 2024 Erik Kassubek
//
// File: analysisConcurrentCommunication.go
// Brief: Find concurrent operations on the same element
//   For concurrent receive: add panic
//   For concurrent send, receive, (try)(r)lock, once.Do: store to use in fuzzing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/analysis/hb/a_vc"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/results/buginfo"
	"advocate/utils/results/results"
	"advocate/utils/timer"
)

// GetConcurrentSendForFuzzing checks if for the given send, if there is a
// concurrent send on the same channel. If there is, the information is stored
// in baseA.FuzzingFlowSend. This is used for fuzzing.
//
// Parameter:
//   - sender *TraceElementChannel: Send trace element
func GetConcurrentSendForFuzzing(sender *trace.ElementChannel) {
	timer.Start(timer.FuzzingAna)
	defer timer.Stop(timer.FuzzingAna)

	id := sender.ResourceID()
	routine := sender.RoutineID()

	IncFuzzingCounter(sender)

	if sender.Committed() {
		return
	}

	for r, elem := range a_base.LastSendRoutine {
		if r == routine {
			continue
		}

		if elem[id].Vc == nil || elem[id].Vc.GetClock() == nil {
			continue
		}

		happensBefore := a_clock.GetHappensBefore(elem[id].Vc, a_vc.CurrentVC[routine])
		if happensBefore == a_hb.Concurrent {
			elem2 := elem[id].Elem
			a_base.FuzzingFlowSend = append(a_base.FuzzingFlowSend, a_base.ConcurrentEntry{Elem: elem2, Counter: getFuzzingCounter(elem2), Type: a_base.CERecv})
		}
	}

	if sender.Committed() {
		if _, ok := a_base.LastSendRoutine[routine]; !ok {
			a_base.LastSendRoutine[routine] = make(map[int]a_base.ElemWithVc)
		}

		a_base.LastSendRoutine[routine][id] = a_base.ElemWithVc{Vc: a_vc.CurrentVC[routine].Copy(), Elem: sender}
	}
}

// CheckForConcurrentRecv checks if for the given recv, if there is a
// concurrent recv on the same channel. If there is, the information is stored
// in baseA.FuzzingFlowRecv.
//
// Parameter:
//   - ch *TraceElementChannel: recv trace element
func CheckForConcurrentRecv(ch *trace.ElementChannel, vc map[int]*a_clock.VectorClock) {
	if a_base.AnalysisFuzzingFlow {
		timer.Start(timer.FuzzingAna)
		defer timer.Stop(timer.FuzzingAna)
	}
	timer.Start(timer.AnaConcurrent)
	defer timer.Stop(timer.AnaConcurrent)

	id := ch.ResourceID()
	routine := ch.RoutineID()

	IncFuzzingCounter(ch)

	for r, elem := range a_base.LastRecvRoutine {
		if r == routine {
			continue
		}

		if elem[id].Vc == nil || elem[id].Vc.GetClock() == nil {
			continue
		}

		happensBefore := a_clock.GetHappensBefore(elem[id].Vc, vc[routine])
		if happensBefore == a_hb.Concurrent {

			elem2 := elem[id].Elem

			if a_base.AnalysisFuzzingFlow {
				if !ch.Committed() {
					a_base.FuzzingFlowRecv = append(a_base.FuzzingFlowRecv, a_base.ConcurrentEntry{Elem: elem2, Counter: getFuzzingCounter(elem2), Type: a_base.CERecv})
				}
			}

			if a_base.AnalysisCasesMap[flags.ConcurrentRecv] {
				arg1 := results.TraceElementResult{
					RoutineID:  routine,
					ResourceID: id,
					TRequest:   ch.T(trace.Request),
					ObjType:    "CR",
					File:       ch.File(),
					Line:       ch.Line(),
				}

				arg2 := results.TraceElementResult{
					RoutineID:  r,
					ResourceID: id,
					TRequest:   elem2.T(trace.Request),
					ObjType:    "CR",
					File:       elem2.File(),
					Line:       elem2.Line(),
				}

				results.Result(results.WARNING, buginfo.AConcurrentRecv,
					[]results.ResultElem{arg1}, []results.ResultElem{arg2})
			}
		}
	}

	if ch.Committed() {
		if _, ok := a_base.LastRecvRoutine[routine]; !ok {
			a_base.LastRecvRoutine[routine] = make(map[int]a_base.ElemWithVc)
		}

		a_base.LastRecvRoutine[routine][id] = a_base.ElemWithVc{Vc: vc[routine].Copy(), Elem: ch}
	}
}

// GetConcurrentMutexForFuzzing checks if for the given mutex operations, if there is a
// concurrent mutex operations on the same mutex. If there is, the information is stored
// in baseA.FuzzingFlowMutex.
//
// Parameter:
//   - mu *TraceElementMutex: mutex operations
func GetConcurrentMutexForFuzzing(mu *trace.ElementMutex) {
	timer.Start(timer.FuzzingAna)
	defer timer.Stop(timer.FuzzingAna)

	// operation executed normally
	if mu.IsSuc() {
		return
	}

	id := mu.ResourceID()

	// not executed try lock
	// get currently hold lock because of witch the try lock failed

	if val, ok := a_base.CurrentlyHoldLock[id]; !ok || val == nil {
		log.Error("Failed trylock even throw mutex is not locked: ", mu.String())
	}

	elem := a_base.CurrentlyHoldLock[id]

	if a_clock.GetHappensBefore(mu.VC(a_clock.Strong), elem.VC(a_clock.Strong)) == a_hb.Concurrent {
		a_base.FuzzingFlowMutex = append(a_base.FuzzingFlowMutex, a_base.ConcurrentEntry{Elem: elem, Counter: getFuzzingCounter(elem), Type: a_base.CEMutex})
	}

}

// GetConcurrentOnceForFuzzing checks if for the given once operations, if there is a
// concurrent once operations on the same primitive. If there is, the information is stored
// in baseA.FuzzingFlowOnce.
//
// Parameter:
//   - on *TraceElementOnce: once.Do operations
func GetConcurrentOnceForFuzzing(on *trace.ElementOnce) {
	timer.Start(timer.FuzzingAna)
	timer.Stop(timer.FuzzingAna)

	id := on.ResourceID()
	vc := on.VC(a_clock.Strong)

	IncFuzzingCounter(on)

	if on.GetSuc() {
		a_base.ExecutedOnce[id] = &a_base.ConcurrentEntry{Elem: on, Counter: getFuzzingCounter(on), Type: a_base.CEOnce}
		return
	}

	if exec, ok := a_base.ExecutedOnce[id]; ok {
		if a_clock.GetHappensBefore(exec.Elem.VC(a_clock.Strong), vc) == a_hb.Concurrent {
			a_base.FuzzingFlowOnce = append(a_base.FuzzingFlowOnce, *exec)
		}
	}
}

// GetConcurrentInfoForFuzzing returns the required fuzzing information for
// the flow fuzzing mutation.
//
// Returns:
//   - *[]ConcurrentEntry: once that can be delayed in flow fuzzing
//   - *[]ConcurrentEntry: mutex operations that can be delayed in flow fuzzing
//   - *[]ConcurrentEntry: send that can be delayed in flow fuzzing
//   - *[]ConcurrentEntry: recv that can be delayed in flow fuzzing
func GetConcurrentInfoForFuzzing() (*[]a_base.ConcurrentEntry, *[]a_base.ConcurrentEntry, *[]a_base.ConcurrentEntry, *[]a_base.ConcurrentEntry) {
	return &a_base.FuzzingFlowOnce, &a_base.FuzzingFlowMutex, &a_base.FuzzingFlowSend, &a_base.FuzzingFlowRecv
}

// getFuzzingCounter returns the fuzzing counter for an element. If the element
// has no counter it is set to 0. The fuzzing counter gives for a given
// primitive how often an operation has been executed on the primitive before.
//
// Parameter:
//   - te TraceElement: The trace element to get the counter for
//
// Returns:
//   - int: the current fuzzing counter for the element
func getFuzzingCounter(te trace.Element) int {
	id := te.ResourceID()
	pos := te.Pos().String()

	if _, ok := a_base.FuzzingCounter[id]; !ok {
		return 0
	}

	if val, ok := a_base.FuzzingCounter[id][pos]; ok {
		return val
	}
	return 0
}

// IncFuzzingCounter increases the fuzzing counter of a given element
//
// Parameter:
//   - te TraceElement: The element to increase the counter for
func IncFuzzingCounter(te trace.Element) {
	id := te.ResourceID()
	pos := te.Pos().String()

	if _, ok := a_base.FuzzingCounter[id]; !ok {
		a_base.FuzzingCounter[id] = make(map[string]int)
	}

	a_base.FuzzingCounter[id][pos] = a_base.FuzzingCounter[id][pos] + 1
}
