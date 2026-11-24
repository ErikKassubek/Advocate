// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSelect.go
// Brief: Select Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"

	"golang.org/x/tools/go/ssa"
)

type InstructionSelect struct {
	InstructionBase
}

func newSelect(f *Function, inst *ssa.Select, i int) *InstructionSelect {
	return &InstructionSelect{InstructionBase: newInstructionBase(f, Ic_select, inst, i)}
}

func (this *InstructionSelect) Instruction() *ssa.Select {
	return this.inst.(*ssa.Select)
}

func (this *InstructionSelect) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

// TODO: check that select has correct direction
// TODO: check that send and receive have the same type
func (this *InstructionSelect) CanUnblockSimple(elem trace.Element) bool {
	switch elem.(type) {
	case *trace.ElementChannel, *trace.ElementSelect:
		return true
	}

	return false
}
