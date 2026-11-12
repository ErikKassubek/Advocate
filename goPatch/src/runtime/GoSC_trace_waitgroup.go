// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: gosc_trace_waitgroup.go
// Brief: Functionality for wait groups
//
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an operation on a wait group
//
// Fields
//   - tReq int64: time when the operation started
//   - tCom int64: time when the operation finished
//   - res GoscTraceResource: the resource the op is applied to
//   - op Operation: operation type
//   - delta int: value by which the internal counter was changed with this operation
//     for Add > 0, for Done -1 and for wait = 0
//   - val int32: value of the internal counter after the operation was executed
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
//   - pan bool: true if operation caused panic
type GoscTraceWaitGroup struct {
	tReq  int64
	tCom  int64
	res   GoscTraceResource
	op    Operation
	delta int
	val   int32
	file  string
	line  int
	pan   bool
}

// GoscWaitGroupAdd adds a waitgroup add or done to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id: id of the waitgroup
//   - delta: delta of the waitgroup
//   - val: value of the waitgroup after the operation
//
// Returns:
//   - index of the operation in the trace
func GoscWaitGroupAdd(mem unsafe.Pointer, id uint64, delta int, val int32) int {
	if GoscTracingDisabled {
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

	if GoscIgnore(file) {
		return -1
	}

	res := GoscTraceResource{id: id, addr: mem}

	elem := GoscTraceWaitGroup{
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

// GoscWaitGroupWaitReq adds a waitgroup wait to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id: id of the waitgroup
//
// Returns:
//   - index of the operation in the trace
func GoscWaitGroupWaitReq(mem unsafe.Pointer, id uint64) int {
	if GoscTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipWaitGroupAddWait)

	if GoscIgnore(file) {
		return -1
	}

	res := GoscTraceResource{id: id, addr: mem}

	elem := GoscTraceWaitGroup{
		tReq: timer,
		res:  res,
		op:   OperationWaitgroupWait,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// GoscWaitGroupWaitPost adds the end counter to an operation of the trace
// Wait Post
//
// Parameter:
//   - index: index of the operation in the trace
func GoscWaitGroupCom(index int) {
	if GoscTracingDisabled {
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

	elem := currentGoRoutineInfo().getElement(index).(GoscTraceWaitGroup)

	elem.tCom = timer

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     W,[tPre],[tPost],[id],[op],[delta],[val],[file],[line]
func (self GoscTraceWaitGroup) toString() string {
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
func (self GoscTraceWaitGroup) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoscTraceWaitGroup) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoscTraceResource: recources
func (self GoscTraceWaitGroup) resource() []GoscTraceResource {
	return []GoscTraceResource{self.res}
}

func (self GoscTraceWaitGroup) causedPanic() traceElem {
	self.pan = true
	return self
}
