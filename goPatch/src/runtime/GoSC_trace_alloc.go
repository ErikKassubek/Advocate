// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: gosc_trace_new_elem.go
// Brief: Functionality to record a make
//
//
// License: BSD-3-Clause

package runtime

// Struct to store an make
// For now only channel makes are recorded
//
// Fields
//   - tPost int64: time when the operation finished
//   - res GoscTraceResource: the resource the op is applied to
//   - op Opertions: type of created primitive
//   - qSizeu int: for channel the buffer size, for all other type always 0
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoscTraceAlloc struct {
	t     int64
	res   GoscTraceResource
	qSize int
	file  string
	line  int
	op    Operation
}

/*
 * GoscAlloc adds a channel make to the trace.
 * Args:
 *  opjType string: object type. Must be qualt to string(Operation)
 *  qSizeum int: for channel the buffer size, for all other type always 0
 * Return:
 * 	(int): id for the channel
 */
func GoscAlloc(objType string, qSize int) uint64 {
	if GoscTracingDisabled {
		return 0
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	id := GetGoscObjectID()

	if GoscIgnore(file) {
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
	case "A":
		op = OperationAllocAtomic
	}

	res := GoscTraceResource{id: id, addr: nil}

	elem := GoscTraceAlloc{
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
func (self GoscTraceAlloc) toString() string {
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
	case OperationAllocOnce:
		operationType = "O"
	case OperationAllocAtomic:
		operationType = "A"
	}

	return buildTraceElemString("N", self.t, self.res.id, operationType, self.qSize, posToString(self.file, self.line), "-")
}

// Get the string representation for the primitive type
//
// Returns:
//   - string representation of the primitive type
// func (self GoscTraceAlloc) getOpStr() string {
// 	switch self.op {
// 	case OperationAllocChan:
// 		return "NC"
// 	case OperationAllocMutex:
// 		return "NM"
// 	case OperationAllocCond:
// 		return "ND"
// 	case OperationAllocWg:
// 		return "NW"
// 	}

// 	return "NX"
// }

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self GoscTraceAlloc) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoscTraceAlloc) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoscTraceResource: recources
func (self GoscTraceAlloc) resource() []GoscTraceResource {
	return []GoscTraceResource{self.res}
}

func (self GoscTraceAlloc) causedPanic() traceElem { return self }
