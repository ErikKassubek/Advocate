// Copyright (c) 2026 
//
// File: gosc_fuzzing.go
// Brief: Fuzzing
//
//
// License: BSD-3-Clause

package runtime

const (
	selectPreferredTimeoutSec int64   = 2
	flowSleepTimeSec          float64 = 3
)

var (
	goscFuzzingEnabled       = false
	goscFuzzingReplayEnabled = false
	fuzzingSelectData            = make(map[string][]int)
	fuzzingSelectDataIndex       = make(map[string]int)

	finishFuzzingFunc func()
)

// Init fuzzing based on preferred select cases
//
// Parameter:
//   - selectData map[string][]int: preferred select cases (select pos -> []preferred case index)
//   - fuzzingFlow map[string][]int: operations to delay (file -> []lines)
//   - finishFuzzing func(): function that should be called if fuzzing has finished
func InitFuzzingSelect(selectData map[string][]int, fuzzingFlow map[string][]int, finishFuzzing func()) {
	finishFuzzingFunc = finishFuzzing
	selects = selectData

	goscFuzzingEnabled = true
	selectEnabled = true
}

// InitFuzzingReplay initializes fuzzing based on full replay
//
// Parameter:
//   - finishFuzzing func(): gosc.FinishFuzzing function
func InitFuzzingReplay(finishFuzzing func()) {
	finishFuzzingFunc = finishFuzzing
	goscFuzzingEnabled = true
	goscFuzzingReplayEnabled = true
}

// Get if fuzzing is enables
//
// Returns:
//   - bool: true if fuzzing is enabled, false otherwise
func IsGoscFuzzingEnabled() bool {
	return goscFuzzingEnabled
}

// TODO: fix for gfuzz

// Get the preferred case for the specified select
//
// Parameter:
//   - skip int: skip for runtime.Caller
//
// Returns:
//   - bool: true if a preferred case exists, false otherwise
//   - int: preferred case, -1 for default
func GoscFuzzingGetPreferredCase(skip int) (bool, int) {
	if !goscFuzzingEnabled {
		return false, 0
	}

	routine := GetReplayRoutineID()

	_, file, line, _ := Caller(skip)
	if GoscIgnore(file) {
		return false, 0
	}
	key := intToString(routine) + ":" + file + "#" + intToString(line)

	if val, ok := fuzzingSelectData[key]; ok {
		index := fuzzingSelectDataIndex[key]
		if index >= len(val) {
			return false, 0
		}
		fuzzingSelectDataIndex[key]++
		return true, val[index]
	}

	return false, 0
}
