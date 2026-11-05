// Copyright (c) 2024 Erik Kassubek
//
// File: routine.go
// Brief: Functions and structs for one routine
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"sort"
)

// ========================================================
// MARK: Routine
// ========================================================

type Routine struct {
	id        int
	fork      *ElementFork
	elems     []Element
	resources []Resource // for resource aware blocking bug detection
}

func NewRoutine(id int) *Routine {
	return &Routine{id, nil, make([]Element, 0), make([]Resource, 0)}
}

func (this *Routine) addElement(elem Element) {
	this.elems = append(this.elems, elem)
}

func (this *Routine) setFork(fork *ElementFork) {
	this.fork = fork
}

// ========================================================
// Properties
// ========================================================

func (this *Routine) ID() int {
	return this.id
}

func (this *Routine) Sort() {
	sort.Sort(sortByTSort(this.elems))
}

func (this *Routine) size() (int, int) {
	return cap(this.elems), len(this.elems)
}

func (this *Routine) Len() int {
	return len(this.elems)
}

func (this *Routine) Empty() bool {
	return len(this.elems) == 0
}

func (this *Routine) isBlocked() bool {
	if this.Empty() {
		return false
	}

	return !this.Last().Committed()
}

func (this *Routine) IsTerminated() bool {
	if this.Empty() {
		return false
	}

	switch this.Last().(type) {
	case *ElementRoutineEnd:
		return true
	}

	return false
}

// ========================================================
// Elements
// ========================================================

func (this *Routine) Fork() *ElementFork {
	return this.fork
}

func (this *Routine) Elems() []Element {
	return this.elems
}

func (this *Routine) At(index int) Element {
	return this.elems[index]
}

func (this *Routine) Last() Element {
	if this.Empty() {
		return nil
	}

	return this.elems[len(this.elems)-1]
}

func (this *Routine) First() Element {
	if this.Empty() {
		return nil
	}

	return this.elems[0]
}

func (this *Routine) Resources() []Resource {
	return this.resources
}

// ========================================================
// Modify
// ========================================================

func (this *Routine) shortenIndex(index int) {
	this.elems = this.elems[:index]
}

func (this *Routine) removeAtIndex(index int) {
	this.elems = append(this.elems[:index], this.elems[index+1:]...)
}

func (this *Routine) shortenTime(time int) {
	for index, elem := range this.elems {
		if elem.T(Sorting) >= time {
			this.shortenIndex(index)
			break
		}
	}
}

func (this *Routine) shift(startTSort, shift int) {
	for index, elem := range this.elems {
		if elem.T(Request) >= startTSort {
			this.elems[index].SetTWithoutNotExecuted(elem.T(Sorting) + shift)
		}
	}
}

func (this *Routine) SetTSortAtIndex(tPost, index int) {
	if this.Len() <= index {
		return
	}
	this.At(index).SetT(Sorting, tPost)
}

// ========================================================
// Resource Aware
// ========================================================

func (this *Routine) addResource(res Resource) {
	this.resources = append(this.resources, res)
}
