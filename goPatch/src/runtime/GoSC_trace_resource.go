// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: gosc_trace_new_elem.go
// Brief: Functionality to record a make
//
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an make
// For now only channel makes are recorded
//
// Fields
//   - id string: id of the resource
//   - addr unsafe.Pointer: address
type GoscTraceResource struct {
	id   uint64
	addr unsafe.Pointer
}
