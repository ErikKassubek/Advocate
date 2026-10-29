// Copyright (c) 2026 Erik Kassubek
//
// File: equivalent.go
// Brief: Check if an roc is considered equivalent to the trace it is based on
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_constraints

import (
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_vc"
)

// isEquivalent checks if the constraint is sufficiently different from the
// trace it is based on to warred an execution of the constraint
// It is sufficiently different if at least one hb relation is broken
//
//	Parameter:
//	  - cr baseF.Constraint: the constraint to check
//
// Return:
//   - bool: true if the constraint should be skipped, false if it should be executed
func IsEquivalent(cr Constraint) bool {
	for i, elem1 := range cr.Elems {
		for j := i + 1; j < len(cr.Elems); j++ {
			elem2 := cr.Elems[j]
			if a_vc.GetHappensBefore(elem2, elem1, false) == a_hb.Before {
				return false
			}
		}
	}

	return true
}
