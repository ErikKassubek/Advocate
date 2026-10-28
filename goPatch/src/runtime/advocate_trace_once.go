// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace_once.go
// Brief: Functionality for once
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an operation on a once
//
// Fields
//   - tReq int64: time when the operation started
//   - tCom int64: time when the operation finished
//   - res AdvocateTraceResource: the resource the op is applied to
//   - suc bool: true if the func in the Do was executed, false otherwise
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceOnce struct {
	tReq int64
	tCom int64
	res  AdvocateTraceResource
	suc  bool
	file string
	line int
}

// AdvocateOnceReq adds a once to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id uint64: id of the once
//
// Returns:
//   - int: index of the operation in the trace
func AdvocateOnceReq(mem unsafe.Pointer, id uint64) int {
	if AdvocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	if AdvocateIgnore(file) {
		return -1
	}

	res := AdvocateTraceResource{id: id, addr: mem}

	elem := AdvocateTraceOnce{
		tReq: timer,
		res:  res,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// Add the end counter to an operation of the trace
//
// Parameter:
//   - index int: index of the operation in the trace
//   - suc bool: true if the do on the once was called for the first time, false otherwise
func AdvocateOnceCom(index int, suc bool) {
	if AdvocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}
	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceOnce)

	elem.tCom = timer
	elem.suc = suc

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     O,[tPre],[tPost],[id],[suc],[file],[line]
func (self AdvocateTraceOnce) toString() string {
	return buildTraceElemString("O", self.tReq, self.tCom, self.res.id, self.suc, posToString(self.file, self.line), "-")
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceOnce) getOperation() Operation {
	return OperationOnceDo
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceOnce) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceOnce) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{self.res}
}

func (self AdvocateTraceOnce) causedPanic() traceElem { return self }
