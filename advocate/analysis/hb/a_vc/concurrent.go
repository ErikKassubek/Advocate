// Copyright (c) 2025 Erik Kassubek
//
// File: vc.go
// Brief: Data required for calculating the vector clocks
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_vc

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_clock"
	"advocate/analysis/hb/a_helper"
	"advocate/trace"
	"advocate/utils/log"
)

// Current vector clocks
var (
	// current happens before vector clocks
	CurrentVC = make(map[int]*a_clock.VectorClock)

	// current must happens before vector clocks
	CurrentWVC = make(map[int]*a_clock.VectorClock)

	// vector clock for each buffer place in vector clock
	// the map key is the channel id. The slice is used for the buffer positions
	chanBuffer = make(map[int]([]a_base.BufferedVC))
	// the current buffer position
	chanBufferSize = make(map[int]int)
)

// InitVC initializes the current vector clocks
func InitVC() {
	chanBuffer = make(map[int][]a_base.BufferedVC)
	chanBufferSize = make(map[int]int)
	CurrentVC = make(map[int]*a_clock.VectorClock)
	CurrentWVC = make(map[int]*a_clock.VectorClock)

	noRoutine := a_base.MainTrace.GetNoRoutines()
	for i := 1; i <= noRoutine; i++ {
		CurrentVC[i] = a_clock.NewVectorClock(noRoutine)
		CurrentWVC[i] = a_clock.NewVectorClock(noRoutine)
	}
}

// GetConcurrent find a/all element(s) that are concurrent to a given element
// This function assumes that the vector clocks have already been calculated
// The function iterates over all elements, and compares the vector clocks
//
// Parameter:
//   - elem trace.TraceElement: the element to search for
//   - all bool: if true, find all concurrent elements, if false, find only one
//   - sameElem bool: if true, only return concurrent operations on the same element,
//     otherwise return all concurrent elements
//   - sameType bool: only count values on the same type (no effect if same element is true)
//   - weak bool: use the weak happens before relation
//
// Returns:
//   - []trace.Element: set of elements concurrent to elem
func GetConcurrent(elem trace.Element, all, sameElem, sameType, weak bool) []trace.Element {
	if !a_base.HBWasCalc() {
		log.Error("Cannot find concurrent elements: VCs have not been calculated")
		return make([]trace.Element, 0)
	}

	res := make([]trace.Element, 0)
	for rout, routine := range a_base.MainTrace.Routines() {
		if rout == elem.RoutineID() {
			continue
		}

		for _, tElem := range routine.Elems() {
			if !tElem.Committed() {
				continue
			}

			if sameElem && elem.ResourceID() != tElem.ResourceID() {
				continue
			}

			elemType := elem.Type(false)
			tElemType := tElem.Type(false)

			if sameType && elemType != tElemType &&
				!((elemType == trace.Select && tElemType == trace.Channel) ||
					(elemType == trace.Channel && tElemType == trace.Select)) {
				continue
			}

			if !a_helper.Valid(tElem) {
				continue
			}

			if weak {
				if a_clock.IsConcurrent(elem.GetVC(a_clock.Weak), tElem.GetVC(a_clock.Weak)) {
					res = append(res, tElem)
					if !all {
						return res
					}
				}
			} else {
				if a_clock.IsConcurrent(elem.GetVC(a_clock.Strong), tElem.GetVC(a_clock.Strong)) {
					res = append(res, tElem)
					if !all {
						return res
					}
				}
			}
		}
	}

	return res
}
