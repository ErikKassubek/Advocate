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

import (
	"internal/runtime/atomic"
)

const releaseSleep = 0.01
const timeoutSec = 4
const maxTimeout = 5

var lastReleseTime atomic.Int64
var timeoutRun = false
var timeoutMaxReached = false

type waitingOp struct {
	ch chan int
	t  int64
}

type waitingOps struct {
	waiting map[string]waitingOp // key -> waiting chan
	mu      mutex
}

func newWaitingOps() waitingOps {
	return waitingOps{
		waiting: make(map[string]waitingOp),
		mu:      mutex{},
	}
}

func (this *waitingOps) add(key string, cWait chan int) {
	lock(&this.mu)
	this.waiting[key] = waitingOp{cWait, currentTime()}
	unlock(&this.mu)
}

func (this *waitingOps) contains(elem *ConstraintElem) waitingOp {
	key := elem.key
	lock(&this.mu)
	if c, ok := this.waiting[key]; ok {
		unlock(&this.mu)
		return c
	}
	unlock(&this.mu)
	return waitingOp{nil, -1}
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
//   - string: element key
//   - chan ReplayElement: channel to wait on, if nil no wait
//
// TODO:
//   - maybe add the bool return back in so we do not need to channel if it is released without additional info
func WaitForReplay(op Operation, skip int) (string, chan int) {
	if !replayEnabled {
		return "", nil
	}

	// if Caller is run in the garbage collector, the execution stops
	// this should prevent this
	if currentGoRoutineInfo() != nil && mgcRoutine != 0 && currentGoRoutineInfo().id == mgcRoutine {
		return "", nil
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
//   - string: element key
//   - chan int: channel to wait on, if nil the operation does not need to wait
func WaitForReplayPath(op Operation, file string, line int) (string, chan int) {
	if !replayEnabled {
		return "", nil
	}

	if AdvocateIgnoreReplay(op, file) {
		return "", nil
	}

	routine := GetReplayRoutineID()
	key := intToString(routine) + "$" + file + "#" + intToString(line)

	if !constraintMain.isInConstraint(key) {
		if printDebug {
			println("ReleaseN: ", key)
		}
		return "", nil
	}

	if constraintMain.alreadyExecuted(key, true) {
		if printDebug {
			println("ReleaseS: ", key)
		}
		return "", nil
	}

	chWait := make(chan int, 1)

	// is next, can be release directly
	if ok, info := constraintMain.isNext(key); ok {
		if printDebug {
			println("ReleaseD: ", key)
		}

		release(chWait, info)
		return key, chWait
	}

	if printDebug {
		println("Wait: ", key)
		println("Next: ", constraintMain.getNext().key)
	}

	waiting.add(key, chWait)

	return key, chWait
}

func release(chWait chan int, info int) {
	chWait <- info
	lastReleseTime.Store(currentTime())
}

func AckReplay(key string) {
	if printDebug {
		println("ACK: ", key)
	}
	constraintMain.update(key)
}

// Call as background routine
func Timeout() {
	lastReleseTime.Store(currentTime())
	timeoutNs := sToNs(timeoutSec)
	timeoutCounter := 0

	for {
		sleep(timeoutSec)

		last := lastReleseTime.Load()

		if last-currentTime() > timeoutNs {
			if printDebug {
				println("TIMEOUT")
			}
			timeoutRun = true
			timeoutCounter++

			if timeoutCounter > maxTimeout {
				timeoutMaxReached = true
				DisableReplay()
				sleep(0.01)
				ReleaseAllBlocked()
				return
			}
		}
		constraintMain.releaseOldest()
	}
}

func GetReplayInfo() (bool, bool, bool) {
	return timeoutRun, timeoutMaxReached, false // TODO: maybe add ack timeout
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

func ReleaseAllBlocked() {
	constraintMain.releaseAllBlocked()
}

func GetWaiting() []string {
	res := make([]string, len(waiting.waiting))

	i := 0
	for key := range waiting.waiting {
		res[i] = key
	}

	return res
}
