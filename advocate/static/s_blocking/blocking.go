// Copyright (c) 2026 Erik Kassubek
//
// File: blocking.go
// Brief: Entry point for static blocking analysis
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/analysis/a_base"
	"advocate/static/static"
	"advocate/utils/flags"
	"advocate/utils/log"
)

// init to static blocking analysis
func BuildStaticBlockingAnalysis() (err error) {
	log.Info("Build static Analysis")

	data, err = static.BuildStaticData(flags.RootPath)
	if err != nil {
		return err
	}

	data.Ssa().PrintAnalysis()
	// data.Ssa().PrintSsa(true)

	// isBlockingBug()
	isBlockingSimple()

	// tr := a_base.MainTrace.AsIterator()
	// for elem := tr.Next(); elem != nil; elem = tr.Next() {
	// 	f, instr := data.Ssa().TraceToSSA(elem)
	// 	if instr != nil {
	// 		println(f.Name(), " # ", elem.ToString(), " # ", instr.String())
	// 	}
	// }

	return nil
}

func isBlockingBug() {
	blocking.blocked, blocking.blockedResources = getBlockedResources()

	for _, res := range blocking.blockedResources {
		f, s := data.Ssa().TraceToSSA(res.Alloc())

		if f == nil || s == nil {
			continue
		}

		log.Debug(res.Id(), " | ", f.Name(), " | ", s.String())
	}

	buildFuncCallToSSAFunc()

	determineResouceToSSAAtTermination()

	for rout, path := range blocking.pathPerRoutine {
		if ok, _ := a_base.MainTrace.IsRoutTerm(rout); ok {
			continue
		}
		unbl := MayUnblock(rout, path.Peek().last())

		for ub := range blocking.blocked {
			log.Resultf(true, true, "", "Found blocking bug after static: %s", ub)
		}

		for ub := range unbl {
			log.Resultf(true, true, "", "Found possible partner for: %s", ub)
		}
	}
}

// TODO: add break and label back in
func isBlockingSimple() {
	blocking.blocked, blocking.blockedResources = getBlockedResources()

	for b := range blocking.blocked {
		found := false
		// here:
		for _, f := range data.Ssa().Funcs() {
			for _, block := range f.Blocks() {
				for _, inst := range block.Instrs() {
					if inst.CanUnblockSimple(b) {
						found = true
						// break here
					}
				}
			}
		}

		if !found {
			log.Resultf(true, true, "", "Found blocking bug after static: %s", b)
		}
	}
}
