// Copyright (c) 2026 
//
// File: gosc_ids.go
// Brief: Get required ids and timestamps
//
//
// License: BSD-3-Clause

package runtime

import (
	"internal/runtime/atomic"
)

var goscCurrentRoutineID atomic.Uint64

// GetNewGoscRoutineID returns a new id for a routine
//
// Returns:
//   - new id
func GetNewGoscRoutineID() uint64 {
	id := goscCurrentRoutineID.Add(1)
	if id > 184467440 {
		panic("Overflow Error: Two many routines. Max: 184467440")
	}
	return id
}

// GetNewGoscRoutineID returns the next routine id that will be provided
// by GetNewGoscRoutineID without advancing the counter
//
// Returns:
//   - next id
func GetNextGoscRoutineID() uint64 {
	return goscCurrentRoutineID.Load() + 1
}

// GetGoscObjectID returns a new id for an primitive
// Return:
//   - new id
func GetGoscObjectID() uint64 {
	routine := currentGoRoutineInfo()

	if routine == nil {
		return 0
	}

	routine.maxObjectId++
	if routine.maxObjectId > 999999999 {
		panic("Overflow Error: Tow many objects in one routine. Max: 999999999")
	}
	id := routine.id*1000000000 + routine.maxObjectId

	return id
}
