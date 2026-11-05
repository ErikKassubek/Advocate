// Copyright (c) 2025 Sebastian Pohsner
//
// File: statsMisc.go
// Brief: Collect miscellaneous statistics about the advocate run
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package stats

import (
	"advocate/analysis/a_base"
	"advocate/fuzzing/modes/f_gopie"
)

var fuzzStats = []statsType{
	testName,
	nrMut,
	nrMutInvalid,
	activeReleased,
	allActiveReleased,
}

var fuzzStatsStr = []string{
	string(testName),
	string(nrMut),
	string(nrMutInvalid),
	string(activeReleased),
	string(allActiveReleased),
}

// Collect stats about each fuzzing run
//
// Returns:
//   - map[string]int: map with the stats
//   - error
func statsFuzz() (map[statsType]int, error) {
	stats := map[statsType]int{}

	stats[nrMut] = f_gopie.NumberTotalMuts
	stats[nrMutInvalid] = f_gopie.NumberInvalidMuts
	stats[activeReleased] = a_base.ActiveReleased
	stats[allActiveReleased] = a_base.AllActiveReleased

	return stats, nil
}
