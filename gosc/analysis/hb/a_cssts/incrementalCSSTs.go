// Copyright (c) 2025 
//
// File: incrementalCSSTs.go
// Brief: Implementation of IncrementalCSSTs.java from CSST paper (see csst.go)
//
//
// License: BSD-3-Clause

// CSST is based on
// Hünkar Can Tunç, Ameya Prashant Deshmukh, Berk Cirisci, Constantin Enea,
// and Andreas Pavlogiannis. 2024. CSSTs: A Dynamic Data Structure for Partial
// Orders in Concurrent Execution Analysis. In Proceedings of the 29th ACM
// International Conference on Architectural Support for Programming Languages
// and Operating Systems, Volume 3 (ASPLOS '24), Vol. 3. Association for
// Computing Machinery, New York, NY, USA, 223–238.
// https://doi.org/10.1145/3620666.3651358

package a_cssts

import (
	"gosc/utils/types"
	"math"
)

// IncrementalCSST implements a Collective Sparse Segment Trees
type IncrementalCSST struct {
	width   int
	lengths []int
	ssts    []map[int]sparseSegmentTree
}

// NewIncrementalCSST creates a new Collective Sparse Segment Trees
//
// Parameter:
//   - lengths []int: number of elements for each routine
//
// Returns:
//   - IncrementalCSST: the new tree
func NewIncrementalCSST(lengths []int) IncrementalCSST {
	iCSST := IncrementalCSST{
		len(lengths), lengths, make([]map[int]sparseSegmentTree, 0),
	}

	for i := 0; i < iCSST.width; i++ {
		iCSST.ssts = append(iCSST.ssts, make(map[int]sparseSegmentTree))
		for j := 0; j < iCSST.width; j++ {
			if i != j {
				if iCSST.lengths[i] == 0 {
					iCSST.ssts[i][j] = newSparseSegmentTree1(1)
				} else {
					iCSST.ssts[i][j] = newSparseSegmentTree1(iCSST.lengths[i])
				}
			}
		}
	}

	return iCSST
}

func (this *IncrementalCSST) getSuccessor1(p types.Pair[int, int]) []int {
	s := make([]int, this.width)
	for i := 0; i < this.width; i++ {
		s[i] = this.getSuccessor2(p, i)
	}

	return s
}

func (this *IncrementalCSST) getSuccessor2(p types.Pair[int, int], i int) int {
	if p.X == i {
		if p.Y < this.lengths[p.X]-1 {
			return p.Y + 1
		}
		return -1
	}

	v := this.ssts[p.X][i].sumRange1(p.Y)
	if v < math.MaxInt {
		return v
	}
	return -1
}

func (this *IncrementalCSST) getPredecessor1(p types.Pair[int, int]) []int {
	s := make([]int, this.width)
	for i := 0; i < this.width; i++ {
		s[i] = this.getPredecessor2(p, i)
	}

	return s
}

func (this *IncrementalCSST) getPredecessor2(p types.Pair[int, int], i int) int {
	if p.X == i {
		if p.Y > 0 {
			return p.Y - 1
		}
		return -1
	}

	v := this.ssts[i][p.X].argMin1(p.Y)
	if v > math.MinInt {
		return v
	}
	return -1
}

func (this *IncrementalCSST) reachable(from, to types.Pair[int, int]) bool {
	if from.X == to.X && from.Y == to.Y {
		return true
	}

	v := this.getSuccessor2(from, to.X)
	return v >= 0 && v <= to.Y
}

func (this *IncrementalCSST) addSuccessor(from, to types.Pair[int, int]) {
	if from.X != to.X {
		this.ssts[from.X][to.X].update1(from.Y, to.Y)
	}
}

// AddEdge adds a new edge to the tree
//
// Parameter:
//   - from types.Pair[int, int]: indices of the start node
//   - from types.Pair[int, int]: indices of the end node
//
// Returns:
//   - types.Set[types.Pair[types.Pair[int, int], types.Pair[int, int]: the added edge
func (this *IncrementalCSST) AddEdge(from, to types.Pair[int, int]) types.Set[types.Pair[types.Pair[int, int], types.Pair[int, int]]] {
	addedEdges := types.NewSet[types.Pair[types.Pair[int, int], types.Pair[int, int]]]()

	if this.reachable(from, to) {
		return addedEdges
	}

	q := types.Stack[types.Pair[types.Pair[int, int], types.Pair[int, int]]]{}

	q.Push(types.Pair[types.Pair[int, int], types.Pair[int, int]]{X: from, Y: to})

	for !q.IsEmpty() {
		e := q.Pop()
		f := e.X
		t := e.Y

		if this.reachable(f, t) {
			continue
		}

		this.addSuccessor(f, t)
		addedEdges.Add(e)

		suc := this.getSuccessor1(t)
		pred := this.getPredecessor1(f)

		for i := 0; i < this.width; i++ {
			if i != f.X && i != t.X {
				tt := types.NewPair(i, suc[i])
				if suc[i] >= 0 {
					q.Push(types.NewPair(f, tt))
				}
				ff := types.NewPair(i, pred[i])
				if pred[i] >= 0 {
					q.Push(types.NewPair(ff, t))
				}
			}
		}
	}

	return addedEdges
}

func (this *IncrementalCSST) getChainLength(i int) int {
	return this.lengths[i]
}
