// Copyright (c) 2025 
//
// File: queue.go
// Brief: Implementation of a fifo queue
//
//
// License: BSD-3-Clause

package types

// Queue implements a fifo queue
type Queue[T any] struct {
	items []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{items: make([]T, 0)}
}

// Push adds an item to the top
//
// Parameter:
//   - item T: the item to add
func (this *Queue[T]) Push(item T) {
	this.items = append(this.items, item)
}

// Pop removes and returns the top item
//
// Returns:
//   - T: the removed element
func (this *Queue[T]) Pop() T {
	if len(this.items) == 0 {
		var zero T
		return zero // empty stack
	}
	first := this.items[0]
	this.items = this.items[1:]

	return first
}

// Peek returns the top item without removing it
//
// Returns:
//   - T: the top element
func (this *Queue[T]) Peek() T {
	if len(this.items) == 0 {
		var zero T
		return zero
	}
	return this.items[0]
}

// IsEmpty returns whether the stack is empty
//
// Returns:
//   - bool: true if the stack is empty, false otherwise
func (this *Queue[T]) IsEmpty() bool {
	return len(this.items) == 0
}

// Size returns the number of items
//
// Returns:
//   - int: the number of elements in the stack
func (this *Queue[T]) Size() int {
	return len(this.items)
}
