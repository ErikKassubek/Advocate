// Copyright (c) 2025 
//
// File: pair.go
// Brief: Implement a set data type
//
//
// License: BSD-3-Clause

package types

// Set represents an unordered set with unique values
type Set[T comparable] struct {
	data map[T]struct{}
}

// NewSet creates a new, empty set
//
// Returns:
//   - Set[T]: the new set
func NewSet[T comparable]() Set[T] {
	return Set[T]{data: make(map[T]struct{})}
}

// Add adds an element to the set
//
// Parameter:
//   - e T: the element to add
func (this *Set[T]) Add(e T) {
	this.data[e] = struct{}{}
}

// Remove removes an element from the set
// If the set does not contain the element, Remove is a no-op
//
// Parameter:
//   - e T: the element to remove
func (this *Set[T]) Remove(e T) {
	delete(this.data, e)
}

// Contains returns if the set contains a given element
//
// Parameter:
//   - e T: the element to check
//
// Returns:
//   - bool: true if the e is in the set, false otherwise
func (this *Set[T]) Contains(e T) bool {
	_, ok := this.data[e]
	return ok
}
