// Copyright (c) 2025 
//
// File: addEdges.go
// Brief: Add edges to the graph
//
//
// License: BSD-3-Clause

package a_pog

import (
	"gosc/analysis/a_base"
	"gosc/analysis/hb/a_helper"
	"gosc/trace"
)

// AddEdgeSameRoutineAndFork adds a given element to the children of the last element that
// was analyzed in the same routine.
// Then set this element to be the last element analyzed in the routine
// If it is the first element in a routine, add an edge to the corresponding fork
//
// Parameter:
//   - graph PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - elem trace.Element: the element to add an edge for
func AddEdgeSameRoutineAndFork(graph *PoGraph, elem trace.Element) {
	if !a_helper.Valid(elem) {
		return
	}
	routineID := elem.RoutineID()

	gr := graph
	if graph == nil {
		gr = &po
	}

	forks := a_base.ForkOperations
	if graph != nil {
		forks = gr.ForkOps
	}

	if lastElem, ok := gr.lastAdded[routineID]; ok {
		if graph != nil {
			graph.AddEdge(lastElem, elem)
		} else {
			AddEdge(lastElem, elem, true)
		}
	} else {
		// first element, add edge from fork if exists
		if fork, okF := forks[routineID]; okF {
			if graph != nil {
				graph.AddEdge(fork, elem)
			} else {
				AddEdge(fork, elem, true)
			}
		}
	}
	gr.lastAdded[routineID] = elem
}
