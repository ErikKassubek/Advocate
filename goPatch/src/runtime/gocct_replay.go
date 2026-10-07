// Copyright (c) 2024 Erik Kassubek
//
// File: gocct_replay.go
// Brief: Functions for the replay
//
// Author: Erik Kassubek
// Created: 2023-10-24
//
// License: BSD-3-Clause

package runtime

const (
	ExitCodeDefault          = 0
	ExitCodePanic            = 3
	ExitCodeTimeout          = 10
	ExitCodeLeakUnbuf        = 20
	ExitCodeLeakBuf          = 21
	ExitCodeLeakMutex        = 22
	ExitCodeLeakCond         = 23
	ExitCodeLeakWG           = 24
	ExitCodeSendClose        = 30
	ExitCodeRecvClose        = 31
	ExitCodeCloseClose       = 32
	ExitCodeCloseNil         = 33
	ExitCodeNegativeWG       = 34
	ExitCodeUnlockBeforeLock = 35
	ExitCodeCyclic           = 41
	ExitCodeMixedDeadlock    = 42
)

const (
	releaseOldestWaitLastMax  int64   = 6
	releaseWaitMaxWait        float64 = 3
	releaseWaitMaxNoWait      float64 = 2
	acknowledgementMaxWaitSec float64 = 1
)

var ExitCodeNames = map[int]string{
	0:  "The replay terminated normally",
	3:  "The program panicked unexpectedly",
	10: "Timeout",
	20: "Leak: Leaking unbuffered channel or select was unstuck",
	21: "Leak: Leaking buffered channel was unstuck",
	22: "Leak: Leaking Mutex was unstuck",
	23: "Leak: Leaking Cond was unstuck",
	24: "Leak: Leaking WaitGroup was unstuck",
	30: "Send on close",
	31: "Receive on close",
	32: "Close on close",
	33: "Close on nil",
	34: "Negative WaitGroup counter",
	35: "Unlock of unlocked mutex",
	41: "Cyclic deadlock",
	42: "Mixed Deadlock",
}

var (
	hasReturnedExitCode = false
	ignoreAtomicsReplay = true
	printDebug          = true

	tPostWhenFirstTimeout    = 0
	tPostWhenReplayDisabled  = 0
	tPostWhenAckFirstTimeout = 0

	releasedFork = 0
)

func SetReplayAtomic(repl bool) {
	ignoreAtomicsReplay = !repl
}

func GetReplayAtomic() bool {
	return !ignoreAtomicsReplay
}

// String representation of the replay operation.
//
// Returns:
//   - string: string representation of the replay operation
func (ro Operation) ToString() string {
	switch ro {
	case OperationNone:
		return "OperationNone"
	case OperationSpawn:
		return "OperationSpawn"
	case OperationSpawned:
		return "OperationSpawned"
	case OperationChannelSend:
		return "OperationChannelSend"
	case OperationChannelRecv:
		return "OperationChannelRecv"
	case OperationChannelClose:
		return "OperationChannelClose"
	case OperationMutexLock:
		return "OperationMutexLock"
	case OperationMutexUnlock:
		return "OperationMutexUnlock"
	case OperationMutexTryLock:
		return "OperationMutexTryLock"
	case OperationRWMutexLock:
		return "OperationRWMutexLock"
	case OperationRWMutexUnlock:
		return "OperationRWMutexUnlock"
	case OperationRWMutexTryLock:
		return "OperationRWMutexTryLock"
	case OperationRWMutexRLock:
		return "OperationRWMutexRLock"
	case OperationRWMutexRUnlock:
		return "OperationRWMutexRUnlock"
	case OperationRWMutexTryRLock:
		return "OperationRWMutexTryRLock"
	case OperationOnceDo:
		return "OperationOnceDo"
	case OperationWaitgroupAddDone:
		return "OperationWaitgroupAddDone"
	case OperationWaitgroupWait:
		return "OperationWaitgroupWait"
	case OperationSelect:
		return "OperationSelect"
	case OperationSelectCase:
		return "OperationSelectCase"
	case OperationSelectDefault:
		return "OperationSelectDefault"
	case OperationCondSignal:
		return "OperationCondSignal"
	case OperationCondBroadcast:
		return "OperationCondBroadcast"
	case OperationCondWait:
		return "OperationCondWait"
	case OperationReplayEnd:
		return "OperationReplayEnd"
	case OperationAtomicLoad:
		return "OperationAtomicLoad"
	case OperationAtomicStore:
		return "OperationAtomicStore"
	case OperationAtomicAdd:
		return "OperationAtomicAdd"
	case OperationAtomicSwap:
		return "OperationAtomicSwap"
	case OperationAtomicCompareAndSwap:
		return "OperationAtomicCompareAndSwap"
	case OperationAtomicAnd:
		return "OperationAtomicAnd"
	case OperationAtomicOr:
		return "OperationAtomicOr"
	default:
		return "Unknown"
	}
}

var (
	replayEnabled bool // replay is on
	replayLock    mutex

	// only run partial replay based on active
	PartialReplay        bool
	partialReplayCounter = make(map[string]int)
	partialReplayMutex   = mutex{}

	// detection of deadlock
	waitDeadlockDetect     bool
	waitDeadlockDetectLock mutex

	// replay info
	replayData            = make(GoCCTReplayTrace, 0)
	numberElementsInTrace int
	active                map[string][]int
	startTimeActive       = -1
	spawns                = make(map[int][]int)
	selects               = make(map[string][]ReplayElement)

	replayIndex = 0

	// exit code
	// replayForceExit  bool
	expectedExitCode int

	// for leak, TimePre of stuck elem
	stuckReplayExecutedSuc = false

	// for replay timeout
	lastTime              int64
	lastTimeWithoutOldest int64
	releaseOldestWait     = releaseOldestWaitLastMax

	// Map of all currently waiting operations
	waitingOps      = make(map[string]replayChan)
	waitingOpsMutex mutex
	counter         = 0

	// element
	waitForAck released
)

// Enable the replay by starting the replay manager
func EnableReplay() {
	numberElementsInTrace = len(replayData)

	if printDebug {
		println("\nTRACE\n")
		replayData.Print()
		println("\n\n")
	}

	replayEnabled = true

	go ReplayManager()

}

/*
 * Disable the replay. This is called when a stop character in the trace is
 * encountered.
 */
func DisableReplay() {
	if !replayEnabled {
		return
	}

	replayEnabled = false

	ReleaseAllWaiting()
}

func IsReplayEnabled() bool {
	return replayEnabled
}

// When called, the calling routine is blocked and cannot be woken up again
func BlockForever() {
	gopark(nil, nil, waitReasonZero, traceBlockForever, 1)
}

// Return the replay status
//
// Returns:
//   - int: what was the next tPost, when the manager released an oldest for the first time, if never return 0
//   - int: what was the next tPost, when the manager disables the replay because it was stuck, if not return 0
//   - int: what was the next tPost, when an expected Acknowledgement timed out, if never return 0

func GetReplayStatus() (int, int, int) {
	return tPostWhenFirstTimeout, tPostWhenReplayDisabled, tPostWhenAckFirstTimeout
}
