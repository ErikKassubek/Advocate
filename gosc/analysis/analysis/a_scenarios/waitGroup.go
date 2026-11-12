// Copyright (c) 2024 
//
// File: analysisWaitGroup.go
// Brief: Trace analysis for possible negative wait group counter
//
//
// License: BSD-3-Clause

package a_scenarios

import (
	"gosc/analysis/a_base"
	"gosc/analysis/a_hb"
	"gosc/analysis/hb/a_clock"
	"gosc/trace"
	"gosc/utils/log"
	"gosc/utils/results/buginfo"
	"gosc/utils/results/results"
	"gosc/utils/timer"
	"gosc/utils/types"
)

// CheckForDoneBeforeAddChange collect all adds and dones for the analysis
//
// Parameter:
//   - wa *TraceElementWait: the trace wait or done element
func CheckForDoneBeforeAddChange(wa *trace.ElementWait) {
	timer.Start(timer.AnaWait)
	defer timer.Stop(timer.AnaWait)

	delta := wa.GetDelta()

	if delta > 0 {
		CheckForDoneBeforeAddAdd(wa)
	} else if delta < 0 {
		CheckForDoneBeforeAddDone(wa)
	} else {
		// checkForImpossibleWait(routine, id, pos, vc)
	}
}

// CheckForDoneBeforeAddAdd collect all adds for the analysis
//
// Parameter:
//   - wa *TraceElementWait: the trace wait element
func CheckForDoneBeforeAddAdd(wa *trace.ElementWait) {
	id := wa.ResourceID()

	// if necessary, create maps and lists
	if _, ok := a_base.WGAddData[id]; !ok {
		a_base.WGAddData[id] = make([]trace.Element, 0)
	}

	// add the vector clock and position to the list
	a_base.WGAddData[id] = append(a_base.WGAddData[id], wa)
}

// CheckForDoneBeforeAddDone collect all dones for the analysis
//
// Parameter:
//   - wa *TraceElementWait: the trace done element
func CheckForDoneBeforeAddDone(wa *trace.ElementWait) {
	id := wa.ResourceID()

	// if necessary, create maps and lists
	if _, ok := a_base.WgDoneData[id]; !ok {
		a_base.WgDoneData[id] = make([]trace.Element, 0)

	}

	// add the vector clock and position to the list
	a_base.WgDoneData[id] = append(a_base.WgDoneData[id], wa)
}

// CheckForDoneBeforeAdd checks if a wait group counter could become negative
// For each done operation, build a bipartite st graph.
// Use the Ford-Fulkerson algorithm to find the maximum flow.
// If the maximum flow is smaller than the number of done operations, a negative wait group counter is possible.
func CheckForDoneBeforeAdd() {
	timer.Start(timer.AnaWait)
	defer timer.Stop(timer.AnaWait)

	for id := range a_base.WGAddData { // for all waitgroups
		graph := buildResidualGraph(a_base.WGAddData[id], a_base.WgDoneData[id])

		maxFlow, graph, err := calculateMaxFlow(graph)
		if err != nil {
			log.Error("Could not check for done before add: ", err)
		}
		nrDone := len(a_base.WgDoneData[id])

		addsNegWg := make([]trace.Element, 0)
		donesNegWg := make([]trace.Element, 0)

		if maxFlow < nrDone {
			// sort the adds and dones, that do not have a partner is such a way,
			// that the i-th add in the result message is concurrent with the
			// i-th done in the result message

			for _, add := range a_base.WGAddData[id] {
				if !types.Contains(graph[&drain], add) {
					addsNegWg = append(addsNegWg, add)
				}
			}

			for _, dones := range graph[&source] {
				donesNegWg = append(donesNegWg, dones)
			}

			addsNegWgSorted := make([]trace.Element, 0)
			donesNEgWgSorted := make([]trace.Element, 0)

			for i := 0; i < len(addsNegWg); {
				removed := false
				for j := 0; j < len(donesNegWg); {
					if a_clock.GetHappensBefore(addsNegWg[i].VC(a_clock.Strong), donesNegWg[j].VC(a_clock.Strong)) == a_hb.Concurrent {
						addsNegWgSorted = append(addsNegWgSorted, addsNegWg[i])
						donesNEgWgSorted = append(donesNEgWgSorted, donesNegWg[j])
						// remove the element from the list
						addsNegWg = append(addsNegWg[:i], addsNegWg[i+1:]...)
						donesNegWg = append(donesNegWg[:j], donesNegWg[j+1:]...)
						// fix the index
						removed = true
						break
					} else {
						j++
					}
				}
				if !removed {
					i++
				}
			}

			if len(donesNEgWgSorted) == 0 {
				return
			}

			args1 := []results.ResultElem{} // dones
			args2 := []results.ResultElem{} // adds

			for _, done := range donesNEgWgSorted {
				if done.ID() == 0 {
					continue
				}

				args1 = append(args1, results.TraceElementResult{
					RoutineID:  done.RoutineID(),
					ResourceID: id,
					TRequest:   done.T(trace.Request),
					ObjType:    "WD",
					File:       done.File(),
					Line:       done.Line(),
				})
			}

			for _, add := range addsNegWgSorted {
				if add.ID() == 0 {
					continue
				}

				args2 = append(args2, results.TraceElementResult{
					RoutineID:  add.RoutineID(),
					ResourceID: id,
					TRequest:   add.T(trace.Request),
					ObjType:    "WA",
					File:       add.File(),
					Line:       add.Line(),
				})

			}

			results.Result(results.CRITICAL, buginfo.PNegWG, args1, args2)
		}
	}
}
