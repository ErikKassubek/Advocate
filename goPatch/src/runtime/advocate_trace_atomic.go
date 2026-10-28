// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace_atomic.go
// Brief: Functionality for atomics
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an operation on an atomic variable
//
// Fields
//   - timer int64: time when the operation was executed
//   - res AdvocateTraceResource: the resource the op is applied to
//   - op Operation: operation type
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceAtomic struct {
	t    int64
	res  AdvocateTraceResource
	op   Operation
	file string
	line int
}

// Add an atomic operation to the trace
// Args:
//   - addr *T: memory address of the atomic
//   - op Operation: the operation type
//   - skip iny: skip for Caller
func AdvocateAtomic[T int32 | int64 | uint32 | uint64 | uintptr | unsafe.Pointer](addr *T, op Operation, skip int) {
	if AdvocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(skip)

	if AdvocateIgnore(file) {
		return
	}

	unsafeAddr := unsafe.Pointer(addr)

	id := uint64(uintptr(unsafeAddr))

	res := AdvocateTraceResource{id: id, addr: unsafeAddr}

	elem := AdvocateTraceAtomic{
		t:    timer,
		res:  res,
		op:   op,
		file: file,
		line: line,
	}

	insertIntoTrace(elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     U,[timer],[id],[operation],[file],[line]
func (self AdvocateTraceAtomic) toString() string {
	opStr := "U"
	switch self.op {
	case OperationAtomicLoad:
		opStr = "L"
	case OperationAtomicStore:
		opStr = "S"
	case OperationAtomicAdd:
		opStr = "A"
	case OperationAtomicSwap:
		opStr = "W"
	case OperationAtomicCompareAndSwap:
		opStr = "C"
	case OperationAtomicAnd:
		opStr = "N"
	case OperationAtomicOr:
		opStr = "O"
	}

	return buildTraceElemString("A", self.t, self.res.id, opStr, posToString(self.file, self.line), "-")
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceAtomic) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceAtomic) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceAtomic) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{self.res}
}

func (self AdvocateTraceAtomic) causedPanic() traceElem { return self }
