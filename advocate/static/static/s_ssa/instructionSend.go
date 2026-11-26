// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSend.go
// Brief: Send Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"

	"golang.org/x/tools/go/ssa"
)

type InstructionSend struct {
	InstructionBase
}

func newSend(f *Function, inst *ssa.Send, i int) *InstructionSend {
	return &InstructionSend{InstructionBase: newInstructionBase(f, Ic_send, inst, i)}
}

func (this *InstructionSend) Instruction() *ssa.Send {
	return this.inst.(*ssa.Send)
}

func (this *InstructionSend) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

func (this *InstructionSend) CanUnblockSimple(elem trace.Element) bool {
	switch elem := elem.(type) {
	case *trace.ElementChannel:
		if elem.Type(true) == trace.ChannelRecv {
			return true
		}
	case *trace.ElementSelect:
		if _, recv, _ := elem.ContainsDirection(); recv {
			return true
		}
	}

	return false
}
