// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_replay_data.go
// Brief: Data for constraint execution
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

var (
	constraintMain Constraint
	waiting        = newWaitingOps()

	spawns  = make(map[int][]int)
	selects = make(map[string][]int)

	replayEnabled = false
	selectEnabled = false
)

var (
	NumberElemConstraint = 0
	NumberElemClosure    = 0
	NumberElemReleased   = 0
)

var (
	waitDeadlockDetect     bool
	waitDeadlockDetectLock mutex

	mgcRoutine uint64 = 0
)

var (
	maxWaitingSize     = 0
	currentWaitingSize = 0
)

const printDebug = false
const experiments = true
const withTimeout = true

func SetSpawns(sp map[int][]int) {
	if printDebug {
		println("SET SPAWN")
		for i, j := range sp {
			print(i, " -> ")
			for _, e := range j {
				print(e, ", ")
			}
			println("")
		}
		println("SPAWN SET")
	}

	spawns = sp
}
