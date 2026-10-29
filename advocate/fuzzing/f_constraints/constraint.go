// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-sc.go
// Brief: scheduling Chains for GoPie
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_constraints

import (
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_hbcalc"
	"advocate/trace"
	"fmt"
	"sort"
)

// Constraint is a representation of a scheduling Constraint
// A Constraint is an ordered list of adjacent element from the trace,
// where two neighboring elements must be from different routines
type Constraint struct {
	Elems []trace.Element
}

// NewConstraint create a new, empty chain
//
// Returns: chain: the new chain
func NewConstraint() Constraint {
	elems := make([]trace.Element, 0)
	return Constraint{elems}
}

type ElemWithQual struct {
	Elem    trace.Element
	Quality float64
}

// Add a new element to the chain
//
// Parameter:
//   - elems ...analysis.TraceElement: Elements to Add, in the order they are added
func (this *Constraint) Add(elems ...trace.Element) {
	if elems == nil {
		return
	}

	for _, elem := range elems {
		if elem == nil {
			continue
		}
		this.Elems = append(this.Elems, elem)
	}
}

// Replace replaces the element at a given index in a chain with another element
//
// Parameter:
//   - index int: index to change at
//   - elem analysis.TraceElement: element to set at index
func (this *Constraint) Replace(index int, elem trace.Element) {
	if elem == nil {
		return
	}

	if index < 0 || index >= len(this.Elems) {
		return
	}
	this.Elems[index] = elem
}

// Returns if the chain Contains a specific element
//
// Parameter:
//   - elem analysis.TraceElement: the element to check for
//
// Returns:
//   - bool: true if the chain Contains elem, false otherwise
func (this *Constraint) Contains(elem trace.Element) bool {
	if elem == nil {
		return false
	}

	for _, c := range this.Elems {
		if elem.IsEqual(c) {
			return true
		}
	}

	return false
}

// Remove the first element from the chain
func (this *Constraint) RemoveHead() {
	this.Elems = this.Elems[1:]
}

// Remove the last element from the chain
func (this *Constraint) RemoveTail() {
	this.Elems = this.Elems[:len(this.Elems)-1]
}

// Return the first element of a chain
//
// Returns:
//   - analysis.TraceElement: the first element in the chain, or nil if chain is empty
func (this *Constraint) ElemWithSmallestTPost() trace.Element {
	if this.Len() == 0 {
		return nil
	}

	var min trace.Element

	for _, c := range this.Elems {
		if min == nil || c.T(trace.Sorting) < min.T(trace.Sorting) {
			min = c
		}
	}

	return min
}

// Return the last element of a chain
//
// Returns:
//   - analysis.TraceElement: the last element in the chain, or nil if chain is empty
func (this *Constraint) LastElem() trace.Element {
	if this.Len() == 0 {
		return nil
	}
	return this.Elems[len(this.Elems)-1]
}

// Swap the two elements in the chain given by the indexes.
// If at least on index is not in the chain, nothing is done
//
// Parameter:
//   - i int: index of the first element
//   - j int: index of the second element
//
// Return:
//   - bool: true if Swap was possible, false otherwise
func (this *Constraint) Swap(i, j int) bool {
	if a_hbcalc.GetHappensBefore(this.Elems[i], this.Elems[j], true) != a_hb.Concurrent {
		return false
	}
	if i >= 0 && i < len(this.Elems) && j >= 0 && j < len(this.Elems) {
		this.Elems[i], this.Elems[j] = this.Elems[j], this.Elems[i]
	}
	return true
}

// Create a Copy of the chain
//
// Returns:
//   - chain: a Copy of the chain
func (this *Constraint) Copy() Constraint {
	newElems := make([]trace.Element, len(this.Elems))

	copy(newElems, this.Elems)

	newChain := Constraint{
		Elems: newElems,
	}
	return newChain
}

// Len returns the number of elements in a scheduling chain
//
// Returns:
//   - the number of elements in the chain
func (this *Constraint) Len() int {
	if this == nil {
		return 0
	}
	return len(this.Elems)
}

// Get a string representation of a scheduling chain
//
// Returns:
//   - A string representation of the chain
func (this *Constraint) String() string {
	res := ""
	for _, e := range this.Elems {
		res += fmt.Sprintf("%d:%s", e.RoutineID(), e.Pos())
		switch f := e.(type) {
		case *trace.ElementSelect:
			res += fmt.Sprintf("%d", f.GetChosenIndex())
		}
		res += "&"
	}
	return res
}

// Get a string representation of a scheduling chain
//
// Returns:
//   - A string representation of the chain
func (this *Constraint) StringDebug() string {
	res := ""
	for _, e := range this.Elems {
		res += e.StringDebug()
		res += "\n"
	}
	return res
}

// Check if a chain is valid.
// A chain is valid if it isn't violation the HB relation
// If the analyzer did not run and therefore did not calculate the HB relation,
// the function will always return true
// Since HB relations are transitive, it is enough to check neighboring elements
//
// Returns:
//   - bool: True if the mutation is valid, false otherwise
func (this *Constraint) IsValid() bool {
	for i := 0; i < len(this.Elems)-1; i++ {
		for j := 1; j < len(this.Elems); j++ {
			hbInfo := a_hbcalc.GetHappensBefore(this.Elems[i], this.Elems[j], true)
			if hbInfo == a_hb.After {
				return false
			}
		}
	}

	return true
}

// Get the map of all elements in the constraint
//
// Returns:
//
//	map[trace.Element]struct{}: map containing all elements in the constraint as keys
func (this *Constraint) toMap() map[trace.Element]struct{} {
	res := make(map[trace.Element]struct{})
	for _, elem := range this.Elems {
		res[elem] = struct{}{}
	}
	return res
}

// Sort the Elements by commit time
func (this *Constraint) sort() {
	sort.Slice(this.Elems, func(i, j int) bool {
		return this.Elems[i].T(trace.Commit) < this.Elems[j].T(trace.Commit)
	})
}

// MergeClosureConstraint two constraints
func MergeClosureConstraint(closure, constr *Constraint) *Constraint {
	new := NewConstraint()

	for _, e := range closure.Elems {
		new.Add(e)
	}

	for _, e := range constr.Elems {
		new.Add(e)
	}

	// must ensure, that inter routine order is keept
	// for now we do this by removing the additional values from closure
	for _, elem1 := range constr.Elems {
		rConst := elem1.Routine()
		tConst := elem1.T(trace.Commit)
		remove := make([]int, 0, len(closure.Elems))
		for i, elem2 := range closure.Elems {
			rClosure := elem2.Routine()
			tClosure := elem2.T(trace.Commit)

			if rConst == rClosure && tClosure > tConst {
				remove = append(remove, i)
			}
		}

		closure.Elems = removeIndices(closure.Elems, remove)
	}

	return &new
}

func removeIndices(s []trace.Element, d []int) []trace.Element {
	j := 0
	for i := range s {
		if j < len(d) && i == d[j] {
			j++
			continue
		}
		s[i-j] = s[i]
	}
	return s[:len(s)-len(d)]
}
