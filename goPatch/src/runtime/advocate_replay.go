// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_replay.go
// Brief: Entry point for replay
//
// Author: Erik Kassubek
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
