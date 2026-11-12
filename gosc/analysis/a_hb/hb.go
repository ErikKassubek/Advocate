// Copyright (c) 2025 
//
// File: hb.go
// Brief: Happens before enum
//
//
// License: BSD-3-Clause

package a_hb

// HappensBefore is an enum for the possible hb values
type HappensBefore int

// Possible hb values
const (
	Before = iota
	After
	Concurrent
	None
)
