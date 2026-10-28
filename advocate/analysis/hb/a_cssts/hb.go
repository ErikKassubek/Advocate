// Copyright (c) 2025 Erik Kassubek
//
// File: concurrent.go
// Brief: Get concurrency information
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_cssts

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_helper"
	"advocate/trace"
	"advocate/utils/types"
)

// GetConcurrentAllPairs return concurrent events for a given event based on the CSST
// Parameter:
//   - elem trace.TraceElem: the element to search for
//   - all bool: if true, return all concurrent events, otherwise return one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - weak bool: get based on weak happens before
//
// Returns:
//   - []trace.TraceElement: the concurrent element(s)
func GetConcurrentAllPairs(elem trace.Element, all, sameElem, weak bool) []trace.Element {
	res := make([]trace.Element, 0)
	id := elem.ResourceID()
	routId := elem.RoutineID()

	for r, routine := range a_base.MainTrace.Routines() {
		// same routine
		if routId == r {
			continue
		}

		// different routine
		for _, e := range routine.Elems() {
			if sameElem && e.ResourceID() != id {
				continue
			}

			if !a_helper.Valid(e) {
				continue
			}

			if isConcurrent(elem, e, weak) {
				res = append(res, e)
				if !all {
					return res
				}
			}
		}
	}

	return res
}

// Given two trace elements, return if the elements are concurrent
//
// Parameter:
//   - elem1 trace.Element: the first elem
//   - elem2 trace.Element: the second elem
//   - weak bool: get based on weak happens before
//
// Returns:
//   - bool: true if the elements are concurrent, false otherwise
func isConcurrent(elem1, elem2 trace.Element, weak bool) bool {
	hbInfo := GetHappensBefore(elem1, elem2, weak)
	return hbInfo == a_hb.Concurrent
}

// GetConcurrent find one or all elements that are concurrent to a given event
//
// Parameter
//   - elem TraceElement: the element the results should be concurrent with
//   - all bool: if true, return all elements that are concurrent, if false, only return one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - weak bool: get based on weak happens before
//
// Returns
//   - []trace.TraceElement: element(s) that are concurrent to elem
func GetConcurrent(elem trace.Element, all, sameElem, weak bool) []trace.Element {
	reachableFromN := types.NewSet[types.Pair[int, int]]()
	reachableToN := types.NewSet[types.Pair[int, int]]()

	rout, ind := elem.TraceIndex()
	elemInd := types.NewPair(rout, ind)

	dfsCSST(elemInd, reachableFromN, weak, false)
	dfsCSST(elemInd, reachableToN, weak, true)

	res := make([]trace.Element, 0)

	for routID, routine := range a_base.MainTrace.Routines() {
		if routID == elem.RoutineID() {
			continue
		}

		for _, tElem := range routine.Elems() {
			if sameElem && elem.ResourceID() != tElem.ResourceID() {
				continue
			}

			if !a_helper.Valid(tElem) {
				continue
			}

			r, i := tElem.TraceIndex()
			tElemInd := types.NewPair(r, i)

			if !reachableFromN.Contains(tElemInd) && !reachableToN.Contains(tElemInd) {
				res = append(res, tElem)
				if !all {
					return res
				}
			}
		}
	}
	return res
}

// GetHappensBefore returns the happens before relation between two operations given there
// csst
//
// Parameter:
//   - t1 trace.Element: the trace element
//   - t2 trace.Element: the second element
//   - weak bool: get based on weak happens before
//
// Returns:
//   - happensBefore: The happens before relation between the elements
func GetHappensBefore(t1, t2 trace.Element, weak bool) a_hb.HappensBefore {
	t1Ind := getIndicesFromTraceElem(t1)
	t2Ind := getIndicesFromTraceElem(t2)

	if weak {
		if CsstWeak.reachable(t1Ind, t2Ind) {
			return a_hb.Before
		}
		if CsstWeak.reachable(t2Ind, t1Ind) {
			return a_hb.After
		}
		return a_hb.Concurrent
	}

	if Csst.reachable(t1Ind, t2Ind) {
		return a_hb.Before
	}
	if Csst.reachable(t2Ind, t1Ind) {
		return a_hb.After
	}
	return a_hb.Concurrent
}

// Pass the partial order graph using dfs
// Store all visited nodes
//
// Parameter:
//   - start trace.TraceElement: element to start from
//   - reachable map[int]bool: traceID of all visited nodes
//   - weak bool: get based on weak happens before
//   - inverted bool: If false, find all nodes that can be reached from start,
//     if true, find all nodes from which start can be reached
func dfsCSST(from types.Pair[int, int], reachable types.Set[types.Pair[int, int]], weak, inverted bool) types.Set[types.Pair[int, int]] {
	visited := types.NewSet[types.Pair[int, int]]()

	queue := []types.Pair[int, int]{from}

	st := &Csst
	stInverted := &CsstInverted
	if weak {
		st = &CsstWeak
		stInverted = &CsstWeakInverted
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited.Contains(current) {
			continue
		}
		visited.Add(current)

		// Skip adding the origin to the reachable set if needed
		if !(current.X == from.X && current.Y == from.Y) {
			reachable.Add(current)
		}

		// Traverse next event on the same chain
		if inverted {
			if current.Y-1 >= 0 {

				next := types.NewPair(current.X, current.Y-1)
				queue = append(queue, next)
			}
		} else {
			if current.Y+1 < st.getChainLength(current.X) {

				next := types.NewPair(current.X, current.Y+1)
				queue = append(queue, next)
			}
		}

		// Traverse all cross-chain successors
		sucs := make([]int, 0)
		if inverted {
			sucs = stInverted.getSuccessor1(current)
		} else {
			sucs = st.getSuccessor1(current)
		}
		for targetChain, sucY := range sucs {
			if targetChain == current.X || sucY == -1 {
				continue
			}
			suc := types.NewPair(targetChain, sucY)
			queue = append(queue, suc)
		}
	}

	return reachable
}
