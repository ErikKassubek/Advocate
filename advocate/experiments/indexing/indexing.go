// Copyright (c) 2026 Erik Kassubek
//
// File: indexing.go
// Brief: Implement id abstraction based on light-weight execution indexing
//        Based on P. Joshi, C.-S. Park, K. Sen, and M. Naik, “A randomized dynamic program analysis technique for detecting real deadlocks,” SIGPLAN Not., vol. 44, no. 6, pp. 110–120, Jun. 2009, doi: 10.1145/1543135.1542489.
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package indexing

import (
	"advocate/analysis/a_base"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
	"fmt"
	"strings"
)

var totalPair = 0
var totalEqual = 0

const k = 15

type stackElem struct {
	isC bool
	c   string
	q   int
}

func (this *stackElem) String() string {
	if this.isC {
		return this.c
	}

	return fmt.Sprintf("%d", this.q)
}

var depth = make(map[int]int)                         // rout id -> depth
var counters = make(map[int]map[int]map[string]int)   //  rout id -> d_t -> c (pos string) -> value
var callStack = make(map[int]*types.Stack[stackElem]) // rout id -> call stack of c

var indices = make(map[trace.Element]string)

func initValues(tr *trace.Trace) {
	depth = make(map[int]int)                       // rout id -> depth
	counters = make(map[int]map[int]map[string]int) //  rout id -> d_t -> c (pos string) -> value
	callStack = make(map[int]*types.Stack[stackElem])
	indices = make(map[trace.Element]string)

	for rout := range tr.Routines() {
		depth[rout] = 0
		counters[rout] = make(map[int]map[string]int)
		callStack[rout] = types.NewStack[stackElem]()
	}
}

func pushToCS(routID, d int, c string) {
	c_stack_elem := stackElem{true, c, 0}
	q_stacl_elem := stackElem{false, "", counters[routID][d][c]}
	callStack[routID].Push(c_stack_elem)
	callStack[routID].Push(q_stacl_elem)
}

func BuildExecutionIndexing() {
	tr := &a_base.MainTrace
	traceIter := tr.AsIterator()

	initValues(tr)

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		routID := elem.RoutineID()
		// c := elem.Pos().Short()
		c := fmt.Sprintf("%d", elem.Line())
		d := depth[routID]

		if _, ok := counters[routID][d]; !ok {
			counters[routID][d] = make(map[string]int)
		}

		switch elem.(type) {
		case *trace.ElementFunc:
			updateCall(routID, d, c)
		case *trace.ElementReturn:
			depth[routID] = depth[routID] - 1
			callStack[routID].Pop()
			callStack[routID].Pop()
		case *trace.ElementAlloc:
			updateAlloc(elem, routID, d, c)
		case *trace.ElementFork:
			updateAlloc(elem, routID, d, c)
			updateCall(routID, d, c)
		}
	}

}

func updateCall(routID, d int, c string) {
	counters[routID][d][c] = counters[routID][d][c] + 1

	pushToCS(routID, d, c)

	depth[routID] = depth[routID] + 1

	d = depth[routID]
	for c_iter := range counters[routID][d] {
		counters[routID][d][c_iter] = 0
	}
}

func updateAlloc(elem trace.Element, routID, d int, c string) {
	counters[routID][d][c] = counters[routID][d][c] + 1

	pushToCS(routID, d, c)

	arr := callStack[routID].AsArray()

	var index strings.Builder

	startVal := max(0, len(arr)-2*k)
	for i := startVal; i < len(arr); i++ {
		if i != startVal {
			index.WriteString(",")
		}
		val := arr[i]
		index.WriteString(val.String())

	}

	indices[elem] = index.String()

	callStack[routID].Pop()
	callStack[routID].Pop()
}

const ignoreInternal = true
const printPairs = true

func CheckForEq() {
	pairCount := 0
	allPairs := 0

	keys := make([]trace.Element, 0, len(indices))
	for i := range indices {
		keys = append(keys, i)
	}

	for i, e1 := range keys {
		i1 := indices[e1]

		_, okFork1 := e1.(*trace.ElementFork)
		_, okAlloc1 := e1.(*trace.ElementAlloc)

		if !okFork1 && !okAlloc1 {
			continue
		}

		if ignoreInternal && strings.Contains(e1.File(), "/Advocate/goPatch/src") {
			continue
		}

		for j := i + 1; j < len(keys); j++ {
			e2 := keys[j]
			i2 := indices[e2]

			_, okFork2 := e2.(*trace.ElementFork)
			_, okAlloc2 := e2.(*trace.ElementAlloc)
			if !okFork2 && !okAlloc2 {
				continue
			}

			// if e1.Type(false) != e2.Type(false) {
			// 	continue
			// }

			// if e1.Pos() != e2.Pos() {
			// 	continue
			// }

			if ignoreInternal && strings.Contains(e2.File(), "/Advocate/goPatch/src") {
				continue
			}

			if i1 == i2 {
				if printPairs {
					log.Debugf2("Same Index for different Objects:\n%s -> %s\n%s -> %s", e1.StringDebug(), i1, e2.StringDebug(), i2)
				}
				pairCount += 1
			}
			allPairs += 1
		}
	}

	totalPair += allPairs
	totalEqual += pairCount

	if pairCount != 0 {
		log.Debugf2("Found Index Violation Pairs: %d (%d)", pairCount, allPairs)
	} else {
		log.Debugf2("Found No Violation Pairs (%d)", allPairs)
	}

	var perc float64 = 0
	if totalPair != 0 {
		perc = float64(totalEqual) / float64(totalPair) * 100
		log.Debugf2("Total: %d / %d (%f %%)", totalEqual, totalPair, perc)
	} else {
		log.Debugf2("Total: 0")
	}

}

func PrintIndexes() {
	log.Debug("INDICES")
	for elem, index := range indices {
		log.Debug(elem.StringDebug(), " -> ", index)
	}
}
