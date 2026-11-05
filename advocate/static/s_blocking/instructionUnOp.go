// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionUnOp.go
// Brief: Unary instruction Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"go/token"
)

var recvForward = make(map[*s_ssa.InstructionUnOp]map[int]map[trace.Resource]bool) // inst -> res -> routine

func instInfoPointerDereference(inst *s_ssa.InstructionUnOp, rout int) *instructionWithInfo {
	term := inst.Term()
	iwi := getDecOfSSAVar(rout, term)

	iwi_new := newIwiFromIwi(inst, iwi)

	return addPathInstr(rout, iwi_new)
}

func instInfoReceive(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element, forward bool) *instructionWithInfo {
	var receivedValue *instructionWithInfo
	if elem != nil {
		receivedValue = blocking.chanBuffer[elem.ResourceID()].Pop()
	}

	if forward {
		v := inst.Instruction().X.Name()
		// TODO: implement
		iwiReceiver := getDecOfSSAVar(rout, v)

		receivedValue = &instructionWithInfo{Inst: inst, Variable: inst.Instruction().Name(), Reference: make(map[int]map[*instructionWithInfo]map[int]bool)}
		receivedValue.Reference[0] = make(map[*instructionWithInfo]map[int]bool)

		for res := range iwiReceiver.Resource[0] {
			for _, iwiSend := range sendForward[res.Id()] {
				receivedValue.addReferenceAll(iwiSend)
			}
		}
	}

	if receivedValue == nil {
		return addPathInstr(rout, newIWI2(inst))
	}

	return addPathInstr(rout, newIwiFromIwi(inst, receivedValue))
}

func instInfoUnOp(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element, forward bool) *instructionWithInfo {
	switch inst.Instruction().Op {
	case token.MUL: // pointer dereference
		return instInfoPointerDereference(inst, rout)
	case token.ARROW: // channel receive
		return instInfoReceive(inst, rout, elem, forward)
	}

	return nil
}

func ParseUnOp(inst *s_ssa.InstructionUnOp, rout int, elem trace.Element, forward bool) (s_ssa.Instruction, *instructionWithInfo) {
	iwi := instInfoUnOp(inst, rout, elem, forward)

	return inst.Next(), iwi
}
