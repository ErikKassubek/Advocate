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

	stableID = true
	spawnsRout      = make(map[int][]int)
	spawnsPos       = make(map[int][]string)
	maxRoutineIndex = 0
	selects         = make(map[string][]int)

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

func SetSpawns(spRout map[int][]int, spPos map[int][]string) {
	if spRout == nil {
		stableID = false 
		if printDebug {
			println("NO STABLE ID")
		}
		return
	}


	if printDebug {
		println("SET SPAWN")
		for i, j := range spRout {
			print(i, " -> ")
			for k, e := range j {
				print(e, "-", spPos[i][k], ", ")
			}
			println("")
		}
		println("SPAWN SET")
	}

	for _, routs := range spRout {
		for _, rout := range routs {
			maxRoutineIndex = max(maxRoutineIndex, rout)
		}
	}

	spawnsRout = spRout
	spawnsPos = spPos
}
