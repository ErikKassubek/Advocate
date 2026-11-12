// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: gosc_replay.go
// Brief: Entry point for replay
//
//
// License: BSD-3-Clause

package runtime

// Enable the replay
func EnableReplay() {
	replayEnabled = true
}

func DisableReplay() {
	replayEnabled = false
}
