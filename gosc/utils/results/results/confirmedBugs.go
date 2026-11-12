// Copyright (c) 2025 
//
// File: confirmedBugs.go
// Brief: Collect confirmed bug string
//
//
// License: BSD-3-Clause

package results

var replayedBugs = make(map[string]bool, 0)

// AddBug stores a bug as a bug.
//
// Parameter:
//   - bugStr string: the bug string
//   - confirmed bool: set true if the replay confirmed the bug, to false otherwise
func AddBug(bugStr string, confirmed bool) {
	replayedBugs[bugStr] = confirmed
}

// WasAlreadyConfirmed returns if a bug string has already been confirmed
//
// Parameter:
//   - bugStr string: the bug string to check
//
// Returns:
//   - bool: true if it has been confirmed, false otherwise
func WasAlreadyConfirmed(bugStr string) bool {
	if conf, ok := replayedBugs[bugStr]; ok {
		return conf
	}
	return false
}
