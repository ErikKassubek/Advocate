// Copyright (c) 2025 
//
// File: addEdges.go
// Brief: Functions to add the required edges
//
//
// License: BSD-3-Clause

package a_cssts

import (
	"gosc/analysis/a_base"
	"gosc/trace"
	"gosc/utils/types"
)

// AddEdge adds an edge to the cssts
//
// Parameter:
//   - from trace.Element: start node
//   - to trace.Element: end node
//   - weak bool: if true, add to weak hb
func AddEdge(from, to trace.Element, weak bool) {
	if from == nil || to == nil {
		return
	}

	fromInd := getIndicesFromTraceElem(from)
	toInd := getIndicesFromTraceElem(to)
	Csst.AddEdge(fromInd, toInd)
	CsstInverted.AddEdge(toInd, fromInd)
	if weak {
		CsstWeak.AddEdge(fromInd, toInd)
		CsstWeakInverted.AddEdge(toInd, fromInd)
	}
}

// Add an edge to the cssts
//
// Parameter:
//   - from trace.Element: start node
//   - to trace.Element: end node
//   - weak bool: if true, add to weak hb
func addEdgeIndex(from, to types.Pair[int, int], weak bool) {
	Csst.AddEdge(from, to)
	CsstInverted.AddEdge(to, from)

	if weak {
		CsstWeak.AddEdge(from, to)
		CsstWeakInverted.AddEdge(to, from)
	}
}

// AddEdgeFork adds an edge between a fork element and the first element in the newly
// crated routine
//
// Parameter:
//   - elem *trace.ElementFork: the fork element
func AddEdgeFork(elem *trace.ElementFork) {
	routine, index := elem.TraceIndex()
	newRout := elem.ResourceID()
	if a_base.GetTraceLength(newRout) > 0 {
		addEdgeIndex(
			types.Pair[int, int]{X: routine, Y: index},
			types.Pair[int, int]{X: newRout, Y: 0}, true)
	}
}
