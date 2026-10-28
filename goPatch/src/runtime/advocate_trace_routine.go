// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace_routine.go
// Brief: Functionality for routines
//
// Author: Erik Kassubek
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
type AdvocateTraceSpawn struct {
	t     int64
	newID uint64
	file  string
	line  int
}

// Struct to store the termination of a routine
//
// Fields
//   - tPost int64: time when the routine finished
type AdvocateTraceRoutineExit struct {
	tPost int64
}

// AdvocateSpawnCaller adds a routine spawn to the trace
//
// Parameter:
//   - callerRoutine *AdvocateRoutine: routine that created the new routine
//   - newID uint64: id of the new routine
//   - file string: file where the routine was created
//   - line int32: line where the routine was created
func AdvocateSpawnCaller(callerRoutine *AdvocateRoutine, newID uint64, file string,
	line int32) {
	if AdvocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if AdvocateIgnoreRoutine(file) {
		return
	}

	elem := AdvocateTraceSpawn{
		t:     timer,
		newID: newID,
		file:  file,
		line:  int(line),
	}

	callerRoutine.addToTrace(elem)
}

// Record the finish of a routine
func AdvocatRoutineExit() {
	if AdvocateTracingDisabled {
		return
	}

	if currentGoRoutineInfo().hasPanicked {
		return
	}

	timer := GetNextTimeStep()
	elem := AdvocateTraceRoutineExit{
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
func (self AdvocateTraceSpawn) toString() string {
	return buildTraceElemString("G", self.t, self.newID, posToString(self.file, self.line), "-")
}

// Get a string representation of the routine element
//
// Returns:
//   - string: the string representation of the form
//     E,[tPost]
func (elem AdvocateTraceRoutineExit) toString() string {
	return buildTraceElemString("E", elem.tPost, "-")
}

// getOperation is a getter for the spawn element
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceSpawn) getOperation() Operation {
	return OperationSpawn
}

// getOperation is a getter for the routine exit element
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceRoutineExit) getOperation() Operation {
	return OperationRoutineExit
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (self AdvocateTraceSpawn) hasFinished() bool {
	return true
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (elem AdvocateTraceRoutineExit) hasFinished() bool {
	return true
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceSpawn) hasCommit() bool {
	return true
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceRoutineExit) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceSpawn) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{}
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceRoutineExit) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{}
}

func (self AdvocateTraceSpawn) causedPanic() traceElem       { return self }
func (self AdvocateTraceRoutineExit) causedPanic() traceElem { return self }
