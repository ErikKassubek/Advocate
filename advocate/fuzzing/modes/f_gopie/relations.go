// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-relations.go
// Brief: Calculate the relation for goPie
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gopie

import (
	"advocate/fuzzing/f_base"
	"advocate/trace"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/types"
	"sort"
)

// We define <c, c'> in CPOP1, if c and c' are neighboring operations in the same routine.
// We define <c, c'> in CPOP2, if c and c' are operations in different routines
// but on the same primitive.
// From this we define the relations Rel1 and Rel2 with the following rules:
// Rule 1: exists c, c', <c, c'> in CPOP1 -> c' in Rel1(c)  (same routine, element before and after)
// Rule 2: exists c, c', <c, c'> in CPOP2 -> c' in Rel2(c)  (different routine, same primitive)
// Rule 3: exists c, c', c'', c' in Rel1(c), c'' in Rel2(c') -> c'' in Rel2(c)
// Rule 4: exists c, c', c'', c' in Rel2(c), c'' in Rel2(c') -> c'' in Rel2(c)

// CalculateRelRule1 store the rule 1 information for each element in a routine trace
//
// Parameter:
//   - routineTrace trace.Routine: the list of elems in the same trace
func CalculateRelRule1(routineTrace *trace.Routine) {
	if routineTrace.Len() < 2 {
		return
	}

	for i := 0; i < routineTrace.Len()-1; i++ {
		elem1 := routineTrace.At(i)
		if !isGoPieElem(elem1) {
			continue
		}
		elem2 := routineTrace.At(i + 1)
		if !isGoPieElem(elem2) {
			continue
		}
		if _, ok := rel1[elem1]; !ok {
			rel1[elem1] = make(map[trace.Element]struct{})
		}
		if _, ok := rel1[elem2]; !ok {
			rel1[elem2] = make(map[trace.Element]struct{})
		}
		rel1[elem1][elem2] = struct{}{}
		rel1[elem2][elem1] = struct{}{}
		counterCPOP1++
	}
	if control.WasCanceled() {
		return
	}
}

// CalculateRelRule2AddElem add each element in a routine trace to the map from id to operation
//
// Parameter:
//   - elem analysis.TraceElement: Element to add
func CalculateRelRule2AddElem(elem trace.Element) {
	if !isGoPieElem(elem) {
		return
	}

	switch elem := elem.(type) {
	case *trace.ElementSelect:
		for _, c := range elem.GetCases() {
			id := c.ResourceID()
			if _, ok := ElemsByID[id]; !ok {
				ElemsByID[id] = make([]trace.Element, 0)
			}
			ElemsByID[id] = append(ElemsByID[id], elem)
			counterCPOP2++
		}
	default:
		id := elem.ResourceID()
		if _, ok := ElemsByID[id]; !ok {
			ElemsByID[id] = make([]trace.Element, 0)
		}
		ElemsByID[id] = append(ElemsByID[id], elem)
		counterCPOP2++
	}
}

// CalculateRelRule2And4 applies rule 2 and directly rule 4 for all elements
// This calculates rel2 as the tuples, of elements on the same primitive
// but in different routines
func CalculateRelRule2And4() {
	for _, elems := range ElemsByID {
		sort.Slice(elems, func(i, j int) bool {
			return elems[i].T(trace.Sorting) < elems[j].T(trace.Sorting)
		})

		for i := 0; i < len(elems)-1; i++ {
			elem1 := elems[i]
			for j := i + 1; j < len(elems)-1; j++ {
				elem2 := elems[j]
				if elem1.RoutineID() != elem2.RoutineID() {
					if _, ok := rel2[elem1]; !ok {
						rel2[elem1] = make(map[trace.Element]struct{})
					}
					if _, ok := rel2[elem2]; !ok {
						rel2[elem2] = make(map[trace.Element]struct{})
					}

					rel2[elem1][elem2] = struct{}{}
					rel2[elem2][elem1] = struct{}{}
					counterCPOP2++
				}
				if control.WasCanceled() {
					return
				}
			}
		}
	}
}

// CalculateRelRule3 applies rule 3 for all elements
func CalculateRelRule3() {
	changed := true
	for changed {
		changed = false

		// Rule 3
		for c, rel1Elems := range rel1 {
			for cPrime := range rel1Elems {
				if rel2Elems, ok := rel2[cPrime]; ok {
					if _, exists := rel2[c]; !exists {
						rel2[c] = make(map[trace.Element]struct{})
					}
					for cDoublePrime := range rel2Elems {
						if _, exists := rel2[c][cDoublePrime]; !exists {
							rel2[c][cDoublePrime] = struct{}{}
							changed = true
						}
					}
				}
			}
		}

		// Rule 4
		// for c, rel2Elems := range rel2 {
		// 	newTargets := make(map[trace.TraceElement]struct{})
		// 	for cPrime := range rel2Elems {
		// 		if nestedElems, ok := rel2[cPrime]; ok {
		// 			for cDoublePrime := range nestedElems {
		// 				if _, exists := rel2Elems[cDoublePrime]; !exists {
		// 					newTargets[cDoublePrime] = struct{}{}
		// 					changed = true
		// 				}
		// 			}
		// 		}
		// 	}

		// 	for k := range newTargets {
		// 		rel2[c][k] = struct{}{}
		// 	}
		// }
	}
}

// isGoPieElem returns if an element is part of goPie
// GoPie only looks at fork, mutex, rwmutex and channel (and select)
// GoCRHB uses all repayable elements
//
// Parameter:
//   - elem analysis.TraceElement: the element to check
//
// Returns:
//   - bool: true if elem should be used in chains, false if not
func isGoPieElem(elem trace.Element) bool {
	elemTypeShort := elem.Type(false)

	if flags.FuzzingMode == f_base.GoPie {
		validTypes := []trace.OperationType{
			trace.Mutex, trace.Channel,
			trace.Select}
		return types.Contains(validTypes, elemTypeShort)
	}

	invalidTypes := []trace.OperationType{trace.New,
		trace.Replay}
	if types.Contains(invalidTypes, elemTypeShort) {
		return false
	}

	return elem.Type(true) != trace.RoutEnd
}
