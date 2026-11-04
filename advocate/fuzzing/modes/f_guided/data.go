// Copyright (c) 2025 Erik Kassubek
//
// File: data.go
// Brief: Data for guided fuzzing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_guided

const (
	maxNumberConstraints    = 10
	maxNumberOfMutsPerConst = 5
	maxTries                = 1000
	lengthConstraint        = 8
	propToSkipEquiv         = 0.9
)

var (
	numberMuts = 0
	traceID    = 0
)
