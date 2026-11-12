// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the pog for once
//
//
// License: BSD-3-Clause

package a_pog

import (
	"gosc/trace"
)

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(graph *PoGraph, on *trace.ElementOnce) {
	gr := graph
	if graph == nil {
		gr = &po
	}

	objId := on.ResourceID()

	// suc once does not create edge -> only not suc
	if on.GetSuc() {
		gr.oSuc[objId] = on
	} else {
		suc := gr.oSuc[objId]
		if graph != nil {
			graph.AddEdge(suc, on)
		} else {
			AddEdge(suc, on, false)
		}
	}
}
