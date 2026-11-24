// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionUnOp.go
// Brief: Unary instruction Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"
	"go/token"

	"golang.org/x/tools/go/ssa"
)

type InstructionUnOp struct {
	InstructionBase
}

func newUnOp(f *Function, inst *ssa.UnOp, i int) *InstructionUnOp {
	return &InstructionUnOp{InstructionBase: newInstructionBase(f, Ic_unOp, inst, i)}
}

func (this *InstructionUnOp) Instruction() *ssa.UnOp {
	return this.inst.(*ssa.UnOp)
}

func (this *InstructionUnOp) setRelevant(_ *Data) {
	this.relevant = this.Conc().Resource()
	this.inTrace = this.isRecv()
}

func (this *InstructionUnOp) isRecv() bool {
	return this.Inst().(*ssa.UnOp).Op == token.ARROW
}

// TODO: check that send and receive have the same type
// TODO: check that select has correct direction
func (this *InstructionUnOp) CanUnblockSimple(elem trace.Element) bool {
	if !this.isRecv() {
		return false
	}

	switch elem := elem.(type) {
	case *trace.ElementChannel:
		if elem.Type(true) == trace.ChannelSend {
			return true
		}
	case *trace.ElementSelect:
		return true
	}

	return false
}
