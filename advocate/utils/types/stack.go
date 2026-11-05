// Copyright (c) 2025 Erik Kassubek
//
// File: pair.go
// Brief: Implement a stack data type
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package types

// Stack implements a stack
type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0)}
}

// Push adds an item to the top
//
// Parameter:
//   - item T: the item to add
func (this *Stack[T]) Push(item T) {
	this.items = append(this.items, item)
}

// Pop removes and returns the top item
//
// Returns:
//   - T: the removed element
func (this *Stack[T]) Pop() T {
	if len(this.items) == 0 {
		var zero T
		return zero // empty stack
	}
	top := this.items[len(this.items)-1]
	this.items = this.items[:len(this.items)-1]
	return top
}

// Peek returns the top item without removing it
//
// Returns:
//   - T: the top element
func (this *Stack[T]) Peek() T {
	if len(this.items) == 0 {
		var zero T
		return zero
	}
	return this.items[len(this.items)-1]
}

// IsEmpty returns whether the stack is empty
//
// Returns:
//   - bool: true if the stack is empty, false otherwise
func (this *Stack[T]) IsEmpty() bool {
	return len(this.items) == 0
}

// Size returns the number of items
//
// Returns:
//   - int: the number of elements in the stack
func (this *Stack[T]) Size() int {
	return len(this.items)
}

// AsArray returns the array of the stack
// The first element is the bottom element in the stack
//
// Returns:
//   - []T: The stack as an array
func (this *Stack[T]) AsArray() []T {
	return this.items
}
