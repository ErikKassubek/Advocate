// Copyright (c) 2025 Erik Kassubek
//
// File: hb.go
// Brief: Function to find concurrent operations by building direct order graph
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_pog

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_helper"
	"advocate/trace"
	"reflect"
)

// GetConcurrent find one or all elements that are concurrent to a given element
//
// Parameter
//   - elem TraceElement: the element the results should be concurrent with
//   - all bool: if true, return all elements that are concurrent, if false, only return one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - weak bool: if true, use weak hb, otherwise strong hb
//
// Returns
//   - []trace.TraceElement: element(s) that are concurrent to elem
func GetConcurrent(elem trace.Element, all bool, sameElem bool, weak bool) []trace.Element {
	reachableFromN := make(map[int]bool)
	reachableToN := make(map[int]bool)

	dfsPartialOrderGraph(elem, nil, reachableFromN, false, weak)
	dfsPartialOrderGraph(elem, nil, reachableToN, true, weak)

	res := make([]trace.Element, 0)

	for rout, routine := range a_base.MainTrace.Routines() {
		if rout == elem.RoutineID() {
			continue
		}

		for _, tElem := range routine.Elems() {
			if sameElem && elem.ResourceID() != tElem.ResourceID() {
				continue
			}

			if !a_helper.Valid(tElem) {
				continue
			}

			if !reachableFromN[tElem.ID()] && !reachableToN[tElem.ID()] {
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
// POG
//
// Parameter:
//   - t1 trace.Element: the trace element
//   - t2 trace.Element: the second element
//   - weak bool: get based on weak happens before
//
// Returns:
//   - happensBefore: The happens before relation between the elements
func GetHappensBefore(t1, t2 trace.Element, weak bool) a_hb.HappensBefore {
	if dfsPartialOrderGraph(t1, t2, nil, false, weak) {
		return a_hb.Before
	}
	if dfsPartialOrderGraph(t2, t1, nil, false, weak) {
		return a_hb.After
	}

	return a_hb.Concurrent
}

// Pass the partial order graph using dfs
// Store all visited nodes
//
// Parameter:
//   - start trace.TraceElement: element to start from
//   - end trace.TraceElement: if not non, stop when end is reached
//   - visited map[int]bool: traceID of all visited nodes
//   - inverted bool: If false, find all nodes that can be reached from start,
//     if true, find all nodes from which start can be reached
//   - weak bool: if true, use weak partial order
//
// Returns:
//   - bool: if end is nil, return if end has been reached, otherwise return true
func dfsPartialOrderGraph(start, end trace.Element, reachable map[int]bool,
	inverted, weak bool) bool {
	if end != nil && start.ID() == end.ID() {
		return true
	}

	if reachable == nil {
		reachable = make(map[int]bool)
	}

	stack := []trace.Element{start}

	var g *PoGraph
	if weak {
		if inverted {
			g = &poWeakInverted
		} else {
			g = &poWeak
		}
	} else {
		if inverted {
			g = &poInverted
		} else {
			g = &po
		}
	}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if reachable[curr.ID()] {
			continue
		}

		reachable[curr.ID()] = true

		if end != nil && start.ID() == end.ID() {
			return true
		}

		for child := range g.GetChildren(curr) {
			if child == nil || reflect.ValueOf(child).IsNil() {
				continue
			}
			if !reachable[child.ID()] {
				stack = append(stack, child)
			}
		}
	}

	return end == nil
}
