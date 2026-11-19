// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace_new_elem.go
// Brief: Functionality to record a make
//
// Author: Erik Kassubek
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
type AdvocateTraceResource struct {
	id   uint64
	addr unsafe.Pointer
}

func newAdvocateTraceResource(id uint64, addr unsafe.Pointer) (AdvocateTraceResource, uint64) {
	if id == 0 {
		id = GetAdvocateObjectID()
	}

	return AdvocateTraceResource{id, addr}, id
}
