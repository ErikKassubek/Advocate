// gosc/analysis/baseA/settings.go

// Copyright (c) 2025 
//
// File: vc.go
// Brief: Data required for how the analysis is run
//
//
// License: BSD-3-Clause

package a_base

import "gosc/utils/flags"

// Settings for the analysis
var (
	ModeIsFuzzing bool

	// analysis cases to run
	AnalysisCasesMap    = make(map[flags.AnalysisCases]bool)
	AnalysisFuzzingFlow = false

	// exit code info
	ExitCode int
	ExitPos  string

	BugWasFound = false

	// replay timeout info
	replayTimeoutOldest   int
	replayTimeoutDisabled int
	replayTimeoutAck      int

	// replay partial info
	ActiveReleased    int
	AllActiveReleased int

	durationInSeconds = -1 // the duration of the recording in seconds
)

// SetExitInfo stores the exit code and exit position of a run
//
// Parameter:
//   - code int: the exit code
//   - pos string: the exit position
func SetExitInfo(code int, pos string) {
	ExitCode = code
	ExitPos = pos
}

// SetReplayInfo stores information about wether a run that was guided
// by replay (especially in GoPie fuzzing) had a timeout
//
// Parameter:
//   - oldest int: the timer when the the replay released the oldest waiting
//     or the current next for the first time, if never it should be 0
//   - disabled int: the timer when the the replay was so stuck, that the
//     replay had to be disabled for the first time, if never it should be 0
//   - ack int: the timer when the the replay timed out on an acknowledgement,
//     if never it should be 0
//   - activeReached int: 1 if at least one active element was released, 0 otherwise
//   - allActiveReleased int: 1 if all active elements have been released, 0 otherwise
func SetReplayInfo(oldest, disabled, ack, activeReached, allActiveReleased int) {
	replayTimeoutOldest = oldest
	replayTimeoutDisabled = disabled
	replayTimeoutAck = ack
	ActiveReleased = activeReached
	AllActiveReleased = allActiveReleased
}

// GetTimeoutHappened return if any kind of timeout happened
// A timeout happened if at least one of the three timeout var is not 0
//
// Parameter:
//   - ack bool: if false, ignore acknowledgement timeouts
//
// Returns:
//   - - bool: true if a timeout happened, false otherwise
func GetTimeoutHappened(ack bool) bool {
	if ack {
		return (replayTimeoutOldest + replayTimeoutDisabled + replayTimeoutAck) != 0
	} else {
		return (replayTimeoutOldest + replayTimeoutDisabled) != 0
	}
}

// SetRuntimeDurationSec is a setter for durationInSeconds
//
// Parameter:
//   - sec int: the runtime duration of a run in second
func SetRuntimeDurationSec(sec int) {
	durationInSeconds = sec
}

// GetRuntimeDurationInSec is a getter for durationInSeconds
//
// Returns:
//   - int: the runtime duration of a run in second
func GetRuntimeDurationInSec() int {
	return durationInSeconds
}
