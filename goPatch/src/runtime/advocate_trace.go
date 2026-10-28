// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace.go
// Brief: Functionality for the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

type Operation string // enum for operation

const (
	OperationNone        Operation = "none"
	OperationSpawn       Operation = "routineSpawn"
	OperationSpawned     Operation = "routineSpawned"
	OperationRoutineExit Operation = "routineExit"

	OperationChannelSend  Operation = "chanSend"
	OperationChannelRecv  Operation = "chanRecv"
	OperationChannelClose Operation = "chanClose"

	OperationMutexLock       Operation = "mutexLock"
	OperationMutexUnlock     Operation = "mutexUnlock"
	OperationMutexTryLock    Operation = "mutexTrylock"
	OperationRWMutexLock     Operation = "rwmutexLock"
	OperationRWMutexUnlock   Operation = "rwmutexUnlock"
	OperationRWMutexTryLock  Operation = "rwmutexTrylock"
	OperationRWMutexRLock    Operation = "rwmutexRlock"
	OperationRWMutexRUnlock  Operation = "rwmutexrunlock"
	OperationRWMutexTryRLock Operation = "rwmutexTryrlock"

	OperationOnceDo    Operation = "onceDo"
	OperationAllocOnce Operation = "allocOnce"

	OperationWaitgroupAddDone Operation = "wgAdddone"
	OperationWaitgroupWait    Operation = "wgWait"

	OperationSelect        Operation = "select"
	OperationSelectCase    Operation = "selectCase"
	OperationSelectDefault Operation = "selectDefault"

	OperationCondSignal    Operation = "condSignal"
	OperationCondBroadcast Operation = "condBroadcast"
	OperationCondWait      Operation = "condWait"

	OperationAtomicLoad           Operation = "atomicLoad"
	OperationAtomicStore          Operation = "atoicStore"
	OperationAtomicAdd            Operation = "atomicAdd"
	OperationAtomicSwap           Operation = "atomicSwap"
	OperationAtomicCompareAndSwap Operation = "atomicCompareandswap"
	OperationAtomicAnd            Operation = "atomicAnd"
	OperationAtomicOr             Operation = "atomicOr"

	OperationAllocChan  Operation = "allocChan"
	OperationAllocMutex Operation = "allocMutex"
	OperationAllocCond  Operation = "allocCond"
	OperationAllocWg    Operation = "allocWg"

	OperationFunctionCall   Operation = "funcCall"
	OperationFunctionReturn Operation = "funcReturn"

	OperationReplayNever Operation = "replayNever"
	OperationReplayEnd   Operation = "replayEnd"

	OperationControllIf     Operation = "controllIf"
	OperationControllSwitch Operation = "controllSwitch"
	OperationControllLoop   Operation = "controllLoop"
)

type Primitive string

const (
	PrimitiveNone     Primitive = "none"
	PrimitiveRoutine  Primitive = "routine"
	PrimitiveChannel  Primitive = "channel"
	PrimitiveMutex    Primitive = "mutex"
	PrimitiveRWMutex  Primitive = "rwmutex"
	PrimitiveOnce     Primitive = "once"
	PrimitiveWG       Primitive = "waitgroup"
	PrimitiveSelect   Primitive = "select"
	PrimitiveCond     Primitive = "cont"
	PrimitiveAtomic   Primitive = "atomic"
	PrimitiveReplay   Primitive = "replay"
	PrimitiveControll Primitive = "controll"
	PrimitiveAlloc    Primitive = "alloc"
	PrimitivUnknown   Primitive = "unknown"
)

const posSep = "#"

type prePost int // enum for pre/post
const (
	pre prePost = iota
	post
	none
)

var AdvocateTracingDisabled = true

// var advocateTraceWritingDisabled = false

// Given an Operation enum, return a string representation
//
// Parameter:
//   - op Operation: the operation
//
// Return:
//   - Primitive: the string representation
func getOperationPrimitive(op Operation) Primitive {
	switch op {
	case OperationNone:
		return PrimitiveNone
	case OperationSpawn, OperationSpawned, OperationRoutineExit:
		return PrimitiveRoutine
	case OperationChannelSend, OperationChannelRecv, OperationChannelClose:
		return PrimitiveChannel
	case OperationMutexLock, OperationMutexUnlock, OperationMutexTryLock:
		return PrimitiveMutex
	case OperationRWMutexLock, OperationRWMutexUnlock, OperationRWMutexTryLock, OperationRWMutexRLock, OperationRWMutexRUnlock, OperationRWMutexTryRLock:
		return PrimitiveRWMutex
	case OperationOnceDo:
		return PrimitiveOnce
	case OperationWaitgroupAddDone, OperationWaitgroupWait:
		return PrimitiveWG
	case OperationSelect, OperationSelectCase, OperationSelectDefault:
		return PrimitiveSelect
	case OperationCondSignal, OperationCondBroadcast, OperationCondWait:
		return PrimitiveCond
	case OperationAtomicLoad, OperationAtomicStore, OperationAtomicAdd, OperationAtomicSwap, OperationAtomicCompareAndSwap, OperationAtomicAnd, OperationAtomicOr:
		return PrimitiveAtomic
	case OperationReplayEnd:
		return PrimitiveReplay
	case OperationControllIf, OperationControllSwitch, OperationControllLoop:
		return PrimitiveControll
	case OperationAllocChan, OperationAllocMutex, OperationAllocCond, OperationAllocOnce, OperationAllocWg:
		return PrimitiveAlloc
	}
	return PrimitivUnknown
}

// Interface to define an trace element
type traceElem interface {
	toString() string
	getOperation() Operation
	hasCommit() bool
	resource() []AdvocateTraceResource
	causedPanic() traceElem
}

// Return a string representation of the trace of the current go routine
//
// Returns:
//   - string representation of the trace
func CurrentTraceToString() string {
	res := ""
	for i, elem := range currentGoRoutineInfo().Trace {
		if i != 0 {
			res += "\n"
		}
		res += elem.toString()
	}

	return res
}

// Add an operation to the trace
//
// Parameter:
//   - elem: element to add to the trace
//
// Returns:
//   - index of the element in the trace
func insertIntoTrace(elem traceElem) int {
	if currentGoRoutineInfo().hasReturned {
		return -1
	}
	return currentGoRoutineInfo().addToTrace(elem)
}

// Print the trace of the current routines
func PrintTrace() {
	routineID := GetRoutineID()
	println("Routine", routineID, ":", CurrentTraceToString())
}

// Return the trace of the routine by id
//
// Parameter:
//   - id: id of the routine
//
// Returns:
//   - chan: the channel the trace is send over
func TraceToChanByID(id uint64) chan string {
	lock(&AdvocateRoutinesLock)

	c := make(chan string, 20)
	if routine, ok := AdvocateRoutines[id]; ok {
		unlock(&AdvocateRoutinesLock)
		go func() {
			res := ""
			blockSize := 1000
			// if atomic recording is disabled
			for i, elem := range routine.Trace {
				res += elem.toString() + "\n"

				if i%blockSize == 0 {
					c <- res
					res = ""
				}
			}

			if !routine.hasReturned && len(routine.oat) != 0 {
				oatElems := "OAT,"
				for i, obj := range routine.oat {
					if i != 0 {
						oatElems += "-"
					}
					oatElems += uint64ToString(obj)
				}

				res += oatElems + "\n"
			}

			if res != "" {
				c <- res
			}

			close(c)
		}()
	} else {
		unlock(&AdvocateRoutinesLock)
	}

	return c
}

// Return whether the trace of a routine' is empty
//
// Parameter:
//   - routine: id of the routine
//
// Returns:
//   - true if the trace is empty, false otherwise
func TraceIsEmptyByRoutine(routine int) bool {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	if routine, ok := AdvocateRoutines[uint64(routine)]; ok {
		return len(routine.Trace) == 0
	}
	return true
}

// Given a list of element, return a string representation of the elements
// separated by ,
//
// Parameter:
//   - values ...any: the elements
//
// Returns:
//   - a concatenated string representation of all values separated by ,
func buildTraceElemString(values ...any) string {
	return buildTraceElemStringSep(",", values...)
}

// Given a list of element, return a string representation of the elements
// separated by a given separator
//
// Parameter:
//   - values ...any: the elements
//   - sep string: the separator
//
// Returns:
//   - a concatenated string representation of all values separated by the separator
func buildTraceElemStringSep(sep string, values ...any) string {
	res := ""
	for i, v := range values {
		if i != 0 {
			res += sep
		}

		res += convToString(v)
	}
	return res
}

// GetNumberOfRoutines returns the number of routines in the trace
//
// Returns:
//   - number of routines in the trace
func GetNumberOfRoutines() int {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	return len(AdvocateRoutines)
}

// DeleteTrace removes all trace elements from the trace
// It does not remove the routine objects them self
// Make sure to call BlockTrace(), before calling this function
func DeleteTrace() {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	for i := range AdvocateRoutines {
		AdvocateRoutines[i].Trace = AdvocateRoutines[i].Trace[:0]
	}
}

// We are only interested in the behaviour of the actual program, not the details
// of the internal implementation.
// Additionally, some operations, like garbage collection and internal operations, can
// cause the replay to get stuck or are not needed.
// For this reason, we ignore all internal operations
//
// Parameter:
//   - file: file in which the operation is executed
//
// Returns:
//   - bool: true if the operation should be ignored, false otherwise
func AdvocateIgnore(file string) bool {
	return (containsStr(file, "goPatch/src/") || containsStr(file, "go/pkg/mod")) &&
		!containsStr(file, "goPatch/src/time/tick.go")
}

func AdvocateIgnoreRoutine(file string) bool {
	return containsStr(file, "goPatch/src/runtime/advocate_trace.go") || containsStr(file, "goPatch/src/runtime/advocate_replay.go")
}

func RemoveActive(id uint64) {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)

	delete(AdvocateRoutines, uint64(id))
}

// IsActive returns if a routine of the given id has been created/started but not yet been written to file, and if it exists, if the writing of the trace has started
//
// Parameter:
//   - id int: routine id
//
// Returns:
//   - bool: true if not started or written to file
func IsActive(id int) (bool, bool) {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)

	if g, ok := AdvocateRoutines[uint64(id)]; ok {
		return ok, g.startedWritingToFile
	}

	return false, false
}

// Write the trace of the current routine to file. After writing, remove from active
//
// Parameter:
//   - id int: routine id
//
// Returns:
//   - bool: true if not started or written to file
func AdvocateWriteTraceToFile() {
	if AdvocateTracingDisabled {
		return
	}

	g := currentGoRoutineInfo()

	if g == nil {
		return
	}

	if AdvocateIgnore(g.forkFile) {
		return
	}

	g.startedWritingToFile = true
	ok := writeTraceToFileFunc(int(g.id), true)
	if !ok { // writing from finishTracing has already started
		return
	}

	RemoveActive(g.id)
}
