// Copyright (c) 2026 
//
// File: equivalent.go
// Brief: Check if an roc is constraint equivalent to the trace it is based on
//
//
// License: BSD-3-Clause

package f_constraints

import (
	"gosc/analysis/a_hb"
	"gosc/analysis/hb/a_clock"
	"gosc/analysis/hb/a_vc"
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
			if a_vc.GetHappensBefore(elem2, elem1, a_clock.Strong) == a_hb.Before {
				return false
			}
		}
	}

	return true
}
