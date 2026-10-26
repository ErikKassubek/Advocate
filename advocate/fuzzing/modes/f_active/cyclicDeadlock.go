// Copyright (c) 2024 Erik Kassubek
//
// File: cyclicDeadlock.go
// Brief: Rewrite trace for cyclic deadlocks
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_active

import (
	"advocate/fuzzing/f_roc"
	"advocate/trace"
	"advocate/utils/results/bugs"
	"errors"
)

// TODO: rewrite with constraints and closure

// rewriteCyclicDeadlock rewrites the trace in such a way, that it should
// trigger the cyclic/resource deadlock described in the bug
//
// Parameter:
//   - bug bugs.Bug: the bug that should be triggered by the rewrite
//
// Returns:
//   - *f_roc.Constraint: the constraint containing the bug
//   - error
func rewriteCyclicDeadlock(bug bugs.Bug) (*f_roc.Constraint, error) {
	if len(bug.TraceElement2) == 0 {
		return nil, errors.New("no trace elements in bug")
	}

	if len(bug.TraceElement2) < 2 {
		return nil, errors.New("at least 2 trace elements are needed for a deadlock")
	}

	constr := f_roc.NewConstraint()

	for _, elem := range bug.TraceElement2 { // acquired
		constr.Add(elem)
	}

	for _, elem := range bug.TraceElement1 { // blocking
		constr.Add(elem)
	}

	// lastTime := findLastTime(bug.TraceElement2)

	// // remove tail after lastTime and the last lock
	// tr.ShortenTrace(lastTime, true)
	// for _, elem := range bug.TraceElement2 {
	// 	tr.ShortenRoutine(elem.RoutineID(), elem.T(trace.Sorting))
	// }

	// var locksetElements []trace.Element

	// // Find the lockset elements
	// for i, elem := range bug.TraceElement2 {
	// 	// This is one is guranteed to be in the lockset of elem
	// 	prevElement := bug.TraceElement2[(i+len(bug.TraceElement2)-1)%len(bug.TraceElement2)]
	// 	for j := tr.GetRoutineTrace(elem.RoutineID()).Len() - 1; j >= 0; j-- {
	// 		locksetElement := tr.GetRoutineTrace(elem.RoutineID()).At(j)
	// 		if locksetElement.ResourceID() != prevElement.ResourceID() {
	// 			continue
	// 		}
	// 		if !locksetElement.(*trace.ElementMutex).IsLock() {
	// 			continue
	// 		}
	// 		locksetElements = append(locksetElements, locksetElement)
	// 		break
	// 	}
	// }

	// // If there are any unlocks in the remaining traces, try to ensure that those can happen before we run into the deadlock!
	// for _, relevantRoutineElem := range bug.TraceElement2 {
	// 	routine := relevantRoutineElem.RoutineID()                   // Iterate through all relevant routines
	// 	for _, unlock := range tr.GetRoutineTrace(routine).Elems() { // Iterate through all remaining elements in the routine
	// 		switch unlock := unlock.(type) {
	// 		case *trace.ElementMutex:
	// 			if !(*unlock).IsLock() { // Find Unlock elements
	// 				// Check if the unlocked mutex is in the locksets of the deadlock cycle
	// 				for _, lockElem := range locksetElements {
	// 					// If yes, make sure the unlock happens before the final lock attempts!
	// 					if (*unlock).ResourceID() == lockElem.ResourceID() {
	// 						// Do nothing if the unlock already happens before the lockset element
	// 						if (*unlock).T(trace.Request) < lockElem.T(trace.Request) {
	// 							break
	// 						}

	// 						// Move the as much of the routine of the deadlocking element as possible behind this unlock!
	// 						var concurrentStartElem trace.Element = nil
	// 						for _, possibleStart := range tr.GetRoutineTrace(lockElem.RoutineID()).Elems() {
	// 							if a_clock.GetHappensBefore(possibleStart.GetVC(a_clock.Weak), (*unlock).GetVC(a_clock.Weak)) == a_hb.Concurrent {
	// 								concurrentStartElem = possibleStart
	// 								break
	// 							}
	// 						}

	// 						if concurrentStartElem == nil {
	// 							log.Info("Could not find concurrent element for Routine", lockElem.RoutineID(), "so we cannot move it behind unlock", unlock.ResourceID(), "in Routine", unlock.RoutineID())
	// 							break
	// 						}

	// 						routineEndElem := tr.GetRoutineTrace(lockElem.RoutineID()).Last()
	// 						tr.ShiftRoutine(lockElem.RoutineID(), concurrentStartElem.T(trace.Request), ((*unlock).T(trace.Sorting)-concurrentStartElem.T(trace.Sorting))+1)
	// 						if routineEndElem.T(trace.Commit) > lastTime {
	// 							lastTime = routineEndElem.T(trace.Commit)
	// 						}
	// 						tr.ShiftConcurrentOrAfterToAfter(unlock)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	// for _, elem := range bug.TraceElement2 {
	// 	fmt.Println("Deadlocking Element: ", elem.RoutineID(), "M", elem.T(trace.Request), elem.T(trace.Commit), elem.ResourceID())
	// }

	return &constr, nil
}

// findLastTime returns the latest time stamp from the bug elements
//
// Parameters:
//   - bugElements []trace.TraceElement: the bug element to search through
//
// Returns:
//   - int: the highest tPost from the bug elements
func findLastTime(bugElements []trace.Element) int {
	lastTime := -1

	for _, e := range bugElements {
		if lastTime == -1 || e.T(trace.Sorting) > lastTime {
			lastTime = e.T(trace.Sorting)
		}
	}
	return lastTime
}
