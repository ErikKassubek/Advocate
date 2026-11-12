// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: trace_conditional.go
// Brief: Functionality for recording if and switch results
//
//
// License: BSD-3-Clause

package runtime

// Struct to store and if or switch
//
// Fields
//   - t int64: time
//   - numCases int: number of cases
//   - chosenCase int: chosen case in order given in code
//   - ct condType: if or switch
//   - file string: file
//   - line int: first line
type GoscTraceControllFlow struct {
	t          int64
	numCases   int
	chosenCase int
	ct         Operation
	file       string
	line       int
}

// TODO: fix line
// GoscControllFlow inserts a conditional into the trace
//
// Parameter:
//   - ct string: "I" for if, "S" for switch
//   - numCases int: number of cases
//   - chosenCase int: number of chosen case
func goscControllFlow(ct string, numCases, chosenCase int) {
	if GoscTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(1)

	if GoscIgnore(file) {
		return
	}

	var op Operation
	switch ct {
	case "I":
		op = OperationControllIf
	case "S":
		op = OperationControllSwitch
	case "L":
		op = OperationControllLoop
	default:
		panic("Unknown Controll Flow type " + ct)
	}

	elem := GoscTraceControllFlow{
		t:          timer,
		numCases:   numCases,
		chosenCase: chosenCase,
		ct:         op,
		file:       file,
		line:       line,
	}

	insertIntoTrace(elem)
}

func (self GoscTraceControllFlow) toString() string {
	ctString := ""
	switch self.ct {
	case OperationControllIf:
		ctString = "I"
	case OperationControllSwitch:
		ctString = "S"
	case OperationControllLoop:
		ctString = "L"
	default:
		panic("Invalid controll flow type: " + string(self.ct))
	}

	return buildTraceElemString("I", self.t, ctString, self.numCases, self.chosenCase, posToString(self.file, self.line), "-")
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self GoscTraceControllFlow) getOperation() Operation {
	return self.ct
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoscTraceControllFlow) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoscTraceResource: recources
func (self GoscTraceControllFlow) resource() []GoscTraceResource {
	return []GoscTraceResource{}
}

func (self GoscTraceControllFlow) causedPanic() traceElem {
	return self
}
