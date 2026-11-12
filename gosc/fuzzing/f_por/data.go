// Copyright (c) 2025 
//
// File: data.go
// Brief: Data required for por
//
//
// License: BSD-3-Clause

package f_por

import "gosc/fuzzing/f_constraints"

var alreadyRunROC = make([]f_constraints.Constraint, 0)

func Reset() {
	alreadyRunROC = make([]f_constraints.Constraint, 0)
}
