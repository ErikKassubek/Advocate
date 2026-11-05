// advocate/fuzzing/active/rewriteMixedDeadlock.go
//
// Copyright (c) 2024 Erik Kassubek
//
// File: rewriteMixedDeadlock.go
//
// License: BSD-3-Clause

// TODO: rewrite with constraints and closure

package f_active

import (
	"advocate/fuzzing/f_constraints"
	"advocate/trace"
	"advocate/utils/results/bugs"
	"errors"
	"fmt"
)

// TODO: rewrite

func rewriteMixedDeadlock(bug bugs.Bug) (*f_constraints.Constraint, error) {
	if len(bug.TraceElement2) != 4 {
		return nil, fmt.Errorf("rewriteMixedDeadlock: expected 4 elements, got %d", len(bug.TraceElement2))
	}

	chHolder, ok1 := bug.TraceElement2[0].(*trace.ElementChannel)
	lockHolder, ok2 := bug.TraceElement2[1].(*trace.ElementMutex)
	chWaiter, ok3 := bug.TraceElement2[2].(*trace.ElementChannel)
	lockWaiter, ok4 := bug.TraceElement2[3].(*trace.ElementMutex)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, errors.New("rewriteMixedDeadlock: incorrect element types in cycle")
	}

	constr := f_constraints.NewConstraint()

	if chHolder.T(trace.Commit) < lockHolder.T(trace.Commit) {
		constr.Add(chHolder)
		constr.Add(lockHolder)
	} else {
		constr.Add(lockHolder)
		constr.Add(chHolder)
	}

	if chWaiter.T(trace.Commit) < lockWaiter.T(trace.Commit) {
		constr.Add(chWaiter)
		constr.Add(lockWaiter)
	} else {
		constr.Add(lockWaiter)
		constr.Add(chWaiter)
	}

	return &constr, nil
}

// func rewriteMixedDeadlock(bug bugs.Bug, code int) error {
// 	if len(bug.TraceElement2) != 4 {
// 		return fmt.Errorf("rewriteMixedDeadlock: expected 4 elements, got %d", len(bug.TraceElement2))
// 	}

// 	tr := a_base.MainTrace

// 	cdHolder, ok1 := bug.TraceElement2[0].(*trace.ElementChannel)
// 	lockHolder, ok2 := bug.TraceElement2[1].(*trace.ElementMutex)
// 	_, ok3 := bug.TraceElement2[2].(*trace.ElementChannel)
// 	lockWaiter, ok4 := bug.TraceElement2[3].(*trace.ElementMutex)

// 	if !ok1 || !ok2 || !ok3 || !ok4 {
// 		return errors.New("rewriteMixedDeadlock: incorrect element types in cycle")
// 	}

// 	holderRout := cdHolder.RoutineID()
// 	waiterRout := lockWaiter.RoutineID()

// 	mainRout := 0
// 	for rid := range tr.GetTraces() {
// 		if rid != int(holderRout) && rid != int(waiterRout) {
// 			if mainRout == 0 || rid < mainRout {
// 				mainRout = rid
// 			}
// 		}
// 	}
// 	if mainRout == 0 {
// 		mainRout = 1
// 	}

// 	//fmt.Printf("rewriteMixedDeadlock: main=R%d, holder=R%d (lock=%d, chan=%d), waiter=R%d (lock=%d)\n",
// 	//	mainRout, holderRout, lockHolder.GetT(tPre, ), cdHolder.GetT(tPre, ), waiterRout, lockWaiter.GetT(tPre, ))

// 	lastTime := max(lockHolder.T(trace.Commit), lockWaiter.T(trace.Commit))
// 	if mainTrace := tr.GetRoutineTrace(mainRout); mainTrace.Len() > 0 {
// 		if lastElem := mainTrace.Last(); lastElem.T(trace.Commit) > lastTime {
// 			lastTime = lastElem.T(trace.Commit)
// 		}
// 	}
// 	tr.ShortenTrace(lastTime, true)

// 	tr.ShortenRoutine(holderRout, cdHolder.T(trace.Commit)+1)
// 	tr.ShortenRoutine(waiterRout, lockWaiter.T(trace.Commit)+1)

// 	//fmt.Printf("rewriteMixedDeadlock: holder R%d kept to t=%d, waiter R%d kept to t=%d\n",
// 	//	holderRout, cdHolder.GetT(tPost, ), waiterRout, lockWaiter.GetT(tPost, ))

// 	// Reorder
// 	if lockWaiter.T(trace.Request) < lockHolder.T(trace.Request) {
// 		targetTPre := lockHolder.T(trace.Request) + 1
// 		shift := targetTPre - lockWaiter.T(trace.Request)
// 		if shift > 0 {
// 			waiterTrace := tr.GetRoutineTrace(waiterRout)
// 			if waiterTrace.Empty() {
// 				return fmt.Errorf("rewriteMixedDeadlock: waiter R%d has no trace", waiterRout)
// 			}
// 			startElem := waiterTrace.First()
// 			startTPre := startElem.T(trace.Request)
// 			//fmt.Printf("rewriteMixedDeadlock: shifting waiter R%d by %d\n", waiterRout, shift)
// 			tr.ShiftRoutine(waiterRout, startTPre, shift)
// 			tr.ShiftConcurrentOrAfterToAfter(lockWaiter)
// 		}
// 	}

// 	// Ensure holder's channel op is after lock acquire
// 	if cdHolder.T(trace.Request) <= lockHolder.T(trace.Request) {
// 		newTPre := lockHolder.T(trace.Commit) + 1
// 		cdHolder.SetT(trace.Request, newTPre)
// 		cdHolder.SetT(trace.Commit, newTPre)
// 	}

// 	// Clear channel state
// 	forceChannelBlock(cdHolder)

// 	// Calculate final time
// 	newLastTime := 0
// 	for _, routSlice := range tr.GetTraces() {
// 		for _, elem := range routSlice.Elems() {
// 			t := elem.T(trace.Sorting)
// 			if t > newLastTime && t != math.MaxInt {
// 				newLastTime = t
// 			}
// 		}
// 		//fmt.Printf("rewriteMixedDeadlock: R%d has %d elements, last t=%d\n",
// 		//	rid, len(traceSlice), newLastTime)
// 	}

// 	// Use SetTWithoutNotExecuted to set tPost=0 while preserving tPre behavior
// 	blockElement(cdHolder)
// 	blockElement(lockWaiter)

// 	return nil
// }

// blockElement to force an element to block (tPost=0) while preserving tPre
