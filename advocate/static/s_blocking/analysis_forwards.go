// Copyright (c) 2026 Erik Kassubek
//
// File: forward_analysis.go
// Brief: Starting from the point of termination, determine if in the future, an unblocking operation to a blocked operation exists
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
)

const FORWARD = true

type potenitalUnblock struct {
	routine     int
	blockedElem trace.Element
	resource    trace.Resource
	unblock     *instructionWithInfo
	execPath    path
}

func MayUnblock(routine int, start *instructionWithInfo) map[trace.Element][]potenitalUnblock {
	result := make(map[trace.Element][]potenitalUnblock)

	type workItem struct {
		pp       s_ssa.Instruction
		execPath path
	}

	initial := workItem{pp: start.Inst, execPath: path{start}}
	workQueue := types.NewQueue[workItem]()
	workQueue.Push(initial)
	visited := make(map[s_ssa.Instruction][]path) // TODO: add to visited

	for !workQueue.IsEmpty() {
		item := workQueue.Pop()
		point := item.pp
		path := item.execPath

		nexts, alias := parseAllPaths(point, routine)

		for _, next := range nexts {
			if next == nil {
				continue
			}

			newExecPath := append(path, alias)
			wi := workItem{
				pp:       next,
				execPath: newExecPath,
			}

			if old, ok := visited[point]; ok {
				foundOld := false
				for _, path := range old {
					if pathSetEqual(path, newExecPath) {
						foundOld = true
						break
					}
				}
				if foundOld {
					continue
				}
			}

			visited[point] = append(visited[point], newExecPath)

			workQueue.Push(wi)
		}

		for elem := range blocking.blocked {
			if comp, r := compatible(alias, elem); comp {
				pu := potenitalUnblock{
					routine:     routine,
					blockedElem: elem,
					resource:    *r,
					unblock:     alias,
					execPath:    path,
				}
				if _, ok := result[elem]; !ok {
					result[elem] = make([]potenitalUnblock, 0)
				}
				result[elem] = append(result[elem], pu)

				log.Debug("Potential Unlock: ", elem)

				delete(blocking.blocked, elem)

				if len(blocking.blocked) == 0 {
					return result
				}
			}
		}

	}

	// TODO: recheck second order channels/channels on mutex

	return result

}

func pathSetEqual(old, current path) bool {
	for _, inst := range current {
		for _, inst2 := range old {
			if inst.Variable == inst2.Variable && inst.sameResource(inst2) {
				return true
			}
		}
	}

	return false
}

func parseAllPaths(inst s_ssa.Instruction, rout int) ([]s_ssa.Instruction, *instructionWithInfo) {
	if inst == nil {
		return nil, nil
	}

	log.Debug("FORWARD: ", inst)

	var resInst s_ssa.Instruction
	var resIWI *instructionWithInfo

	switch inst := inst.(type) {
	case *s_ssa.InstructionAlloc:
		resInst, resIWI = ParseAlloc(inst, rout, nil)
	case *s_ssa.InstructionCall:
		resInst, resIWI = ParseCall(inst, rout, nil)
	case *s_ssa.InstructionExtract:
		resInst, resIWI = ParseExtract(inst, rout, nil)
	case *s_ssa.InstructionField:
		resInst, resIWI = ParseField(inst, rout, nil)
	case *s_ssa.InstructionFieldAddr:
		resInst, resIWI = ParseFieldAddr(inst, rout, nil)
	case *s_ssa.InstructionGo:
		resInst, resIWI = ParseGo(inst, rout, nil)
	case *s_ssa.InstructionIf:
		return ParseIfAll(inst), nil
	case *s_ssa.InstructionIndex:
		resInst, resIWI = ParseIndex(inst, rout, nil)
	case *s_ssa.InstructionIndexAddr:
		resInst, resIWI = ParseIndexAddr(inst, rout, nil)
	case *s_ssa.InstructionJump:
		resInst, resIWI = ParseJump(inst, rout, nil)
	case *s_ssa.InstructionLookup:
		resInst, resIWI = ParseLookup(inst, rout, nil)
	case *s_ssa.InstructionMakeChan:
		resInst, resIWI = ParseMakeChan(inst, rout, nil)
	case *s_ssa.InstructionMakeClosure:
		resInst, resIWI = ParseMakeClosure(inst, rout, nil)
	case *s_ssa.InstructionMakeInterface:
		resInst, resIWI = ParseMakeInterface(inst, rout, nil)
	case *s_ssa.InstructionMakeMap:
		resInst, resIWI = ParseMakeMap(inst, rout, nil)
	case *s_ssa.InstructionMakeSlice:
		resInst, resIWI = ParseMakeSlice(inst, rout, nil)
	case *s_ssa.InstructionMapUpdate:
		resInst, resIWI = ParseMapUpdate(inst, rout, nil)
	case *s_ssa.InstructionNext:
		resInst, resIWI = ParseNext(inst, rout, nil)
	case *s_ssa.InstructionPhi:
		resInst, resIWI = ParsePhi(inst, rout, nil)
	case *s_ssa.InstructionRange:
		resInst, resIWI = ParseRange(inst, rout, nil)
	case *s_ssa.InstructionReturn:
		resInst, resIWI = ParseReturn(inst, rout, nil)
	case *s_ssa.InstructionRunDefers:
		resInst, resIWI = ParseRunDefer(inst, rout, nil)
	case *s_ssa.InstructionSelect:
		return ParseSelectAll(inst, rout, nil)
	case *s_ssa.InstructionSend:
		resInst, resIWI = ParseSend(inst, rout, nil, FORWARD)
	case *s_ssa.InstructionSlice:
		resInst, resIWI = ParseSlice(inst, rout, nil)
	case *s_ssa.InstructionStore:
		resInst, resIWI = ParseStore(inst, rout, nil)
	case *s_ssa.InstructionUnOp:
		resInst, resIWI = ParseUnOp(inst, rout, nil, FORWARD)
	}

	return []s_ssa.Instruction{resInst}, resIWI
}
