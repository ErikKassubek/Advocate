// Copyright (c) 2026 Erik Kassubek
//
// File: stableClosure.go
// Brief: Calculate the stable closure for a scheduling constraint
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_constraints

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
)

const useHBClosure = true

// The stable closure is calculated based on the following three rules.z
// Assume there is a constraint s and trace T.
//
// For the order closure:
// The GCO-Closure CL_T(S) is the smallest set of events such that
// (a) S ⊆ Cl_T
// (b) for every (k,q,l,o) in T and (k',q',l',o') in Cl_T(S) where o is a commit event, q = q' and k < k' we have that (k, q, l, o) in Cl_T(S), and
// (c) for every (k,q,l,go(p')) in T and (k',q',l',o') with q' = p' we have that (k,q,l,go(p')) in Cl_t(S)
//
// For the HB closure:
// The GCO-Closure CL_T^{HB} is the smallest set of events such that
// (a) S ⊆ Cl_T(S)^{HB}
// (b) for every (k,q,l,o) in T and (k',q',l',o') in Cl_T^{HB}(S) where o is a commit event, if o <HB o' we have that (k, q, l, o) in Cl_T^{HB}(S)
//
// In general we get, that Cl_T(S) ⊆ Cl_T^{HB}(S)
//
// THe Stable closure CL is then CL_T(S) \ S
func CalcuateStableClosure(s *Constraint) *Constraint {
	if useHBClosure {
		return stableClosureHB(s)
	}
	return stableClosureOrder(s)
}

func stableClosureOrder(s *Constraint) *Constraint {
	t := a_base.MainTrace
	tIter := t.AsIterator()

	cl := NewConstraint()

	maxTs := make(map[int]int)
	forks := make(map[int]bool)
	elems := make(map[int]bool)

	for _, scElem := range s.Elems {
		rout := scElem.RoutineID()
		ts := scElem.T(trace.Commit)
		maxTs[rout] = max(maxTs[rout], ts)

		if e, ok := scElem.(*trace.ElementFork); ok {
			forks[e.ResourceID()] = true
		}

		elems[scElem.ID()] = true
	}

	newElemAdded := true
	for newElemAdded {
		newElemAdded = false

		tIter.Reset()

		// (b)
		for elem := tIter.Next(); elem != nil; elem = tIter.Next() {
			if !elem.Replayable() {
				continue
			}

			id := elem.ID()
			if elems[id] {
				continue
			}

			rout := elem.RoutineID()
			ts := elem.T(trace.Commit)
			if ts < maxTs[rout] {
				cl.Add(elem)
				elems[id] = true
				newElemAdded = true
			}
		}

		// (c)
		for rout := range maxTs {
			if forks[rout] {
				continue
			}
			forkElem := t.GetRoutine(rout).Fork()

			if forkElem == nil {
				continue
			}

			id := forkElem.ID()
			if elems[id] {
				continue
			}

			cl.Add(forkElem)
			elems[id] = true

			forks[rout] = true
			maxTs[rout] = max(maxTs[rout], forkElem.T(trace.Commit))
			newElemAdded = true
		}
	}

	cl.sort()

	return &cl
}

func stableClosureHB(s *Constraint) *Constraint {
	t := a_base.MainTrace
	tIter := t.AsIterator()

	cl := NewConstraint()

	elems := make(map[int]bool)

	for _, scElem := range s.Elems {
		elems[scElem.ID()] = true
	}

	newElemAdded := true
	for newElemAdded {
		newElemAdded = false

		tIter.Reset()

		// (b)
		for elem := tIter.Next(); elem != nil; elem = tIter.Next() {
			if !elem.Replayable() {
				continue
			}

			id := elem.ID()
			if elems[id] {
				continue
			}

			vc := elem.GetVC(a_clock.Weak)

			for _, elemConstr := range cl.Elems {
				if a_clock.GetHappensBefore(vc, elemConstr.GetVC(a_clock.Weak)) == a_hb.Before {
					cl.Add(elem)
					elems[elem.ID()] = true
					newElemAdded = true
					break
				}
			}
		}
	}

	cl.sort()

	return &cl
}
