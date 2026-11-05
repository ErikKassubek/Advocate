// Copyright (c) 2026 Erik Kassubek
//
// File: stableClosure.go
// Brief: Calculate the closure for a scheduling constraint
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
	"advocate/utils/flags"
	"advocate/utils/log"
)

const mustHB = a_clock.Strong

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
func CalcuateClosure(s *Constraint) *Constraint {
	switch flags.ClosureType {
	case "order":
		return calculateClosureOrder(s)
	case "hbc":
		return calculateClosureHB(s, mustHB)
	case "scpc":
		return calculateClosureSCPC(s, mustHB)
	case "replay":
		return calculateClosureReplay(s)
	default:
		panic("Unknown closure type " + flags.ClosureType)
	}

}

func calculateClosureOrder(s *Constraint) *Constraint {
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

// Implement the stable closure based on the HB order
//
// HBC(t1,...,tn) is the smallest set with the following properties:
//  1. Each ti is an element in SCPC(t1,...,tn).
//  2. If e in SCPC(t1,...,tn) and f <(M)HB e then f in SCPC(t1,...,tn).
//
// Parameter:
//   - s *Constraint: the constraint
//   - vcType vcType a_clock.VcType: weak or string
//
// Return:
//   - *Constraint: The closue constraint
func calculateClosureHB(s *Constraint, vcType a_clock.VcType) *Constraint {
	res := calculateClosureHB2(s, vcType)

	res.sort()

	return res
}

// Implement the stable closure based on the HB order
//
// HBC(t1,...,tn) is the smallest set with the following properties:
//  1. Each ti is an element in SCPC(t1,...,tn).
//  2. If e in SCPC(t1,...,tn) and f <(M)HB e then f in SCPC(t1,...,tn).
//
// Parameter:
//   - s *Constraint: the constraint
//   - vcType vcType a_clock.VcType: weak or string
//
// Return:
//   - *Constraint: The closue constraint
func calculateClosureHB2(s *Constraint, vcType a_clock.VcType) *Constraint {
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

			vc := elem.VC(vcType)

			for _, elemConstr := range cl.Elems {
				if a_clock.GetHappensBefore(vc, elemConstr.VC(vcType)) == a_hb.Before {
					cl.Add(elem)
					elems[elem.ID()] = true
					newElemAdded = true
					break
				}
			}
		}
	}

	return &cl
}

// Implement the stable closure based on the Sync-preserving closure
//
// SCPC(t1,...,tn) is the smallest set with the following properties:
//  1. Each ti is an element in SCPC(t1,...,tn).
//  2. If e in SCPC(t1,...,tn) and f <(M)HB e then f in SCPC(t1,...,tn).
//  3. If a1, a2 in SCPC(t1,...,tn) where a1=(_,_,_,acq(m)) and a2=(_,_,_,acq(m)) and a1 <T a2 then match(a1) SCPC(t1,...,tn)
//     where match(a1) refers to the matching release of r1.
//
// Parameter:
//   - s *Constraint: the constraint
//   - vcType vcType a_clock.VcType: weak or string
//
// Return:
//   - *Constraint: The closue constraint
func calculateClosureSCPC(s *Constraint, vcType a_clock.VcType) *Constraint {
	res := calculateClosureHB2(s, vcType)

	for i, a1 := range s.Elems {
		added := false

		a1M, ok := a1.(*trace.ElementMutex)
		if !ok {
			continue
		}
		if !a1M.IsAcq() {
			continue
		}

		for j := i + 1; j < s.Len(); j++ {
			a2 := s.At(j)
			if a1.ResourceID() != a2.ResourceID() {
				continue
			}

			a2M, ok := a2.(*trace.ElementMutex)
			if !ok {
				continue
			}

			if !a2M.IsAcq() {
				continue
			}

			rel := a1M.Partner()
			if rel == nil {
				continue
			}

			res.Add(rel)
			added = true
			break
		}
		if added {
			continue
		}
	}

	res.sort()

	return res
}

func calculateClosureReplay(s *Constraint) *Constraint {
	var firstTime = -1

	for _, e := range s.Elems {
		if firstTime == -1 {
			firstTime = e.T(trace.Commit)
			continue
		}

		eTime := e.T(trace.Commit)
		if eTime < firstTime {
			firstTime = eTime
		}
	}

	closure := NewConstraint()

	if firstTime == -1 {
		return &closure
	}

	tr, err := a_base.MainTrace.Copy(false)
	if err != nil {
		log.Error(err.Error())
		return &closure
	}

	tr.RemoveLater(firstTime)

	return TraceToConstraint(&tr)
}
