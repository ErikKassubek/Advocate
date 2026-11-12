// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: gosc_routine.go
// Brief: Functionality for the routines
//
//
// License: BSD-3-Clause

package runtime

var GoscRoutines map[uint64]*GoscRoutine
var GoscRoutinesLock = mutex{}

var projectPath string

// var atomicRecordingDisabled = false

// GoscRoutine is a struct to store the trace of a routine
// Fields:
//   - id uint64: the id of the routine
//   - maxObjectId uint64: the maximum id of elements in the trace
//   - G *g: the g struct of the routine
//   - Trace []traceElem: the trace of the routine
//   - oat []uint64: object aware trace elements, ids of blocked object the routine can access
//   - ignored bool: the routine gets ignored for tracing and replay
//   - replayID int: when used in reply, id of the new routine in the replayed trace
//   - forkFile string: file where the routine was created in, "main" for main routine
//   - forkLine int: line where ther routine was created in, 0 for main routine
//   - parkOn []unsafe.Pointer: list of elements the routine was last parked on
//   - parkPos string: position of last park in form file:line
//   - parkForeverReplay bool: if true, routine parks forever based on replay
//   - wokenByTimeout bool: in replay block was woken up by timeout
//   - hasReturned bool: true if the routine has terminated
//   - hasPaniked bool: true if the routine has panicked
type GoscRoutine struct {
	id                   uint64
	maxObjectId          uint64
	G                    *g
	Trace                []traceElem
	oat                  []uint64
	ignored bool
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

// Create a new gosc routine
// Params:
//   - g: the g struct of the routine
//   - replayRoutine int: when used in reply, id of the new routine in the replayed trace the replay ids of the routines forked from this routine
//   - file string: file, where the routine was created
//   - line int32: line, where the routine was created
//   - ignored bool: ignore routine
//
// Return:
//   - the new gosc routine
func newGoscRoutine(g *g, replayRoutine int, file string, line int32, ignored bool) *GoscRoutine {
	// ignore the internal routines that are run before the main/test function starts
	if GoscTracingDisabled {
		return &GoscRoutine{
			id:          0,
			maxObjectId: 0,
			G:           g,
			Trace:       make([]traceElem, 0),
			forkFile:    file,
			forkLine:    line,
			replayID:    replayRoutine,
			ignored: ignored,
		}
	}

	var id uint64 = 0
	
	if ignored {
		id = 0
		} else{
		if replayRoutine != -1 {
			id = uint64(replayRoutine)
		} else {
			id = GetNewGoscRoutineID()
		}
}

	goscRoutineInfo := &GoscRoutine{
		id:          id,
		maxObjectId: 0,
		G:           g,
		Trace:       make([]traceElem, 0),
		replayID:    replayRoutine,
		forkFile:    file,
		forkLine:    line,
		ignored: ignored,
	}

	lock(&GoscRoutinesLock)
	defer unlock(&GoscRoutinesLock)

	if GoscRoutines == nil {
		GoscRoutines = make(map[uint64]*GoscRoutine)
	}

	GoscRoutines[goscRoutineInfo.id] = goscRoutineInfo

	return goscRoutineInfo
}

// setCurrentRoutineToActive will set the id of the current routine to a valid id
// and add it to GoscRoutine
// If it already contains a valid id, do nothing.
// Call when tracing gets enabled
func setCurrentRoutineToActive() {
	g := getg()

	if g.goscRoutineInfo.id != 0 {
		return
	}

	g.goscRoutineInfo.id = GetNewGoscRoutineID()

	lock(&GoscRoutinesLock)
	defer unlock(&GoscRoutinesLock)

	if GoscRoutines == nil {
		GoscRoutines = make(map[uint64]*GoscRoutine)
	}

	GoscRoutines[g.goscRoutineInfo.id] = g.goscRoutineInfo
}

// Add an element to the trace of the current routine
// Params:
//   - elem: the element to add
//
// Return:
//   - the index of the element in the trace
func (gi *GoscRoutine) addToTrace(elem traceElem) int {
	// never needed in actual code, without it the compiler tests fail
	if gi == nil {
		return -1
	}

	gi.Trace = append(gi.Trace, elem)
	return len(gi.Trace) - 1
}

func (gi *GoscRoutine) getElement(index int) traceElem {
	return gi.Trace[index]
}

func (gi *GoscRoutine) getPosCreated() string {
	return posToString(gi.forkFile, int(int(gi.forkLine)))
}

func (gi *GoscRoutine) getLastElement() (int, traceElem) {
	if len(gi.Trace) == 0 {
		return 0, nil
	}
	index := len(gi.Trace) - 1
	return index, gi.Trace[index]
}

func (gi *GoscRoutine) GetForkPos() string {
	return posToString(gi.forkFile, int(gi.forkLine))
}

// Update an element in the trace of the current routine
// Params:
//   - index: the index of the element to update
//   - elem: the new element
func (gi *GoscRoutine) updateElement(index int, elem traceElem) {
	if GoscTracingDisabled {
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

func (gi *GoscRoutine) panicked() {
	if GoscTracingDisabled {
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
//   - *GoscRoutine: the current routine
func currentGoRoutineInfo() *GoscRoutine {
	return getg().goscRoutineInfo
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

// GOSC-FILE-END
