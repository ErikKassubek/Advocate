// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: gosc_time.go
// Brief: Get the timer
//
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var goscGlobalCounter atomic.Int64

// GetGoscCounter will update the timer and return the new value
// Return:
//   - new time value
func GetNextTimeStep() int64 {
	return goscGlobalCounter.Add(2)
}

// GOSC-FILE-END
