// GOSC-FILE_START

// Copyright (c) 2026
//
// File: gosc_trace_routine.go
// Brief: Functionality for routines
//
//
// License: BSD-3-Clause

package runtime

// Struct to store the spawn of a new routine (go func() {}())
//
// Fields
//   - tPost int64: time when the operation finished
//   - newID uint64: id of new routine
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoscTraceSpawn struct {
	t     int64
	newID uint64
	file  string
	line  int
}

// Struct to store the termination of a routine
//
// Fields
//   - tPost int64: time when the routine finished
type GoscTraceRoutineExit struct {
	tPost int64
}

// GoscSpawnCaller adds a routine spawn to the trace
//
// Parameter:
//   - callerRoutine *GoscRoutine: routine that created the new routine
//   - newID uint64: id of the new routine
//   - file string: file where the routine was created
//   - line int32: line where the routine was created
func GoscSpawnCaller(callerRoutine *GoscRoutine, newID uint64, file string,
	line int32) {
	if GoscTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if GoscIgnoreRoutine(file) {
		return
	}

	elem := GoscTraceSpawn{
		t:     timer,
		newID: newID,
		file:  file,
		line:  int(line),
	}

	callerRoutine.addToTrace(elem)
}

// Record the finish of a routine
func GoscRoutineExit() {
	if GoscTracingDisabled {
		return
	}

	rout := currentGoRoutineInfo()
	if rout.ignored || rout.hasPanicked {
		return
	}

	timer := GetNextTimeStep()
	elem := GoscTraceRoutineExit{
		tPost: timer,
	}
	insertIntoTrace(elem)
	currentGoRoutineInfo().hasReturned = true
}

// Get a string representation of a trace spawn
//
// Returns:
//   - string: the string representation of the form
//     G,[tPost],[newID],[file],[line]
func (self GoscTraceSpawn) toString() string {
	return buildTraceElemString("G", self.t, self.newID, posToString(self.file, self.line), "-")
}

// Get a string representation of the routine element
//
// Returns:
//   - string: the string representation of the form
//     E,[tPost]
func (elem GoscTraceRoutineExit) toString() string {
	return buildTraceElemString("E", elem.tPost, "-")
}

// getOperation is a getter for the spawn element
//
// Returns:
//   - Operation: the operation
func (self GoscTraceSpawn) getOperation() Operation {
	return OperationSpawn
}

// getOperation is a getter for the routine exit element
//
// Returns:
//   - Operation: the operation
func (elem GoscTraceRoutineExit) getOperation() Operation {
	return OperationRoutineExit
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (self GoscTraceSpawn) hasFinished() bool {
	return true
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (elem GoscTraceRoutineExit) hasFinished() bool {
	return true
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoscTraceSpawn) hasCommit() bool {
	return true
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoscTraceRoutineExit) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoscTraceResource: recources
func (self GoscTraceSpawn) resource() []GoscTraceResource {
	return []GoscTraceResource{}
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoscTraceResource: recources
func (self GoscTraceRoutineExit) resource() []GoscTraceResource {
	return []GoscTraceResource{}
}

func (self GoscTraceSpawn) causedPanic() traceElem       { return self }
func (self GoscTraceRoutineExit) causedPanic() traceElem { return self }
