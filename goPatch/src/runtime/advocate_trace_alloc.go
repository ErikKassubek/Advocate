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

// Struct to store an make
// For now only channel makes are recorded
//
// Fields
//   - tPost int64: time when the operation finished
//   - res AdvocateTraceResource: the resource the op is applied to
//   - op Opertions: type of created primitive
//   - qSizeu int: for channel the buffer size, for all other type always 0
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceAlloc struct {
	t     int64
	res   AdvocateTraceResource
	qSize int
	file  string
	line  int
	op    Operation
}

/*
 * AdvocateAlloc adds a channel make to the trace.
 * Args:
 *  opjType string: object type. Must be qualt to string(Operation)
 *  qSizeum int: for channel the buffer size, for all other type always 0
 * Return:
 * 	(int): id for the channel
 */
func AdvocateAlloc(objType string, qSize int) uint64 {
	if AdvocateTracingDisabled {
		return 0
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	id := GetAdvocateObjectID()

	if AdvocateIgnore(file) {
		return id
	}

	op := OperationNone
	switch objType {
	case "C":
		op = OperationAllocChan
	case "M":
		op = OperationAllocMutex
	case "D":
		op = OperationAllocCond
	case "W":
		op = OperationAllocWg
	case "O":
		op = OperationAllocOnce
	}

	res := AdvocateTraceResource{id: id, addr: nil}

	elem := AdvocateTraceAlloc{
		t:     timer,
		res:   res,
		file:  file,
		line:  line,
		qSize: qSize,
		op:    op,
	}

	insertIntoTrace(elem)

	return id
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (self AdvocateTraceAlloc) toString() string {
	operationType := ""

	switch self.op {
	case OperationAllocChan:
		operationType = "C"
	case OperationAllocMutex:
		operationType = "M"
	case OperationAllocCond:
		operationType = "D"
	case OperationAllocWg:
		operationType = "W"
	}

	return buildTraceElemString("N", self.t, self.res.id, operationType, self.qSize, posToString(self.file, self.line), "-")
}

// Get the string representation for the primitive type
// For now always return "NC"
//
// Returns:
//   - string representation of the primitive type
func (self AdvocateTraceAlloc) getOpStr() string {
	return "NC"
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceAlloc) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceAlloc) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceAlloc) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{self.res}
}

func (self AdvocateTraceAlloc) causedPanic() traceElem { return self }
