// Copyright (c) 2025 
//
// File: pair.go
// Brief: Implement a comparable pair data type
//
//
// License: BSD-3-Clause

package types

// Pair is a type to implement a pair of values
// The types of the values do not have to be the same
type Pair[K comparable, V comparable] struct {
	X K
	Y V
}

// NewPair returns a new pair
//
// Parameter:
//   - x comparable: first value
//   - y comparable: second value
func NewPair[K comparable, V comparable](x K, y V) Pair[K, V] {
	return Pair[K, V]{X: x, Y: y}
}

// IsEqual check if two pairs are equal.
//
// Parameter:
//   - other *Pair[K, V]: the pair to check agains
//
// Returns:
//   - bool: true if equal, false if not
func (this *Pair[K, V]) IsEqual(other Pair[K, V]) bool {
	return this.X == other.X && this.Y == other.Y
}
