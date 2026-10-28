// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace_waitgroup.go
// Brief: Functionality for wait groups
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an operation on a wait group
//
// Fields
//   - tReq int64: time when the operation started
//   - tCom int64: time when the operation finished
//   - res AdvocateTraceResource: the resource the op is applied to
//   - op Operation: operation type
//   - delta int: value by which the internal counter was changed with this operation
//     for Add > 0, for Done -1 and for wait = 0
//   - val int32: value of the internal counter after the operation was executed
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
//   - pan bool: true if operation caused panic
type AdvocateTraceWaitGroup struct {
	tReq  int64
	tCom  int64
	res   AdvocateTraceResource
	op    Operation
	delta int
	val   int32
	file  string
	line  int
	pan   bool
}

// AdvocateWaitGroupAdd adds a waitgroup add or done to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id: id of the waitgroup
//   - delta: delta of the waitgroup
//   - val: value of the waitgroup after the operation
//
// Returns:
//   - index of the operation in the trace
func AdvocateWaitGroupAdd(mem unsafe.Pointer, id uint64, delta int, val int32) int {
	if AdvocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	var file string
	var line int
	if delta > 0 {
		_, file, line, _ = Caller(CallerSkipWaitGroupAddWait)
	} else {
		_, file, line, _ = Caller(CallerSkipWaitGroupDone)
	}

	if AdvocateIgnore(file) {
		return -1
	}

	res := AdvocateTraceResource{id: id, addr: mem}

	elem := AdvocateTraceWaitGroup{
		tReq:  timer,
		op:    OperationWaitgroupAddDone,
		res:   res,
		delta: delta,
		val:   val,
		file:  file,
		line:  line,
	}

	return insertIntoTrace(elem)
}

// AdvocateWaitGroupWaitReq adds a waitgroup wait to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id: id of the waitgroup
//
// Returns:
//   - index of the operation in the trace
func AdvocateWaitGroupWaitReq(mem unsafe.Pointer, id uint64) int {
	if AdvocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipWaitGroupAddWait)

	if AdvocateIgnore(file) {
		return -1
	}

	res := AdvocateTraceResource{id: id, addr: mem}

	elem := AdvocateTraceWaitGroup{
		tReq: timer,
		res:  res,
		op:   OperationWaitgroupWait,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// AdvocateWaitGroupWaitPost adds the end counter to an operation of the trace
// Wait Post
//
// Parameter:
//   - index: index of the operation in the trace
func AdvocateWaitGroupCom(index int) {
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

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceWaitGroup)

	elem.tCom = timer

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     W,[tPre],[tPost],[id],[op],[delta],[val],[file],[line]
func (self AdvocateTraceWaitGroup) toString() string {
	opStr := "A"
	if self.op == OperationWaitgroupWait {
		opStr = "W"
	}

	panStr := "-"
	if self.pan {
		panStr = "P"
	}

	return buildTraceElemString("W", self.tReq, self.tCom, self.res.id, opStr, self.delta, self.val, posToString(self.file, self.line), panStr)
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceWaitGroup) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceWaitGroup) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceWaitGroup) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{self.res}
}

func (self AdvocateTraceWaitGroup) causedPanic() traceElem {
	self.pan = true
	return self
}
