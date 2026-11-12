// Copyright (c) 2025 
//
// File: graph.go
// Brief: Store the partial order graph
//
//
// License: BSD-3-Clause

package a_pog

import (
	"gosc/analysis/a_base"
	"gosc/trace"
	"gosc/utils/log"
	"gosc/utils/types"
	"fmt"
)

var (
	po         PoGraph
	poInverted PoGraph

	poWeak         PoGraph
	poWeakInverted PoGraph
)

type PoGraph struct {
	data       map[trace.Element]map[trace.Element]struct{}
	DataSimple map[int]map[int]struct{}

	lastAdded map[int]trace.Element

	lastAtomicWriter map[int]*trace.ElementAtomic
	chanBuffer       map[int]([]a_base.BufferedVC)
	chanBufferSize   map[int]int
	closeData        map[int]trace.Element
	curWaitingCond   map[int]*types.Queue[*trace.ElementCond]
	relR             map[int]*a_base.ElemWithVc
	relW             map[int]*a_base.ElemWithVc
	oSuc             map[int]*trace.ElementOnce
	lastChangeWg     map[int]*trace.ElementWait
	ForkOps          map[int]*trace.ElementFork
}

func NewPoGraph() PoGraph {
	po.chanBuffer = make(map[int]([]a_base.BufferedVC))
	po.chanBufferSize = make(map[int]int)

	return PoGraph{
		data:           make(map[trace.Element]map[trace.Element]struct{}),
		DataSimple:     make(map[int]map[int]struct{}),
		lastAdded:      make(map[int]trace.Element),
		chanBuffer:     make(map[int][]a_base.BufferedVC),
		chanBufferSize: make(map[int]int),
		closeData:      make(map[int]trace.Element),
		curWaitingCond: make(map[int]*types.Queue[*trace.ElementCond]),
		relR:           make(map[int]*a_base.ElemWithVc),
		relW:           make(map[int]*a_base.ElemWithVc),
		oSuc:           make(map[int]*trace.ElementOnce),
		lastChangeWg:   make(map[int]*trace.ElementWait),
		ForkOps:        make(map[int]*trace.ElementFork),
	}
}

// InitPOG initializes the directed acyclic partial order graph
func InitPOG() {

	po = NewPoGraph()
	poInverted = NewPoGraph()

	poWeak = NewPoGraph()
	poWeakInverted = NewPoGraph()
}

func (this *PoGraph) AddEdge(from, to trace.Element) {
	if _, ok := this.data[from]; !ok {
		this.data[from] = make(map[trace.Element]struct{})
	}
	if _, ok := this.DataSimple[from.ID()]; !ok {
		this.DataSimple[from.ID()] = make(map[int]struct{})
	}
	this.data[from][to] = struct{}{}
	this.DataSimple[from.ID()][to.ID()] = struct{}{}
}

func (this *PoGraph) GetChildren(from trace.Element) map[trace.Element]struct{} {
	return this.data[from]
}

func (this *PoGraph) GetChildrenSimple(from int) map[int]struct{} {
	return this.DataSimple[from]
}

func (this *PoGraph) IsEmpty() bool {
	return len(this.data) == 0
}

// Print prints the current graph
//
// Parameter:
//   - weak bool: if true, print the weak hb tree, otherwise print the string
func Print(weak bool) {
	if weak {
		log.Info(poWeak.toString())
	} else {
		log.Info(po.toString())
	}
}

func (this *PoGraph) toString() string {
	res := ""
	for start, end := range this.data {
		res += fmt.Sprintf("%s -> ", start.Pos())
		for e := range end {
			res += fmt.Sprintf("%s, ", e.Pos())
		}
	}
	return res
}

func (this *PoGraph) toStringSimple() string {
	res := ""
	for start, end := range this.DataSimple {
		res += fmt.Sprintf("%d -> ", start)
		for e := range end {
			res += fmt.Sprintf("%d, ", e)
		}
	}
	return res
}

// AddEdge adds an edge between start and end in po
//
// Parameter:
//   - start trace.Element: the start element
//   - end trace.Element: the end element
//   - notWeak bool: if true, add to weak happens before
func AddEdge(start, end trace.Element, weak bool) {
	if start == nil || end == nil {
		return
	}

	po.AddEdge(start, end)
	poInverted.AddEdge(end, start)

	if weak {
		poWeak.AddEdge(start, end)
		poWeakInverted.AddEdge(end, start)
	}
}
