// Copyright (c) 2025 
//
// File: statsType.go
// Brief: Constants for stats types
//
//
// License: BSD-3-Clause

package stats

type statsType string

const (
	testName statsType = "TestName"

	numberElements            statsType = "NumberElements"
	numberRoutines            statsType = "NumberRoutines"
	numberNonEmptyRoutines    statsType = "NumberNonEmptyRoutines"
	numberOfSpawns            statsType = "NumberOfSpawns"
	numberRoutineEnds         statsType = "NumberRoutineEnds"
	atomicElem                statsType = "AtomicElem"
	numberAtomics             statsType = "NumberAtomics"
	numberAtomicOperations    statsType = "NumberAtomicOperations"
	channelElem               statsType = "ChannelElem"
	numberChannels            statsType = "NumberChannels"
	numberBufferedChannels    statsType = "NumberBufferedChannels"
	numberUnbufferedChannels  statsType = "NumberUnbufferedChannels"
	numberChannelOperations   statsType = "NumberChannelOperations"
	numberBufferedOps         statsType = "NumberBufferedOps"
	numberUnbufferedOps       statsType = "NumberUnbufferedOps"
	numberSelects             statsType = "NumberSelects"
	numberSelectCases         statsType = "NumberSelectCases"
	numberSelectChanOps       statsType = "NumberSelectChanOps"
	numberSelectDefaultOps    statsType = "NumberSelectDefaultOps"
	mutexElem                 statsType = "MutexElem"
	numberMutexes             statsType = "NumberMutexes"
	numberMutexOperations     statsType = "NumberMutexOperations"
	waitGroupElem             statsType = "waitGroupElem"
	numberWaitGroups          statsType = "NumberWaitGroups"
	numberWaitGroupOperations statsType = "NumberWaitGroupOperations"
	condVarElem               statsType = "CondVarElem"
	numberCondVars            statsType = "NumberCondVars"
	numberCondVarOperations   statsType = "NumberCondVarOperations"
	onceElem                  statsType = "OnceElem"
	numberOnce                statsType = "NumberOnce"
	numberOnceOperations      statsType = "NumberOnceOperations"

	nrMut             statsType = "NrMut"
	nrMutInvalid      statsType = "NrMutInvalid"
	activeReleased    statsType = "ActiveReleased"
	allActiveReleased statsType = "AllActiveReleased"

	numberFiles         statsType = "NrFiles"
	numberLines         statsType = "NrLines"
	numberNonEmptyLines statsType = "NrNonEmptyLine"

	detected         statsType = "detected"
	replayWritten    statsType = "replayWritten"
	replaySuccessful statsType = "replaySuccessful"
	unexpectedPanic  statsType = "unexpectedPanic"
	falsePositive    statsType = "falsePositive"
)
