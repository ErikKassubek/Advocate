// Copyright (c) 2025 Erik Kassubek
//
// File: goPie.go
// Brief: Main file for goPie fuzzing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gopie

import (
	"gocct/analysis/a_base"
	"gocct/analysis/hb/a_concurrent"
	"gocct/fuzzing/f_base"
	"gocct/trace"
	"gocct/utils/flags"
	"gocct/utils/log"
	"gocct/utils/settings"
	"math"
)

const sameElem = true

// CreateMutations create new mutations for GoPie
//
// Parameter:
//   - mutNumber int: number of the mutation file
//   - error
func CreateMutations(mutNumber int) error {
	mutations := make(map[string]f_base.Constraint)
	specMutations := make(map[string]f_base.Constraint) // special mutations that should be run first

	// check for special chains, that could indicate a bug
	if flags.FuzzingMode != f_base.GoPie && f_base.UseHBInfoFuzzing {
		specialMuts := getSpecialMuts()

		for key, mut := range specialMuts {
			isValid := mut.IsValid()
			if _, ok := allGoPieMutations[key]; !ok {
				if !f_base.UseHBInfoFuzzing || isValid {
					specMutations[key] = mut
				}

				if !isValid {
					NumberInvalidMuts++
				}
				NumberTotalMuts++

				if ok {
					NumberDoubleMuts++
				}

				allGoPieMutations[key] = struct{}{}
			}
		}
	}

	// Original GoPie does not mutate all possible scheduling chains
	// If no SC is given, it creates a new one consisting of two random
	// operations that are in rel2 relation. Otherwise it always mutates the
	// original SC, not newly recorded once
	SchedulingChains = []f_base.Constraint{}
	if flags.FuzzingMode == f_base.GoPie {
		if c, ok := f_base.ChainFiles[mutNumber]; ok {
			c.Old = true
			SchedulingChains = []f_base.Constraint{c}
		}
	}

	if flags.FuzzingMode != f_base.GoPie || len(SchedulingChains) == 0 {
		sc := startChains(maxSCStart)
		for _, c := range sc {
			if c.Len() > 1 {
				SchedulingChains = append(SchedulingChains, c)
			}
		}
	}

	energy := getEnergy()

	log.Infof("Mutate %d scheduling chains", len(SchedulingChains))

	for _, sc := range SchedulingChains {
		muts := f_base.Mutate(sc, energy, rel1, rel2)

		for key, mut := range muts {
			if flags.FuzzingMode != f_base.GoPie && mut.Len() <= 1 {
				NumberTotalMuts++
				continue
			}
			if _, ok := allGoPieMutations[key]; flags.FuzzingMode == f_base.GoPie || !ok {
				// only add if not invalidated by hb
				isValid := mut.IsValid()
				if !f_base.UseHBInfoFuzzing || mut.IsValid() {
					mutations[key] = mut
				}

				if !isValid {
					NumberInvalidMuts++
				}
				NumberTotalMuts++
				allGoPieMutations[key] = struct{}{}
			} else if flags.FuzzingMode == f_base.GoPie && !ok {
				NumberDoubleMuts++
			}
		}
	}

	if len(specMutations) > 0 {
		log.Infof("Write %d special mutation to file", len(specMutations))
	}

	if f_base.MaxNumberRuns > 0 {
		log.Infof("Write %d mutations to file", max(0, min(len(mutations)+len(specMutations), f_base.MaxNumberRuns-f_base.NumberWrittenMutations)))
	} else {
		log.Infof("Write %d mutations to file", max(0, len(mutations)+len(specMutations)))
	}

	first := f_base.NumberFuzzingRuns <= 1

	for _, mut := range specMutations {
		done, err := f_base.WriteMutConstraint(mut, first)
		first = false

		if done { // max number mutations has been reached
			break
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	for _, mut := range mutations {
		done, err := f_base.WriteMutConstraint(mut, first)
		first = false

		if done { // max number mutations has been reached
			break
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

// Calculate the energy for a schedule. This determines how many mutations
// are created
func getEnergy() int {

	// not interesting
	if a_base.GetTimeoutHappened(false) {
		return 0
	}

	w1 := settings.GoPieW1
	w2 := settings.GoPieW2

	score := 0

	if f_base.UseHBInfoFuzzing {
		for _, sc := range SchedulingChains {
			for _, elem := range sc.Elems {
				c := a_concurrent.GetNumberConcurrent(elem, sameElem, false, true)
				score += c
			}
		}
	} else {
		score = int(w1*float64(counterCPOP1) + w2*math.Log(float64(counterCPOP2)))
	}

	if score > maxGoPieScore {
		maxGoPieScore = score
	}

	return int(float64(score+1)/float64(maxGoPieScore)) * 100
}

// Pass the trace and look for
//
//	channel close with concurrent send on the same channel
//
// # Based on those, create chains where the close if before the send
//
// Returns:
//   - map[string]Chain: map with the special chains
func getSpecialMuts() map[string]f_base.Constraint {
	res := make(map[string]f_base.Constraint)

	// send on closed
	for _, c := range a_base.CloseData {
		conc := a_concurrent.GetConcurrent(c, true, false, true, false)
		for _, s := range conc {
			switch t := s.(type) {
			case *trace.ElementSelect:
				for _, cc := range t.GetCases() {
					if cc.Type(true) == trace.ChannelSend {
						chain := f_base.NewConstraint()
						chain.Add(c, s)
						res[chain.ToString()] = chain
					}
				}
			default:
				if s.Type(true) == trace.ChannelSend {
					chain := f_base.NewConstraint()
					chain.Add(c, s)
					res[chain.ToString()] = chain
				}
			}
		}
	}

	// negative wg counter
	for id, dones := range a_base.WgDoneData {
		for _, done := range dones {
			for _, add := range a_base.WGAddData[id] {
				if add.T(trace.Commit) > done.T(trace.Commit) {
					continue
				}

				if a_concurrent.IsConcurrent(done, add) {
					chain := f_base.NewConstraint()
					chain.Add(done, add)
					res[chain.ToString()] = chain
				}
			}
		}
	}

	return res
}
