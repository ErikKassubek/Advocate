// Copyright (c) 2025 Erik Kassubek
//
// File: goPie-sc.go
// Brief: scheduling Chains for GoPie
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gopie

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_concurrent"
	"advocate/fuzzing/f_base"
	"advocate/fuzzing/f_constraints"
	"advocate/utils/settings"
	"math/rand"
)

// startChains returns a slice of chain consisting of a
// pair of operations that are in a rel2 relation
//
// Parameter:
//   - num int: number of chains to return
//
// Returns:
//   - the chain, or an empty chain if no pair exists
func startChains(num int) []f_constraints.Constraint {
	res := make([]f_constraints.Constraint, 0)

	if f_base.UseHBInfoFuzzing {
		routines := a_base.MainTrace.Routines()

		if len(routines) == 0 {
			return res
		}

		top := []f_constraints.ElemWithQual{}

		for i := 0; i < 1000; i++ {
			key := rand.Intn(len(routines)) + 1

			routine := routines[key]
			if routine.Empty() {
				continue
			}

			ind := rand.Intn(routine.Len())
			elem := routine.At(ind)

			if !f_base.CanBeAddedToConstraint(elem) {
				continue
			}

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

			// Trim if longer than n
			if len(top) > num {
				top = top[:num]
			}
		}

		if len(top) == 0 {
			return res
		}

		rounds := settings.GoPieMaxSCLength
		for i := 0; i < rounds; i++ {
			if len(res) == 0 {
				for _, e := range top {
					posPartner := a_concurrent.GetConcurrent(e.Elem, true, true, settings.SameElementTypeInSC, true)
					if len(posPartner) == 0 {
						posPartner = a_concurrent.GetConcurrent(e.Elem, true, false, settings.SameElementTypeInSC, true)
						if len(posPartner) == 0 {
							continue
						}
					}

					partner := posPartner[rand.Intn(len(posPartner))]
					c := f_constraints.NewConstraint()
					c.Add(e.Elem, partner)
					res = append(res, c)
				}
			} else {
				for i, c := range res {
					lastElem := c.LastElem()
					if lastElem == nil {
						continue
					}

					posNext := a_concurrent.GetConcurrent(lastElem, true, true, settings.SameElementTypeInSC, true)
					if len(posNext) == 0 {
						continue
					}
					next := posNext[rand.Intn(len(posNext))]
					res[i].Add(next)
				}
			}
		}
	} else {
		// start with two random elements in rel2
		i := 0
		for elem1, rel := range rel2 {
			for elem2 := range rel {
				c := f_constraints.NewConstraint()
				c.Add(elem1, elem2)
				res = append(res, c)
				i++
				if i > num {
					return res
				}
			}
		}
	}

	return res
}
