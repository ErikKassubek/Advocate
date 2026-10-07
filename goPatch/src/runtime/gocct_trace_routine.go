// GOCCT-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocct_trace_routine.go
// Brief: Functionality for routines
//
// Author: Erik Kassubek
// Created: 2024-02-16
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
type GoCCTTraceSpawn struct {
	t     int64
	newID uint64
	file  string
	line  int
}

// Struct to store the termination of a routine
//
// Fields
//   - tPost int64: time when the routine finished
type GoCCTTraceRoutineExit struct {
	tPost int64
}

// GoCCTSpawnCaller adds a routine spawn to the trace
//
// Parameter:
//   - callerRoutine *GoCCTRoutine: routine that created the new routine
//   - newID uint64: id of the new routine
//   - file string: file where the routine was created
//   - line int32: line where the routine was created
func GoCCTSpawnCaller(callerRoutine *GoCCTRoutine, newID uint64, file string,
	line int32) {
	if GoCCTTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if GoCCTIgnore(file) {
		return
	}

	elem := GoCCTTraceSpawn{
		t:     timer,
		newID: newID,
		file:  file,
		line:  int(line),
	}

	callerRoutine.addToTrace(elem)
}

// Record the finish of a routine
func AdvocatRoutineExit() {
	if GoCCTTracingDisabled {
		return
	}

	timer := GetNextTimeStep()
	elem := GoCCTTraceRoutineExit{
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
func (self GoCCTTraceSpawn) toString() string {
	return buildTraceElemString("G", self.t, self.newID, posToString(self.file, self.line))
}

// Get a string representation of the routine element
//
// Returns:
//   - string: the string representation of the form
//     E,[tPost]
func (elem GoCCTTraceRoutineExit) toString() string {
	return buildTraceElemString("E", elem.tPost)
}

// getOperation is a getter for the spawn element
//
// Returns:
//   - Operation: the operation
func (self GoCCTTraceSpawn) getOperation() Operation {
	return OperationSpawn
}

// getOperation is a getter for the routine exit element
//
// Returns:
//   - Operation: the operation
func (elem GoCCTTraceRoutineExit) getOperation() Operation {
	return OperationRoutineExit
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (self GoCCTTraceSpawn) hasFinished() bool {
	return true
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (elem GoCCTTraceRoutineExit) hasFinished() bool {
	return true
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCCTTraceSpawn) hasCommit() bool {
	return true
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCCTTraceRoutineExit) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCCTTraceResource: recources
func (self GoCCTTraceSpawn) resource() []GoCCTTraceResource {
	return []GoCCTTraceResource{}
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCCTTraceResource: recources
func (self GoCCTTraceRoutineExit) resource() []GoCCTTraceResource {
	return []GoCCTTraceResource{}
}
