// Copyright (c) 2024 Erik Kassubek
//
// File: blocked.go
// Brief: Trace analysis for routine blocks
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/trace"
	"advocate/utils/results/buginfo"
	"advocate/utils/results/results"
	"advocate/utils/types"
	"strings"
)

func Blocked() error {
	tr := &a_base.MainTrace
	blocked := tr.GetBlocked()

	// blocked routines
	b := make(map[int]struct{})
	i := 0
	for rout := range blocked {
		b[rout] = struct{}{}
		i++
	}

	l := tr.GetNotReturned(true)

	for {
		r := make([]int, 0)

		for routB := range b {
			for _, routL := range l {
				found := false
				for _, blockedB := range tr.GetResourcesPerRout(routB) {
					if types.Contains(tr.GetResourcesPerRout(routL), blockedB) {
						r = append(r, routB)
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}

		for _, routR := range r {
			l = append(l, routR)
			delete(b, routR)
		}

		if len(r) == 0 {
			break
		}
	}

	del := checkSelectDefaultOrTImer(tr, b)
	if len(del) != 0 {
		l = types.MergeLists(l, del)
	}

	cyclic := checkCyclic(b, tr.GetResourcesRout())

	for rout := range cyclic {
		delete(b, rout)
	}

	reportBlocking(cyclic, buginfo.ADeadlock)
	reportBlocking(b, buginfo.ABlocking)
	reportLeak(l)

	return nil
}

func checkSelectDefaultOrTImer(tr *trace.Trace, b map[int]struct{}) []int {
	del := make([]int, 0)

	for rout := range b {
		last := tr.GetLastElemInRout(rout)
		if elem, ok := last.(*trace.ElementSelect); ok {
			if elem.GetContainsDefault() {
				del = append(del, rout)
				continue
			}
			if elem.HasTimer() {
				del = append(del, rout)
				continue
			}
		}
		if elem, ok := last.(*trace.ElementChannel); ok {
			if elem.IsTimer() {
				del = append(del, rout)
				continue
			}
		}
	}

	for _, d := range del {
		delete(b, d)
	}

	return del
}

// check for cyclic dependencies
func checkCyclic(b map[int]struct{}, res map[int][]trace.Resource) map[int]struct{} {
	graph := map[int][]int{}
	selfLoop := map[int]bool{}

	for rID := range b {
		for rID2 := range b {
			if types.HasCommonElement(res[rID], res[rID2]) {

				graph[rID] = append(graph[rID], rID2)
				if rID == rID2 {
					selfLoop[rID] = true
				}
			}
		}
	}

	// Tarjan SCC
	index := 0
	stack := []int{}
	onStack := map[int]bool{}
	indices := map[int]int{}
	lowlink := map[int]int{}

	result := map[int]struct{}{}

	var strongConnect func(int)
	strongConnect = func(v int) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range graph[v] {
			if _, seen := indices[w]; !seen {
				strongConnect(w)
				lowlink[v] = min(lowlink[v], lowlink[w])
			} else if onStack[w] {
				lowlink[v] = min(lowlink[v], indices[w])
			}
		}

		// Root of SCC
		if lowlink[v] == indices[v] {
			scc := map[int]struct{}{}
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc[w] = struct{}{}
				if w == v {
					break
				}
			}

			// must be cycle with multiple elements
			if len(scc) == 1 {
				return
			}

			// must be closed
			for r := range scc {
				neigh := graph[r]
				for _, n := range neigh {
					if _, ok := scc[n]; !ok {
						return
					}
				}
			}

			for r := range scc {
				result[r] = struct{}{}
			}
		}
	}

	for r := range b {
		if _, seen := indices[r]; !seen {
			strongConnect(r)
		}
	}

	return result
}

func reportBlocking(routs map[int]struct{}, rt buginfo.ResultType) {
	obj := make([]results.ResultElem, 0)
	tr := &a_base.MainTrace

	for r := range routs {
		elem := tr.GetLastElemInRout(r)

		objRes := results.TraceElementResult{
			RoutineID:  r,
			ResourceID: -1,
			TRequest:   -1,
			ObjType:    elem.Type(true),
			File:       elem.File(),
			Line:       elem.Line(),
		}

		obj = append(obj, objRes)
	}

	if len(obj) != 0 {
		results.Result(results.CRITICAL, rt, obj, nil)
	}
}

func reportLeak(l []int) {
	tr := &a_base.MainTrace
	for _, routID := range l {
		rout := tr.GetRoutine(routID)

		if rout.IsTerminated() {
			continue
		}

		last := rout.Last()

		if last == nil {
			continue
		}

		// go tRunner(t, f) for tests
		if strings.HasSuffix(last.File(), "/goPatch/src/testing/testing.go") && last.Line() == 2262 {
			continue
		}

		objRes := results.TraceElementResult{
			RoutineID:  routID,
			ResourceID: last.ResourceID(),
			TRequest:   last.T(trace.Request),
			ObjType:    last.Type(true),
			File:       last.File(),
			Line:       last.Line(),
		}

		leakType := buginfo.LUnknown

		if !last.Committed() {
			switch last := last.(type) {
			case *trace.ElementChannel:
				if last.ResourceID() == 0 {
					leakType = buginfo.LNilChan
				} else {
					if last.GetSelect() == nil {
						leakType = buginfo.LChan
					} else {
						leakType = buginfo.LSelect
					}
				}
			case *trace.ElementSelect:
				leakType = buginfo.LSelect
			case *trace.ElementMutex:
				leakType = buginfo.LMutex
			case *trace.ElementWait:
				leakType = buginfo.LWaitGroup
			case *trace.ElementCond:
				leakType = buginfo.LCond
			}
		}

		results.Result(results.CRITICAL, leakType, []results.ResultElem{objRes}, nil)

	}
}
