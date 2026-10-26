// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_replay_wait.go
// Brief: Functionality to handle operation wait
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

type waitingOps struct {
	waiting map[string]chan int // key -> waiting chan
	mu      mutex
}

func newWaitingOps() waitingOps {
	return waitingOps{
		waiting: make(map[string]chan int),
		mu:      mutex{},
	}
}

// TODO: locks
func (this *waitingOps) add(key string, c chan int) {
	lock(&this.mu)
	this.waiting[key] = c
	unlock(&this.mu)
}

func (this *waitingOps) contains(elem *ConstraintElem) chan int {
	key := elem.key
	lock(&this.mu)
	if c, ok := this.waiting[key]; ok {
		unlock(&this.mu)
		return c
	}
	unlock(&this.mu)
	return nil
}

func (this *waitingOps) remove(elem *ConstraintElem) {
	key := elem.key
	lock(&this.mu)
	delete(this.waiting, key)
	unlock(&this.mu)
}

// Called by operations during replay. If the operations is the next operation
// to be executed it is release, otherwise the operation is stored into the
// waiting operations. The function provides channels to the operation
// on which they wait to be released and send an acknowledgement when they have
// finished executing
//
// Parameter:
//   - op: the operation type that is about to be executed
//   - skip: number of stack frames to skip
//
// Returns:
//   - chan ReplayElement: channel to wait on, if nil no wait
//
// TODO:
//   - maybe add the bool return back in so we do not need to channel if it is released without additional info
func WaitForReplay(op Operation, skip int) chan int {
	if !replayEnabled {
		return nil
	}

	// if Caller is run in the garbage collector, the execution stops
	// this should prevent this
	if currentGoRoutineInfo() != nil && mgcRoutine != 0 && currentGoRoutineInfo().id == mgcRoutine {
		return nil
	}

	_, file, line, _ := Caller(skip)

	if currentGoRoutineInfo() != nil && containsStr(file, "/src/runtime/mgc.go") {
		mgcRoutine = currentGoRoutineInfo().id
	}

	return WaitForReplayPath(op, file, line)
}

// Called by operations during replay. If the operations is the next operation
// to be executed it is release, otherwise the operation is stored into the
// waiting operations. The function provides channels to the operation
// on which they wait to be released and send an acknowledgement when they have
// finished executing
//
// Parameter:
//   - op: the operation type that is about to be executed
//   - file: file in which the operation is executed
//   - line: line number of the operation
//
// Returns:
//   - chan int: channel to wait on, if nil the operation does not need to wait
func WaitForReplayPath(op Operation, file string, line int) chan int {
	if !replayEnabled {
		return nil
	}

	if AdvocateIgnoreReplay(op, file) {
		return nil
	}

	routine := GetReplayRoutineID()
	key := intToString(routine) + "$" + file + "#" + intToString(line)

	if !constraintMain.isInConstraint(key) {
		if printDebug {
			println("ReleaseN: ", key)
		}
		return nil
	}

	chWait := make(chan int, 1)

	// is next, can be release directly
	if ok, info := constraintMain.isNext(key); ok {
		if printDebug {
			println("ReleaseD: ", key)
		}

		chWait <- info
		constraintMain.update(key)
		return chWait
	}

	if printDebug {
		println("Wait: ", key)
	}

	waiting.add(key, chWait)

	return chWait
}

// AdvocateIgnoreReplay decides if an operation should be ignored for replay.
// Ignored means it is just executed when called without waiting.
// All internal operations are ignored
// Atomic operations are ignored if the corresponding variable is set
//
// Parameter:
//   - operation Operation: the operation to check
//   - file string: the file where the operation is executed
//
// Returns:
//   - bool: true if the operation should be ignored, false otherwise
func AdvocateIgnoreReplay(operation Operation, file string) bool {
	primitive := getOperationPrimitive(operation)

	switch primitive {
	case PrimitiveRoutine:
		if containsStr(file, "goPatch/src/testing/testing.go") {
			return false
		}
	case PrimitiveAlloc, PrimitiveControll:
		return true
	}

	if containsStr(file, "go/pkg/mod/") {
		return true
	}

	return AdvocateIgnore(file)
}

// func BlockForever() {
// 	gopark(nil, nil, waitReasonZero, traceBlockForever, 1)
// }

/*
 * Wait until all operations in the trace are executed.
 * This function should be called after the main routine is finished, to prevent
 * the program to terminate before the trace is finished.
 *
 * Returns:
 *    - bool: true if all events in the pre trace and roc have been executed, false otherwise
 */
func WaitForReplayFinish() bool {
	if printDebug {
		println("Wait for replay finish")
		defer println("Finish Wait")
	}

	startTime := currentTime()
	allEventExec := true

	if replayEnabled {
		for {
			if constraintMain.isDone() {
				break
			}

			if hasTimePast(startTime, 5) {
				allEventExec = false
				break
			}

			sleep(0.001)
		}

		waiting := GetWaiting()
		if len(waiting) != 0 {
			println("WAITING")
		}

		DisableReplay()
	}

	// Ensure that the deadlock detector is finished
	for {
		lock(&waitDeadlockDetectLock)
		if !waitDeadlockDetect {
			unlock(&waitDeadlockDetectLock)
			break
		}
		unlock(&waitDeadlockDetectLock)

		sleep(0.001)
	}

	return allEventExec
}
