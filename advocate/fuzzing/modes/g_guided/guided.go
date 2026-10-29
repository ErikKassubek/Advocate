// Copyright (c) 2025 Erik Kassubek
//
// File: mutation.go
// Brief: ROC Creation and mutation
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_guided

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_concurrent"
	"advocate/fuzzing/f_base"
	"advocate/fuzzing/f_constraints"
	"advocate/fuzzing/modes/f_gfuzz"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/settings"
	"math/rand"
)

// CreateMutations will create randomized mutations.
// This includes both situations, where the predictive analysis directly
// finds a potential bug and where we use the happens before analysis
// to guide the mutation in a better direction.
// Mutations based on predictive analysis are automatically created when a potential
// bug is detected
func CreateMutations() {
	numberMuts = 0

	// drop runs where the mutation could not be fully satisfied
	if a_base.GetTimeoutHappened(false) {
		return
	}

	traceID++

	// select based mutations
	f_gfuzz.CreateMutations(true)

	// add new original trace to equivalence
	// minTrace := equivalence.TraceEqFromTrace(&baseA.MainTrace)
	// equivalence.AddOrig(minTrace, traceID)

	numTry := 0

	constraint := startConstraint(maxNumberConstraints, lengthConstraint)
	for _, c := range constraint {
		for numberMuts < maxNumberOfMutsPerConst || numTry > maxTries {
			mutatedConstr := f_base.Mutate(c, -1, nil, nil)

			for _, cr := range mutatedConstr {
				f_base.TotalRuns++
				numTry++

				if numTry > maxTries {
					return
				}

				// minTrace := equivalence.TraceEqFromConstraint(ch)

				// if minTrace.IllFormedImpossible {
				// 	baseF.IllFormed++
				// 	continue
				// }

				if f_constraints.IsEquivalent(cr) {
					f_base.Equiv++
					if rand.Float64() < propToSkipEquiv {
						continue
					}
				}
				f_base.TotalRuns++

				firstMut := f_base.NumberFuzzingRuns <= 1 && numberMuts == 0
				_, err := f_base.WriteMutConstraint(&a_base.MainTrace, &cr, firstMut)
				if err != nil {
					log.Error("Error in writing mutation: ", err.Error())
				}
				numberMuts++
			}
		}
	}
}

// StartConstraint returns a constraint of given length, consisting of consecutive
// elements from the trace
//
// Parameter:
//   - num int: number of constraints tob be created
//   - length int: max number of elements in the constraint
//
// Returns:
//   - []baseF.Constraint: a set of new constraint consisting of consecutive elements from the constraint
func startConstraint(num, length int) []f_constraints.Constraint {
	res := make([]f_constraints.Constraint, 0)

	routines := a_base.MainTrace.Routines()

	if len(routines) == 0 {
		return res
	}

	top := []f_constraints.ElemWithQual{}

	alreadyAdded := make(map[int]struct{})

	for i := 0; i < 1000; i++ {
		key := rand.Intn(len(routines)) + 1
		rout := routines[key]
		if rout.Empty() {
			continue
		}

		ind := rand.Intn(rout.Len())
		elem := rout.At(ind)

		if _, ok := alreadyAdded[elem.T(trace.Commit)]; ok {
			continue
		}
		alreadyAdded[elem.T(trace.Commit)] = struct{}{}

		if !f_base.CanBeAddedToConstraint(elem) {
			continue
		}

		sameElem := true
		if a_concurrent.GetNumberConcurrent(elem, sameElem, settings.SameElementTypeInSC, false) == 0 {
			continue
		}

		q := f_base.Quality(elem)

		e := f_constraints.ElemWithQual{Elem: elem, Quality: q}

		// find the num with the best quality
		inserted := false
		for i, v := range top {
			if e.Quality > v.Quality {
				top = append(top[:i+1], top[i:]...)
				top[i] = e
				inserted = true
				break
			}
		}
		if !inserted && len(top) < num {
			top = append(top, e)
		}
	}

	if len(top) == 0 {
		return res
	}

	if len(top) > num {
		top = top[:num]
	}

	for _, e := range top {
		c := f_constraints.NewConstraint()
		c.Add(e.Elem)

		for i := 0; i < length; i++ {
			posNext := a_concurrent.GetConcurrent(c.LastElem(), true, true, settings.SameElementTypeInSC, true)
			if len(posNext) == 0 {
				posNext = a_concurrent.GetConcurrent(c.LastElem(), true, false, settings.SameElementTypeInSC, true)
				if len(posNext) == 0 {
					break
				}
			}

			concToAll := make([]trace.Element, 0)

			for _, next := range posNext {
				isPos := true
				for _, e := range c.Elems {
					if !a_concurrent.IsConcurrentWeak(next, e) {
						isPos = false
						break
					}
				}
				if isPos {
					concToAll = append(concToAll, next)
				}
			}

			if len(concToAll) == 0 {
				break
			}

			next := concToAll[rand.Intn(len(concToAll))]
			c.Add(next)
		}

		res = append(res, c)
	}

	return res
}
