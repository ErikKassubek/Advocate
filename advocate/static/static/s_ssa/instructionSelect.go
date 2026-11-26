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
	"go/types"
	"strings"

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

func (this *InstructionSelect) CanUnblockSimple(elem trace.Element) bool {
	switch elem := elem.(type) {
	case *trace.ElementChannel:
		send, recv, _ := this.containsDirection()
		t := elem.Type(true)

		switch t {
		case trace.ChannelSend, trace.ChannelClose:
			return recv
		case trace.ChannelRecv:
			return send
		}

	case *trace.ElementSelect:
		e_send, e_recv, _ := elem.ContainsDirection()
		this_send, this_recv, _ := this.containsDirection()

		return (e_send && this_recv) || (e_recv && this_send)
	}

	return false
}

// Contains send, recv, default
func (this *InstructionSelect) containsDirection() (bool, bool, bool) {

	containsSend := false
	containsRecv := false

	for _, state := range this.Instruction().States {
		switch state.Dir {
		case types.SendOnly:
			containsSend = true
		case types.RecvOnly:
			containsRecv = true
		}

		if containsSend && containsRecv {
			break
		}
	}

	containsDefault := strings.Contains(this.term, "nonblocking")

	return containsSend, containsRecv, containsDefault
}
