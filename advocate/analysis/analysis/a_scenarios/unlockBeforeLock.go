// Copyright (c) 2024 Erik Kassubek
//
// File: analysisUnlockBeforeLock.go
// Brief: Analysis for unlock of not locked mutex
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
	"advocate/utils/log"
	"advocate/utils/results/buginfo"
	"advocate/utils/results/results"
	"advocate/utils/timer"
	"advocate/utils/types"
)

// CheckForUnlockBeforeLockLock collects all locks for the analysis
//
// Parameter:
//   - mu *TraceElementMutex: the trace mutex element
func CheckForUnlockBeforeLockLock(mu *trace.ElementMutex) {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	id := mu.ResourceID()

	if _, ok := a_base.AllLocks[id]; !ok {
		a_base.AllLocks[id] = make([]trace.Element, 0)
	}

	a_base.AllLocks[id] = append(a_base.AllLocks[id], mu)
}

// CheckForUnlockBeforeLockUnlock collects all unlocks for the analysis
//
// Parameter:
//   - mu *TraceElementMutex: the trace mutex element
func CheckForUnlockBeforeLockUnlock(mu *trace.ElementMutex) {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	id := mu.ResourceID()

	if _, ok := a_base.AllLocks[id]; !ok {
		a_base.AllUnlocks[id] = make([]trace.Element, 0)
	}

	a_base.AllUnlocks[id] = append(a_base.AllUnlocks[id], mu)
}

// CheckForUnlockBeforeLock check if we can get a unlock of a not locked mutex
// For each done operation, build a bipartite st graph.
// Use the Ford-Fulkerson algorithm to find the maximum flow.
// If the maximum flow is smaller than the number of unlock operations, a unlock before lock is possible.
func CheckForUnlockBeforeLock() {
	timer.Start(timer.AnaUnlock)
	defer timer.Stop(timer.AnaUnlock)

	for id := range a_base.AllUnlocks { // for all mutex ids
		// if a lock and the corresponding unlock is always in the same routine, this cannot happen
		if trace.SameRoutine(a_base.AllLocks[id], a_base.AllUnlocks[id]) {
			continue
		}

		graph := buildResidualGraph(a_base.AllLocks[id], a_base.AllUnlocks[id])

		maxFlow, graph, err := calculateMaxFlow(graph)
		if err != nil {
			log.Error("Could not check for unlock before lock: ", err)
		}

		nrUnlock := len(a_base.AllUnlocks)

		locks := make([]trace.Element, 0)
		unlocks := make([]trace.Element, 0)

		if maxFlow < nrUnlock {
			for _, l := range a_base.AllLocks[id] {
				if !types.Contains(graph[&drain], l) {
					locks = append(locks, l)
				}
			}

			for _, u := range graph[&source] {
				unlocks = append(unlocks, u)
			}

			locksSorted := make([]trace.Element, 0)
			unlockSorted := make([]trace.Element, 0)

			for i := 0; i < len(locks); {
				removed := false
				for j := 0; j < len(unlocks); {
					if a_clock.GetHappensBefore(locks[i].VC(a_clock.Strong), unlocks[j].VC(a_clock.Strong)) == a_hb.Concurrent {
						locksSorted = append(locksSorted, locks[i])
						unlockSorted = append(unlockSorted, unlocks[i])
						locks = append(locks[:i], locks[i+1:]...)
						unlocks = append(unlocks[:j], unlocks[j+1:]...)
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

			args1 := []results.ResultElem{} // unlocks
			args2 := []results.ResultElem{} // locks

			for _, u := range unlockSorted {
				if u.ID() == 0 {
					continue
				}

				args1 = append(args1, results.TraceElementResult{
					RoutineID:  u.RoutineID(),
					ResourceID: id,
					TRequest:   u.T(trace.Request),
					ObjType:    u.Type(true),
					File:       u.File(),
					Line:       u.Line(),
				})
			}

			for _, l := range locksSorted {
				if l.ID() == 0 {
					continue
				}

				args2 = append(args2, results.TraceElementResult{
					RoutineID:  l.RoutineID(),
					ResourceID: id,
					TRequest:   l.T(trace.Request),
					ObjType:    l.Type(true),
					File:       l.File(),
					Line:       l.Line(),
				})
			}

			results.Result(results.CRITICAL, buginfo.PUnlockBeforeLock, args1, args2)
		}
	}
}
