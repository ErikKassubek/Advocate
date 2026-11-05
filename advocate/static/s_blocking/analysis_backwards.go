// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: For a given alloc, get all SSA variables that are equal to the allocated resource when following the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/analysis/a_base"
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
	"fmt"
)

func determineResouceToSSAAtTermination() {
	tr := &a_base.MainTrace
	trIter := tr.AsIterator()

	// get the first relevant value in init
	f := data.Ssa().InitFunc()
	b := f.Blocks()[1]

	for _, inst := range b.Instrs() {
		if inst.InTrace() {
			blocking.nextPerRout[1] = inst
			break
		} else {
			printInstr(-1, inst, nil, nil)
		}
	}

	blocking.jumpBackPos[1] = types.NewStack[s_ssa.Instruction]()
	blocking.NewPathPerRoutine(1)

	lastWasFork := make(map[int]bool)

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		// Set main as func
		if elemF, ok := elem.(*trace.ElementFunc); ok && elemF.Name() == "main.main" {
			f := data.Ssa().MainFunc()
			b := f.Blocks()[0]

			// skip non relevant instructions in main
			// for _, inst := range b.Instrs() {
			// 	if inst.InTrace() {
			// 		blocking.nextPerRout[1] = inst
			// 		break
			// 	} else {
			// 		printInstr(1, inst, elem, nil)
			// 	}
			// }
			blocking.nextPerRout[1] = parseNonTraceInstructions(b.Instrs()[0], 1)
			continue
		}

		routine := elem.RoutineID()
		if _, ok := elem.(*trace.ElementFunc); ok && lastWasFork[routine] {
			lastWasFork[routine] = false
			continue
		}

		switch elem.(type) {
		case *trace.ElementRoutineEnd:
			continue
		case *trace.ElementFork:
			lastWasFork[elem.ResourceID()] = true
		default:
			lastWasFork[routine] = false
		}

		// we can ignore routines that are no longer running.
		// We also do not need to determine current pos of routine
		// if elem.Routine().IsTerminated() {
		// 	delete(blocking.nextPerRout, routine)
		// 	continue
		// }

		// // we can ignore routines, where the OAT tells us, that it does not contain any blocking operations.
		// // Here, we only need to find the current instruction
		// if !types.HasCommonElement(elem.Routine().Resources(), blocking.blockedResources) {

		// 	blocking.nextPerRout[routine] = skipNonRelevant(blocking.nextPerRout[routine], routine)
		// 	continue
		// }

		next := parseInstructions(elem, blocking.nextPerRout[routine], routine)
		if next != nil {
			blocking.nextPerRout[routine] = next
		} else {
			delete(blocking.nextPerRout, routine)
		}
	}
}

// Iterate over SSA starting from start and stopping before end
func parseInstructions(elem trace.Element, inst s_ssa.Instruction, rout int) s_ssa.Instruction {
	if inst == nil {
		return nil
	}

	if !elem.Committed() {
		return nil
	}

	next := parseInstruction(inst, rout, elem)

	if next == nil {
		return nil
	}

	return parseNonTraceInstructions(next, rout)
}

func parseInstruction(inst s_ssa.Instruction, rout int, elem trace.Element) s_ssa.Instruction {
	next, info := parse(inst, rout, elem)

	printInstr(rout, inst, elem, info)

	return next
}

func printInstr(rout int, inst s_ssa.Instruction, elem trace.Element, info *instructionWithInfo) {
	infoStr := "<NIL>"
	res := info.GetResourcesMap()
	if info != nil && len(res) != 0 {
		infoStr = ""

		if len(res) != 1 {
			for i, r := range res {
				if i != 0 {
					infoStr += " "
				}
				infoStr += fmt.Sprintf("#%d: ", i)
				if len(r) == 0 {
					infoStr += "<NIL>"
				}
				for res := range r {
					infoStr += fmt.Sprint(res.Id())
				}
			}
		} else {
			for res := range res[0] {
				infoStr += fmt.Sprint(res.Id())
			}
		}
	}

	if elem != nil {
		log.Debugf("%2d | %-80.80s | %-25s | %s", rout, inst.StringInfo(), infoStr, elem.StringLocal())
	} else {
		log.Debugf("%2d | %-80.80s | %-25s |", rout, inst.StringInfo(), infoStr)
	}
}

func parseNonTraceInstructions(inst s_ssa.Instruction, rout int) s_ssa.Instruction {
	for p := inst; p != nil; {
		if p.InTrace() {
			return p
		}

		p = parseInstruction(p, rout, nil)

		blocking.lastBlockIdPerRoutine[rout] = inst.Inst().Block().Index

	}

	return nil
}

func parse(inst s_ssa.Instruction, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	if inst == nil {
		return nil, nil
	}

	switch inst := inst.(type) {
	case *s_ssa.InstructionAlloc:
		return ParseAlloc(inst, rout, elem)
	case *s_ssa.InstructionCall:
		return ParseCall(inst, rout, elem)
	case *s_ssa.InstructionExtract:
		return ParseExtract(inst, rout, elem)
	case *s_ssa.InstructionField:
		return ParseField(inst, rout, elem)
	case *s_ssa.InstructionFieldAddr:
		return ParseFieldAddr(inst, rout, elem)
	case *s_ssa.InstructionGo:
		return ParseGo(inst, rout, elem)
	case *s_ssa.InstructionIf:
		return ParseIf(inst, rout, elem)
	case *s_ssa.InstructionIndex:
		return ParseIndex(inst, rout, elem)
	case *s_ssa.InstructionIndexAddr:
		return ParseIndexAddr(inst, rout, elem)
	case *s_ssa.InstructionJump:
		return ParseJump(inst, rout, elem)
	case *s_ssa.InstructionLookup:
		return ParseLookup(inst, rout, elem)
	case *s_ssa.InstructionMakeChan:
		return ParseMakeChan(inst, rout, elem)
	case *s_ssa.InstructionMakeClosure:
		return ParseMakeClosure(inst, rout, elem)
	case *s_ssa.InstructionMakeInterface:
		return ParseMakeInterface(inst, rout, elem)
	case *s_ssa.InstructionMakeMap:
		return ParseMakeMap(inst, rout, elem)
	case *s_ssa.InstructionMakeSlice:
		return ParseMakeSlice(inst, rout, elem)
	case *s_ssa.InstructionMapUpdate:
		return ParseMapUpdate(inst, rout, elem)
	case *s_ssa.InstructionNext:
		return ParseNext(inst, rout, elem)
	case *s_ssa.InstructionPhi:
		return ParsePhi(inst, rout, elem)
	case *s_ssa.InstructionRange:
		return ParseRange(inst, rout, elem)
	case *s_ssa.InstructionReturn:
		return ParseReturn(inst, rout, elem)
	case *s_ssa.InstructionRunDefers:
		return ParseRunDefer(inst, rout, elem)
	case *s_ssa.InstructionSelect:
		return ParseSelect(inst, rout, elem)
	case *s_ssa.InstructionSend:
		return ParseSend(inst, rout, elem, !FORWARD)
	case *s_ssa.InstructionSlice:
		return ParseSlice(inst, rout, elem)
	case *s_ssa.InstructionStore:
		return ParseStore(inst, rout, elem)
	case *s_ssa.InstructionUnOp:
		return ParseUnOp(inst, rout, elem, !FORWARD)
	default:
		return inst.Next(), nil
	}
}
