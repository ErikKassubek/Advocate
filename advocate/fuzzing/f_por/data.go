// Copyright (c) 2025 Erik Kassubek
//
// File: data.go
// Brief: Data required for por
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_por

import "advocate/fuzzing/f_constraints"

var alreadyRunROC = make([]f_constraints.Constraint, 0)

func Reset() {
	alreadyRunROC = make([]f_constraints.Constraint, 0)
}
