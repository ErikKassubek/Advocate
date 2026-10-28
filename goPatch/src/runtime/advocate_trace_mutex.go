// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace_mutex.go
// Brief: Functionality for mutex
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an operation on a mutex
//
// Fields
//   - tReq int64: time when the operation started
//   - tCom int64: time when the operation finished
//   - res AdvocateTraceResource: the resource the op is applied to
//   - op Operation: operation type
//   - suc bool: false if a trymutex did not manage to lock the mutex, true otherwise
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
//   - pan bool: true if caused panic
type AdvocateTraceMutex struct {
	tReq int64
	tCom int64
	res  AdvocateTraceResource
	op   Operation
	suc  bool
	file string
	line int
	pan  bool
}

var lastRWOp = make(map[uint64]int64) // routine -> tCom
var lastRWOpLock mutex

// AdvocateMutexReq adds a mutex lock to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id uint64: id of the mutex
//   - op Operation: type of operation
//
// Returns:
//   - index of the operation in the trace
func AdvocateMutexReq(mem unsafe.Pointer, id uint64, op Operation) int {
	if AdvocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipMutex)

	if AdvocateIgnore(file) {
		return -1
	}

	res := AdvocateTraceResource{id: id, addr: mem}

	elem := AdvocateTraceMutex{
		tReq: timer,
		res:  res,
		op:   op,
		suc:  true,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// AdvocateMutexCom adds the end counter to an operation of the trace.
// For try use AdvocateMutexTryPost.
//
// Parameters:
//   - index: index of the operation in the trace
//   - suc: wether the lock was successfull for try, otherwise true
func AdvocateMutexCom(index int, suc bool) {
	if AdvocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	// internal elements are not in the trace
	if index == -1 {
		return
	}

	// only needed to fix tests

	if currentGoRoutineInfo() == nil {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceMutex)
	routine := currentGoRoutineInfo().id

	lock(&lastRWOpLock)
	if elem.isRw() && lastRWOp[routine] != 0 {
		elem.tCom = lastRWOp[routine] - 1
		lastRWOp[routine] = 0
	} else {
		elem.tCom = timer
	}

	if hasSuffix(elem.file, "sync/rwmutex.go") {
		lastRWOp[routine] = timer
	}
	unlock(&lastRWOpLock)

	elem.suc = suc

	currentGoRoutineInfo().updateElement(index, elem)
}

// Check if the mutex is a rw mutex
//
// Returns:
//   - bool: true if it is a rwMutex, false otherwise
func (self AdvocateTraceMutex) isRw() bool {
	if self.op == OperationMutexLock || self.op == OperationMutexUnlock || self.op == OperationMutexTryLock {
		return false
	}
	return true
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (self AdvocateTraceMutex) toString() string {
	opStr, rw := self.opRwToString()

	panStr := "-"
	if self.pan {
		panStr = "P"
	}

	return buildTraceElemString("M", self.tReq, self.tCom, self.res.id, rw, opStr, self.suc, posToString(self.file, self.line), panStr)
}

// Get the string representations for the operation and rw fields
//
// Returns:
//   - string: the operation string representation
//   - string: the rw string representation
func (self AdvocateTraceMutex) opRwToString() (string, string) {
	opStr := ""
	rw := "f"
	switch self.op {
	case OperationMutexLock:
		opStr = "L"
	case OperationMutexUnlock:
		opStr = "U"
	case OperationMutexTryLock:
		opStr = "T"
	case OperationRWMutexLock:
		opStr = "L"
		rw = "t"
	case OperationRWMutexUnlock:
		opStr = "U"
		rw = "t"
	case OperationRWMutexTryLock:
		opStr = "T"
		rw = "t"
	case OperationRWMutexRLock:
		opStr = "R"
		rw = "t"
	case OperationRWMutexRUnlock:
		opStr = "N"
		rw = "t"
	case OperationRWMutexTryRLock:
		opStr = "Y"
		rw = "t"
	}

	return opStr, rw
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceMutex) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceMutex) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceMutex) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{self.res}
}

func (self AdvocateTraceMutex) causedPanic() traceElem {
	self.pan = true
	return self
}
