// Copyright (c) 2024 Erik Kassubek
//
// File: resources.go
// Brief: Resources
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

// ========================================================
// MARK: Resource
// ========================================================

var shortIDs = make(map[int]int) // long ID -> short ID
var minShortID int = 1

type Resource struct {
	id      int
	shortID int
	alloc   *ElementAlloc
}

func NewResource(id int, alloc *ElementAlloc) Resource {
	if short, ok := shortIDs[id]; ok {
		return Resource{id: id, shortID: short, alloc: alloc}
	}

	shortIDs[id] = minShortID
	minShortID++

	return Resource{id: id, shortID: minShortID, alloc: alloc}
}

func (this *Resource) Id() int {
	return this.id
}

func (this *Resource) Alloc() *ElementAlloc {
	return this.alloc
}
