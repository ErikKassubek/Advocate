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
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/results/buginfo"
	"advocate/utils/results/results"
)

// Check if static data has already been build
func StaticBuild() bool {
	return data != nil
}

// init to static blocking analysis
func BuildStaticBlockingAnalysis() (err error) {
	log.Info("Build static Analysis")

	data, err = static.BuildStaticData(flags.RootPath)
	if err != nil {
		return err
	}

	// data.Ssa().PrintAnalysis()
	// data.Ssa().PrintSsa(true)

	// tr := a_base.MainTrace.AsIterator()
	// for elem := tr.Next(); elem != nil; elem = tr.Next() {
	// 	f, instr := data.Ssa().TraceToSSA(elem)
	// 	if instr != nil {
	// 		println(f.Name(), " # ", elem.ToString(), " # ", instr.String())
	// 	}
	// }

	return nil
}

// Check for blocking bugs based on static
// Blocking bugs are directly reported
//
// Parameter:
//   - routs map[int]bool: only consider blocked operations in the given routines
//
// Returns:
//   - []int: routines that are blocking bugs
//   - []int: routines that are leaks
func CheckBlocking(routs map[int]bool) ([]int, []int) {
	log.Info("Check Blocking with Static")
	log.Debug2("Check Blocking Static")
	// return isBlockingBug(routs)
	return isBlockingSimple(routs)
}

// Check for blocking bugs based on static
// Blocking bugs are directly reported
//
// Parameter:
//   - routs map[int]bool: only consider blocked operations in the given routines
//
// Returns:
//   - []int: routines that are blocking bugs
//   - []int: routines that are leaks
func isBlockingBug(routs map[int]bool) ([]int, []int) {
	blocking.blocked, blocking.blockedResources = getBlockedResources(routs)

	for _, res := range blocking.blockedResources {
		f, s := data.Ssa().TraceToSSA(res.Alloc())

		if f == nil || s == nil {
			continue
		}

		log.Debug(res.Id(), " | ", f.Name(), " | ", s.String())
	}

	buildFuncCallToSSAFunc()

	determineResouceToSSAAtTermination()

	bug := make([]int, 0)
	leak := make([]int, 0)

	for rout, path := range blocking.pathPerRoutine {
		if ok, _ := a_base.MainTrace.IsRoutTerm(rout); ok {
			continue
		}
		unbl := MayUnblock(rout, path.Peek().last())

		for ub := range blocking.blocked {
			arg1 := results.TraceElementResult{
				RoutineID:  ub.RoutineID(),
				ResourceID: ub.ResourceID(),
				TRequest:   ub.T(trace.Request),
				ObjType:    ub.Type(true),
				File:       ub.File(),
				Line:       ub.Line(),
			}

			bug = append(bug, ub.RoutineID())
			results.Result(results.CRITICAL, buginfo.SBlocking,
				[]results.ResultElem{arg1}, []results.ResultElem{})
		}

		for ub := range unbl {
			leak = append(leak, ub.RoutineID())
			log.Resultf(true, true, "", "Found possible partner for: %s", ub)
		}
	}

	return bug, leak
}

// Check for blocking bugs based on static
// Blocking bugs are directly reported
//
// Parameter:
//   - routs map[int]bool: only consider blocked operations in the given routines
//
// Returns:
//   - []int: routines that are blocking bugs
//   - []int: routines that are leaks
func isBlockingSimple(routs map[int]bool) ([]int, []int) {
	log.Debug2("IS BLOCKING SIMPLE: ", blocking.blocked)

	bug := make([]int, 0)
	leak := make([]int, 0)

	tr := &a_base.MainTrace

	for r := range routs {
		e := tr.GetLastElemInRout(r)
		found := false
	here:
		for _, f := range data.Ssa().Funcs() {
			for _, block := range f.Blocks() {
				for _, inst := range block.Instrs() {
					if inst.CanUnblockSimple(e) {
						found = true
						break here
					}
				}
			}
		}

		log.Debug2("AFTER HERE")

		if !found {
			bug = append(bug, e.RoutineID())
			log.Resultf(true, true, "", "Found blocking bug after static: %s", e)
		} else {
			leak = append(leak, e.RoutineID())
		}
	}

	return bug, leak
}
