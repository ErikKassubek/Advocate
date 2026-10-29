// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_routine.go
// Brief: Functionality for the routines
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

var AdvocateRoutines map[uint64]*AdvocateRoutine
var AdvocateRoutinesLock = mutex{}

var projectPath string

// var atomicRecordingDisabled = false

// AdvocateRoutine is a struct to store the trace of a routine
// Fields:
//   - id uint64: the id of the routine
//   - maxObjectId uint64: the maximum id of elements in the trace
//   - G *g: the g struct of the routine
//   - Trace []traceElem: the trace of the routine
//   - oat []uint64: object aware trace elements, ids of blocked object the routine can access
//   - replayID int: when used in reply, id of the new routine in the replayed trace
//   - forkFile string: file where the routine was created in, "main" for main routine
//   - forkLine int: line where ther routine was created in, 0 for main routine
//   - parkOn []unsafe.Pointer: list of elements the routine was last parked on
//   - parkPos string: position of last park in form file:line
//   - parkForeverReplay bool: if true, routine parks forever based on replay
//   - wokenByTimeout bool: in replay block was woken up by timeout
//   - hasReturned bool: true if the routine has terminated
//   - hasPaniked bool: true if the routine has panicked
type AdvocateRoutine struct {
	id                   uint64
	maxObjectId          uint64
	G                    *g
	Trace                []traceElem
	oat                  []uint64
	replayID             int
	forkFile             string
	forkLine             int32
	parkForeverReplay    bool
	hasReturned          bool
	wokenButTimeout      bool
	wokenNoTimeout       bool
	startedWritingToFile bool
	hasPanicked          bool
}

// Create a new advocate routine
// Params:
//   - g: the g struct of the routine
//   - replayRoutine int: when used in reply, id of the new routine in the replayed trace the replay ids of the routines forked from this routine
//   - file string: file, where the routine was created
//   - line int32: line, where the routine was created
//
// Return:
//   - the new advocate routine
func newAdvocateRoutine(g *g, replayRoutine int, file string, line int32) *AdvocateRoutine {
	println("NEW: ", replayRoutine)
	// ignore the internal routines that are run before the main/test function starts
	if AdvocateTracingDisabled {
		return &AdvocateRoutine{
			id:          0,
			maxObjectId: 0,
			G:           g,
			Trace:       make([]traceElem, 0),
			forkFile:    file,
			forkLine:    line,
			replayID:    replayRoutine,
		}
	}

	var id uint64 = 0
	if replayRoutine != -1 {
		id = uint64(replayRoutine)
	} else {
		id = GetNewAdvocateRoutineID()
	}

	advocateRoutineInfo := &AdvocateRoutine{
		id:          id,
		maxObjectId: 0,
		G:           g,
		Trace:       make([]traceElem, 0),
		replayID:    replayRoutine,
		forkFile:    file,
		forkLine:    line,
	}

	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)

	if AdvocateRoutines == nil {
		AdvocateRoutines = make(map[uint64]*AdvocateRoutine)
	}

	AdvocateRoutines[advocateRoutineInfo.id] = advocateRoutineInfo

	return advocateRoutineInfo
}

// setCurrentRoutineToActive will set the id of the current routine to a valid id
// and add it to AdvocateRoutine
// If it already contains a valid id, do nothing.
// Call when tracing gets enabled
func setCurrentRoutineToActive() {
	g := getg()

	if g.advocateRoutineInfo.id != 0 {
		return
	}

	g.advocateRoutineInfo.id = GetNewAdvocateRoutineID()

	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)

	if AdvocateRoutines == nil {
		AdvocateRoutines = make(map[uint64]*AdvocateRoutine)
	}

	AdvocateRoutines[g.advocateRoutineInfo.id] = g.advocateRoutineInfo
}

// Add an element to the trace of the current routine
// Params:
//   - elem: the element to add
//
// Return:
//   - the index of the element in the trace
func (gi *AdvocateRoutine) addToTrace(elem traceElem) int {
	// never needed in actual code, without it the compiler tests fail
	if gi == nil {
		return -1
	}

	gi.Trace = append(gi.Trace, elem)
	return len(gi.Trace) - 1
}

func (gi *AdvocateRoutine) getElement(index int) traceElem {
	return gi.Trace[index]
}

func (gi *AdvocateRoutine) getPosCreated() string {
	return posToString(gi.forkFile, int(int(gi.forkLine)))
}

func (gi *AdvocateRoutine) getLastElement() (int, traceElem) {
	if len(gi.Trace) == 0 {
		return 0, nil
	}
	index := len(gi.Trace) - 1
	return index, gi.Trace[index]
}

func (gi *AdvocateRoutine) GetForkPos() string {
	return posToString(gi.forkFile, int(gi.forkLine))
}

// Update an element in the trace of the current routine
// Params:
//   - index: the index of the element to update
//   - elem: the new element
func (gi *AdvocateRoutine) updateElement(index int, elem traceElem) {
	if AdvocateTracingDisabled {
		return
	}

	if gi == nil {
		return
	}

	if gi.Trace == nil {
		panic("Tried to update element in nil trace")
	}

	if index >= len(gi.Trace) {
		panic("Tried to update element out of bounds")
	}

	gi.Trace[index] = elem
}

func (gi *AdvocateRoutine) panicked() {
	if AdvocateTracingDisabled {
		return
	}

	if gi == nil {
		return
	}

	if gi.Trace == nil {
		panic("Tried to update element in nil trace")
	}

	gi.hasPanicked = true

	i, lastElem := gi.getLastElement()
	if lastElem != nil {
		gi.updateElement(i, lastElem.causedPanic())
	}

}

// Get the current routine
// Return:
//   - *AdvocateRoutine: the current routine
func currentGoRoutineInfo() *AdvocateRoutine {
	return getg().advocateRoutineInfo
}

// GetRoutineID gets the id of the current routine
// Return:
//   - int: id of the current routine, 0 if current routine is nil
func GetRoutineID() int {
	currentRoutine := currentGoRoutineInfo()
	if currentRoutine == nil {
		return 0
	}
	return int(currentRoutine.id)
}

// GetReplayRoutineID returns the replay id of the current routine
// Return:
//   - int: replay id of the current routine, 0 if current routine is nil
func GetReplayRoutineID() int {
	currentRoutine := currentGoRoutineInfo()
	if currentRoutine == nil {
		return 0
	}
	return int(currentRoutine.replayID)
}

// ADVOCATE-FILE-END
