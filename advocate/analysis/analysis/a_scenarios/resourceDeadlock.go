// advocate/analysis/analysis/scenarios/resourceDeadlock.go

// Copyright (c) 2024 Erik Kassubek
//
// File: analysisResourceDeadlock.go
// Brief: Alternative analysis for cyclic mutex deadlocks.
//
// Author: Sebastian Pohsner
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
	"advocate/utils/results/buginfo"
	"advocate/utils/results/results"
	"advocate/utils/timer"
)

// Computation of "abstract" lock dependencies
// Lock dependencies are 3-tuples of the following form:
//    (ThreadID, Lock, LockSet)
// Lock dependencies are computed thread-local.
// For each thread there might be several (acquire) events that lead to "lock" acquired under some "lockset".
//
// Each acquire event carries its own vector clock.
// We wish to make use of vector clocks to eliminate infeasible replay candidates.
//
// This means that lock dependencies are 4-tuples of the following form:
//    (ThreadID, Lock, LockSet, []Event)

// ALGORITHM
//
// There are two phases.
//  1. Recording of lock dependencies.
//  2. Checking if lock dependencies imply a cycle.

// Algorithm phase 1

// We show the event processing functions for acquire and release.

func acquire(s *a_base.State, readLock bool, event trace.Element) {
	if _, exists := s.Routines[event.RoutineID()]; !exists {
		s.Routines[event.RoutineID()] = a_base.Thread{
			CurrentLockset:   make(a_base.Lockset),
			LockDependencies: make(map[a_base.LockID][]a_base.Dependency),
			ReaderCounter:    make(map[a_base.LockID]int),
		}
	}

	lockID := a_base.LockID{
		ID:       event.ResourceID(),
		ReadLock: readLock,
	}

	ls := s.Routines[event.RoutineID()].CurrentLockset
	if !ls.Empty() {
		deps := s.Routines[event.RoutineID()].LockDependencies
		deps[lockID] = insert(deps[lockID], ls, event)
	}

	if lockID.IsRead() {
		lockID.AddReader(s.Routines[event.RoutineID()])
	}
	s.Routines[event.RoutineID()].CurrentLockset.Add(lockID, event)
}

func release(s *a_base.State, readLock bool, event trace.Element) {
	lockID := a_base.LockID{
		ID:       event.ResourceID(),
		ReadLock: readLock,
	}

	if lockID.IsRead() {
		lockID.RemoveReader(s.Routines[event.RoutineID()])
		for _, thread := range s.Routines {
			if lockID.HasReaders(thread) {
				continue
			}
			thread.CurrentLockset.Remove(lockID)
		}
		s.Routines[event.RoutineID()].CurrentLockset.Remove(lockID)
	} else {
		if !s.Routines[event.RoutineID()].CurrentLockset.Remove(lockID) {
			// "Lock not found in lockset! Has probably been released in another thread, this is an unsupported case."
			s.Failed = true
		}
	}
}

// Insert a new lock dependency for a given thread and lock x.
// We assume that event e acquired lock x.
// We might have already an entry that shares the same lock and lockset!
func insert(dependencies []a_base.Dependency, ls a_base.Lockset, event trace.Element) []a_base.Dependency {
	for i, v := range dependencies {
		if v.Lockset.Equal(ls) {
			dependencies[i].Requests = append(dependencies[i].Requests, event)
			return dependencies
		}
	}
	return append(dependencies, a_base.Dependency{
		Lockset:  ls.Clone(),
		Requests: []trace.Element{event}},
	)
}

// Algorithm phase 2

// Based on lock dependencies we can check for cycles.
// A cycle involves n threads and results from some n lock dependencies.
// For each thread we record the requests that might block.

func report(s *a_base.State, c a_base.Cycle) {
	s.Cycles = append(s.Cycles, c)
}

// After phase 1, the following function yields all cycle lock dependencies.

// The implementation below follows the algorithm used in UNDEAD (https://github.com/UTSASRG/UnDead/blob/master/analyzer.hh)
func getCycles(s *a_base.State) []a_base.Cycle {
	if s.Cycles != nil {
		return s.Cycles
	}
	s.Cycles = []a_base.Cycle{}

	traversedThread := make(map[int]bool)
	for tid := range s.Routines {
		traversedThread[tid] = false
	}

	var chainStack []a_base.LockDependency
	for threadID := range traversedThread {
		traversedThread[threadID] = true
		for lock, dependencies := range s.Routines[threadID].LockDependencies {
			for _, dependency := range dependencies {
				chainStack = append(chainStack, a_base.LockDependency{
					Thread:   threadID,
					Lock:     lock,
					Lockset:  dependency.Lockset,
					Requests: dependency.Requests,
				}) // push
				dfs(s, &chainStack, traversedThread)
				chainStack = chainStack[:len(chainStack)-1] // pop
			}
		}
	}

	return s.Cycles
}

func dfs(s *a_base.State, chainStack *[]a_base.LockDependency, traversedThread map[int]bool) {
	for tid, isTraversed := range traversedThread {
		if isTraversed {
			continue
		}

		for l, lD := range s.Routines[tid].LockDependencies {
			for _, lLsD := range lD {
				ld := a_base.LockDependency{
					Thread:   tid,
					Lock:     l,
					Lockset:  lLsD.Lockset,
					Requests: lLsD.Requests,
				}
				if isChain(chainStack, ld) {
					if isCycleChain(chainStack, ld) {
						var c a_base.Cycle = make([]a_base.LockDependency, len(*chainStack)+1)
						for i, d := range *chainStack {
							c[i] = d.Clone()
						}
						c[len(*chainStack)] = ld

						// Check for infeasible deadlocks
						if checkAndFilterConcurrentRequests(&c) {
							report(s, c)
						}
					} else {
						traversedThread[tid] = true
						*chainStack = append(*chainStack, ld) // push
						dfs(s, chainStack, traversedThread)
						*chainStack = (*chainStack)[:len(*chainStack)-1] // pop
						traversedThread[tid] = false
					}
				}
			}
		}
	}
}

// Check if adding dependency to chain will still be a chain.
func isChain(chainStack *[]a_base.LockDependency, dependency a_base.LockDependency) bool {

	for _, d := range *chainStack {
		// Exit early. No two deps can hold the same lock. - Except for read locks
		if d.Lock == dependency.Lock && dependency.Lock.IsWrite() {
			// Two dependencies hold the same lock (early exit)
			return false
		}
		// Check (LD-1) LS(ls_j) cap LS(ls_i+1) for j in {1,..,i}
		// Also (RW-LD-1)
		if !d.Lockset.DisjointCouldBlock(dependency.Lockset) {
			// Locksets are not disjoint (guard)
			return false
		}
	}

	// Check (LD-2) l_i in ls_i+1
	for l := range dependency.Lockset {

		// Also (RW-LD-2)
		if (*chainStack)[len(*chainStack)-1].Lock.EqualsCouldBlock(l) {
			return true
		}

	}
	// Previous lock not in current lockset or both are read locks
	return false
}

// Check (LD-3) l_n in ls_1
// Also (RW-LD-3)
func isCycleChain(chainStack *[]a_base.LockDependency, dependency a_base.LockDependency) bool {
	for l := range (*chainStack)[0].Lockset {
		if l.EqualsCouldBlock(dependency.Lock) {
			return true
		}
	}
	// Cycle Chain does not close
	return false
}

// checkAndFilterConcurrentRequests checks if there is one ore more chains of
// concurrent requests and filter out any requests that are not part of them
//
// Parameter:
//   - cycle *baseA.Cycle: a cycle to check
//
// Returns:
//   - bool: true if the cycle is valid regarding hb, false otherwise
func checkAndFilterConcurrentRequests(cycle *a_base.Cycle) bool {
	for i := range *cycle {
		// Check if each request has a concurrent request in the element before and after
		// All requests that have a previous request
		requestsWithPrev := []trace.Element{}
		for _, req := range (*cycle)[i].Requests {
			for _, prevReq := range (*cycle)[(len(*cycle)+i-1)%len(*cycle)].Requests {
				if a_clock.GetHappensBefore(req.GetVC(a_clock.Weak), prevReq.GetVC(a_clock.Weak)) == a_hb.Concurrent {
					requestsWithPrev = append(requestsWithPrev, req)
					break
				}
			}
		}
		// All requests that have a next request
		requestsWithBoth := []trace.Element{}
		for _, req := range requestsWithPrev {
			for _, nextReq := range (*cycle)[(i+1)%len(*cycle)].Requests {
				if a_clock.GetHappensBefore(req.GetVC(a_clock.Weak), nextReq.GetVC(a_clock.Weak)) == a_hb.Concurrent {
					requestsWithBoth = append(requestsWithBoth, req)
					break
				}
			}
		}

		if len(requestsWithBoth) > 0 {
			// Only requests with predecessors and successors remain
			(*cycle)[i].Requests = requestsWithBoth
		} else {
			// An entry with no requests mean that we no longer have a valid cycle
			// Cycle Entry with no concurrent requests
			return false
		}
	}
	return true
}

// ResetState resets the current state of the resource deadlock detection
func ResetState() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	a_base.CurrentState = a_base.State{
		Routines: make(map[int]a_base.Thread),
		Cycles:   nil,
		Failed:   false,
	}
}

// HandleMutexEventForRessourceDeadlock processes an mutex operation for the
// resource deadlock detection
//
// Parameter:
//   - element trace.ElementMutex: the trace element
func HandleMutexEventForRessourceDeadlock(element trace.ElementMutex) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	if a_base.CurrentState.Failed {
		return
	}

	switch element.Type(true) {
	case trace.MutexLock, trace.MutexTryLock:
		acquire(&a_base.CurrentState, false, &element)
		// We do not check event.suc because that could led to false negatives
	case trace.MutexRLock:
		acquire(&a_base.CurrentState, true, &element)
	case trace.MutexUnlock:
		release(&a_base.CurrentState, false, &element)
	case trace.MutexRUnlock:
		release(&a_base.CurrentState, true, &element)
	}
}

// CheckForResourceDeadlock searches for cycles which imply a cyclic resource
// deadlock
func CheckForResourceDeadlock() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)
	if a_base.CurrentState.Failed {
		// log.Error("Failed flag is set, probably encountered unsupported lock operation. No deadlock analysis possible.")
		return
	}
	// for i, t := range baseA.CurrentState.threads {
	// 	debugLog("Found", len(t.lockDependencies), "dependencies in Thread", i)
	// }

	getCycles(&a_base.CurrentState)

	// debugLog("Found", len(baseA.CurrentState.cycles), "cycles")

	for _, cycles := range a_base.CurrentState.Cycles {
		var stuckElements []results.ResultElem
		var acquiredElements []results.ResultElem
		stuckIds := make(map[int]bool)
		var request = findEarliestRequest(cycles)

		// debugLog("Found cycle with the following entries:", cycle)
		for _, cycle := range cycles {
			for _, r := range cycle.Requests {
				if a_clock.GetHappensBefore(request.GetVC(a_clock.Weak), r.GetVC(a_clock.Weak)) == a_hb.Concurrent {
					request = r
					break
				}
			}

			stuckElements = append(stuckElements, results.TraceElementResult{
				RoutineID:  request.RoutineID(),
				ResourceID: request.ResourceID(),
				TRequest:   request.T(trace.Request),
				ObjType:    request.Type(true),
				File:       request.File(),
				Line:       request.Line(),
			})
			stuckIds[request.ResourceID()] = true

		}

		for _, cycle := range cycles {
			for _, elem := range cycle.Lockset {
				if _, ok := stuckIds[elem.ResourceID()]; ok {
					acquiredElements = append(acquiredElements, results.TraceElementResult{
						RoutineID:  elem.RoutineID(),
						ResourceID: elem.ResourceID(),
						TRequest:   elem.T(trace.Request),
						ObjType:    elem.Type(true),
						File:       elem.File(),
						Line:       elem.Line(),
					})
				}
			}
		}

		if len(stuckElements) == 0 {
			return
		}

		results.Result(results.CRITICAL, buginfo.PCyclicDeadlock, stuckElements, acquiredElements)
	}
}

/////////////////////////////////
// Auxiliary functions.

// Finds the earliest request in a cycle.
func findEarliestRequest(cycle []a_base.LockDependency) trace.Element {
	earliest := cycle[0].Requests[0]
	earliestTime := earliest.T(trace.Request)
	for _, c := range cycle {
		for _, r := range c.Requests {
			requestTime := r.T(trace.Request)

			if requestTime < earliestTime {
				earliest = r
				earliestTime = requestTime
			}
		}
	}
	return earliest
}

// Further notes.
//
// If possible we would like to use a double-indexed map of the following form.
//
// type Deps map[Lock]map[Lockset][]Event
//
// Unfortunately, this is not possible in Go because keys must be comparable (but slices, maps, ... are not comparable).
// This is not an issue in Haskell or C++ where we can extend the set of comparable types (but providing additional definitions for "==",...)
//
// Hence, we use single-indexed (by Lock) map.
