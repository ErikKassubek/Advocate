// Copyright (c) 2025 Erik Kassubek
//
// File: chainMutation.go
// Brief: Mutations for chains
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_base

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_concurrent"
	"advocate/fuzzing/f_constraints"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/settings"
	"maps"
	"math"
	"math/rand/v2"
)

// Create the mutations for a GoPie chain
//
// Parameter:
//   - c chain: The scheduling chain to mutate
//   - energy int: Determines how many mutations are created
//   - rel1 map[trace.Element]map[trace.Element]struct{}: the rel1 info for goPie
//   - rel2 map[trace.Element]map[trace.Element]struct{}: the rel2 info for goPie
//
// Returns:
//   - map[string]chain: Set of mutations
func Mutate(c f_constraints.Constraint, energy int, rel1, rel2 map[trace.Element]map[trace.Element]struct{}) map[string]f_constraints.Constraint {
	if energy > 100 {
		energy = 100
	}

	bound := settings.GoPieMaxSCLength
	if flags.FuzzingMode == GoPie {
		bound = 3
	}

	mutateBound := settings.ChainMutabound

	// in the original goPie, the fuzzing bound is 3
	if flags.FuzzingMode == GoPie {
		bound = 3
	}

	res := make(map[string]f_constraints.Constraint)

	if energy == 0 {
		return res
	}

	if c.Len() == 0 {
		return res
	}

	res[c.String()] = c

	maxMutPerStep := 6
	if flags.FuzzingMode == GoPie {
		maxMutPerStep = -1
	}

	for {
		noNew := false
		for _, ch := range res {
			tSet := make(map[string]f_constraints.Constraint, 0)

			// Rule 1 -> abridge
			if flags.FuzzingMode != Guided && ch.Len() >= 2 && rand.Int()%2 == 1 {
				newCh1, newCh2 := abridge(ch)
				tSet[newCh1.String()] = newCh1
				tSet[newCh2.String()] = newCh2
			}

			// Rule 2 -> flip (not in original implementation, not in GoPie,
			// but in GoCR and GoCRHB)
			if flags.FuzzingMode != GoPie {
				if ch.Len() >= 2 && rand.Int()%2 == 1 {
					newChs := flip(ch)

					if maxMutPerStep != -1 {
						shuffle(&newChs, maxMutPerStep)
					}

					for _, newCh := range newChs {
						tSet[newCh.String()] = newCh
					}
				}
			}

			// Rule 3 -> substitute
			// if ch.len() <= bound && rand.Int()%2 == 1 {
			if rel1 != nil && rand.Int()%2 == 1 {
				newChs := substitute(ch, rel1)

				if maxMutPerStep != -1 {
					shuffle(&newChs, maxMutPerStep)
				}

				for _, newCh := range newChs {
					tSet[newCh.String()] = newCh
				}
			}

			// Rule 4 -> augment
			if rel2 != nil && ch.Len() <= bound && rand.Int()%2 == 1 {
				newChs := augment(c, rel2)

				if maxMutPerStep != -1 {
					shuffle(&newChs, maxMutPerStep)
				}

				for _, newCh := range newChs {
					tSet[newCh.String()] = newCh
				}
			}

			for k, v := range tSet {
				res[k] = v
			}
		}

		if noNew {
			break
		}

		if len(res) > mutateBound {
			break
		}

		if FuzzingModeGoPie && (rand.Int()%200) < energy {
			break
		}
	}

	// mutates selects
	for _, mut := range res {
		new := MutSelect(&mut)
		maps.Copy(res, new)
	}

	return res
}

// Abridge mutation. This creates two new mutations, where either the
// first or the last element has been removed
//
// Parameter:
//   - c chain: the chain to mutate
//
// Returns:
//   - chain: a copy of the chain with the first element removed
//   - chain: a copy of the chain with the last element removed
func abridge(c f_constraints.Constraint) (f_constraints.Constraint, f_constraints.Constraint) {
	ncHead := c.Copy()
	ncHead.RemoveHead()
	ncTail := c.Copy()
	ncTail.RemoveTail()

	return ncHead, ncTail
}

// Flip mutations. For each pair of neighboring elements in the chain, a
// new chain is created where those two elements are flipped
//
// Parameter:
//   - c chain: the chain to mutate
//
// Returns:
//   - []chain: the list of mutated chains
func flip(c f_constraints.Constraint) []f_constraints.Constraint {
	res := make([]f_constraints.Constraint, 0)

	// switch each element with the next element
	// for each flip create a new chain
	for i := 0; i < c.Len()-1; i++ {
		nc := c.Copy()
		suc := nc.Swap(i, i+1)
		if suc {
			res = append(res, nc)
		}
	}
	return res
}

// Substitute mutations. For each element create new mutations, where this
// element is replaced by an element with another trace element from the same
// routine. This new element can not be in the chain already
//
// Parameter:
//   - c chain: the chain to mutate
//   - rel1 map[trace.Element]map[trace.Element]struct{}: the rel1 info for goPie
//
// Returns:
//   - []chain: the list of mutated chains
func substitute(c f_constraints.Constraint, rel1 map[trace.Element]map[trace.Element]struct{}) []f_constraints.Constraint {
	res := make([]f_constraints.Constraint, 0)

	for i, elem := range c.Elems {
		for rel := range rel1[elem] {
			if res != nil && !c.Contains(rel) {
				nc := c.Copy()
				nc.Replace(i, rel)
				res = append(res, nc)
			}
		}
	}

	return res
}

// Augment mutations. For each element in the Rel2 set of the last element
// in the chain that is not in the chain already, created a new chain where
// this element is added at the end.
//
// Parameter:
//   - c chain: the chain to mutate
//   - rel2 map[trace.Element]map[trace.Element]struct{}: the rel2 info for goPie
//
// Returns:
//   - []chain: the list of mutated chains
func augment(c f_constraints.Constraint, rel2 map[trace.Element]map[trace.Element]struct{}) []f_constraints.Constraint {
	res := make([]f_constraints.Constraint, 0)

	if UseHBInfoFuzzing {
		concurrent := a_concurrent.GetConcurrent(c.LastElem(), true, false, settings.SameElementTypeInSC, true)
		for _, elem := range concurrent {
			if c.Contains(elem) {
				continue
			}

			nc := c.Copy()
			nc.Add(elem)
			res = append(res, nc)
		}
	} else {
		for rel := range rel2[c.LastElem()] {
			if c.Contains(rel) {
				continue
			}

			nc := c.Copy()
			nc.Add(rel)
			res = append(res, nc)
		}
	}

	return res
}

func shuffle(c *[]f_constraints.Constraint, n int) {
	if len(*c) <= n {
		return
	}

	rand.Shuffle(len(*c), func(i, j int) {
		(*c)[i], (*c)[j] = (*c)[j], (*c)[i]
	})

	*c = (*c)[:n]
}

func MutSelect(cs *f_constraints.Constraint) map[string]f_constraints.Constraint {
	res := make(map[string]f_constraints.Constraint)

	for i, elem := range cs.Elems {
		if elem.Type(false) != trace.Select {
			continue
		}

		sel := elem.(*trace.ElementSelect)
		chosen := sel.GetChosenCase()
		if chosen != nil {
			partner := sel.GetChosenCase().GetPartner()
			if partner != nil && cs.Contains(partner) {
				continue
			}
		}

		if sel.GetContainsDefault() && !sel.GetChosenDefault() {
			c := cs.Copy()
			c.Elems[i].(*trace.ElementSelect).SetCaseByIndex(-1)
			res[c.String()] = c
		}

		for ca := range sel.GetCases() {
			if ca == sel.GetChosenIndex() {
				continue
			}

			c := cs.Copy()
			c.Elems[i].(*trace.ElementSelect).SetCaseByIndex(ca)
			res[c.String()] = c
		}
	}

	return res
}

// CanBeAddedToConstraint decides if an element can be added to a scheduling chain
// For GoPie without improvements (!useHBInfoFuzzing) those are only mutex and channel (incl. select)
// With improvements those are all not ignored fuzzing elements
//
// Parameter:
//   - elem analysis.TraceElement: Element to check
//
// Returns:
//   - true if it can be added to a scheduling chain, false otherwise
func CanBeAddedToConstraint(elem trace.Element) bool {
	t := elem.Type(false)
	if flags.FuzzingMode == GoPie {
		// for standard GoPie, only mutex, channel and select operations are considered
		return (t == trace.Mutex || t == trace.Channel || t == trace.Select) && !IgnoreFuzzing(elem, true)
	}

	return !IgnoreFuzzing(elem, true)
}

// quality calculates how fit for mutation an element is
// This is based on how many times was an operation called on the same element
// and how many concurrent operation has the operations
//
// Parameters:
//   - elem trace.Element: the element to check for
//
// Returns:
//   - float64: the quality
func Quality(elem trace.Element) float64 {
	w1 := 0.2
	w2 := 0.3
	w3 := 0.5

	numberOps, _ := a_base.GetOpsPerID(elem.ResourceID())
	numberConcurrentTotal := a_concurrent.GetNumberConcurrent(elem, false, false, true)
	numberConcurrentSame := a_concurrent.GetNumberConcurrent(elem, true, false, true)

	if numberConcurrentSame == 0 && numberConcurrentTotal == 0 {
		return 0
	}

	q := w1*math.Log1p(float64(numberOps)) +
		w2*float64(numberConcurrentSame)/float64(numberConcurrentTotal+1) +
		w3*math.Log1p(float64(numberConcurrentTotal))

	return q * ((rand.Float64() * 0.2) - 0.1)
}
