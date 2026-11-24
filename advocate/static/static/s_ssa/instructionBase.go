// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionBase.go
// Brief: Base struct for ssa instructions
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/static/static/s_base"
	"advocate/trace"
	"fmt"
	"strings"

	"golang.org/x/tools/go/ssa"
)

type InstructionBase struct {
	variable string
	term     string
	inst     ssa.Instruction

	variableGlobal bool
	termGlobal     bool

	varPtr bool

	relevant bool
	inTrace  bool

	class InstClass

	conc hasConcInfo

	f    *Function
	b    *ssa.BasicBlock
	i_id int
}

func newInstructionBase(f *Function, c InstClass, inst ssa.Instruction, index int) InstructionBase {
	name := ""
	switch v := inst.(type) {
	case ssa.Value:
		name = v.Name()
	case nil:
		// Be robust against bad transforms.
		name = "<deleted>"
	}

	globalName := false
	globalTerm := false

	if i, ok := inst.(*ssa.Store); ok {
		switch i.Addr.(type) {
		case *ssa.Global:
			globalName = true
		}

		switch i.Val.(type) {
		case *ssa.Global:
			globalTerm = true
		}
	}

	if i, ok := inst.(*ssa.UnOp); ok {
		switch i.X.(type) {
		case *ssa.Global:
			globalTerm = true
		}
	}

	// global var assign
	var b InstructionBase
	if inst != nil {
		term := inst.String()
		if name == "" && strings.Contains(term, " = ") {
			fields := strings.Split(term, " = ")
			name = fields[0]
			term = fields[1]
		}

		switch inst := inst.(type) {
		case *ssa.Send:
			name = inst.Chan.Name()
		}

		b = InstructionBase{class: c, f: f, b: inst.Block(), i_id: index}

		b.setVariable(name, globalName)
		b.setTerm(term, globalTerm)
		b.setInst(inst)
	} else {
		b = InstructionBase{class: c, f: f, b: nil, i_id: 0}
	}

	return b
}

func (this *InstructionBase) String() (res string) {
	if this.variable == "" {
		res = this.term
	} else {
		if this.varPtr {
			res = fmt.Sprintf("*%s = %s", this.variable, this.term)
		} else {
			res = fmt.Sprintf("%s = %s", this.variable, this.term)
		}
	}

	return
}

func (this *InstructionBase) StringInfo() (res string) {
	if this.relevant {
		res += "+"
	} else {
		res += "-"
	}

	if this.inTrace {
		res += "+ "
	} else {
		res += "- "
	}

	res += fmt.Sprintf("%-60s", this.String())

	// name
	obj := func(i int) s_base.ObjName {
		switch concRes(i) {
		case chanInd:
			return s_base.Channel
		case mutexInd:
			return s_base.Mutex
		case condVarInd:
			return s_base.CondVar
		case wgInd:
			return s_base.Wg
		default:
			return s_base.UnknownObj
		}
	}

	if this.class != Ic_unknown {
		res += " | " + fmt.Sprintf("%-20s", string(this.class)) + " | "
	}

	found := false
	for i := 0; i < 4; i++ {
		if this.conc[i] {
			if found {
				res += ", "
			}
			res += string(obj(i))
			found = true
		}
	}

	if !found {
		res += "-"
	}

	return
}

func (this *InstructionBase) Variable() string {
	return this.variable
}

func (this *InstructionBase) Term() string {
	return this.term
}

func (this *InstructionBase) VariableGlobal() bool {
	return this.variableGlobal
}

func (this *InstructionBase) TermGlobal() bool {
	return this.termGlobal
}

func (this *InstructionBase) Relevant() bool {
	return this.relevant
}

func (this *InstructionBase) InTrace() bool {
	return this.inTrace
}

func (this *InstructionBase) Class() InstClass {
	return this.class
}

func (this *InstructionBase) Inst() ssa.Instruction {
	return this.inst
}

func (this *InstructionBase) Conc() hasConcInfo {
	return this.conc
}

func (this *InstructionBase) HasChannel() bool {
	return this.conc[chanInd]
}

func (this *InstructionBase) HasMutex() bool {
	return this.conc[mutexInd]
}

func (this *InstructionBase) HasCond() bool {
	return this.conc[condVarInd]
}

func (this *InstructionBase) HasWG() bool {
	return this.conc[wgInd]
}

func (this *InstructionBase) setVariable(name string, global bool) {
	variable := strings.TrimPrefix(name, "*")
	this.variable = variable
	this.varPtr = (name != variable)
	this.variableGlobal = global
}

func (this *InstructionBase) setTerm(term string, global bool) {
	this.term = term
	this.termGlobal = global
}

func (this *InstructionBase) setInst(inst ssa.Instruction) {
	this.inst = inst
}

func (this *InstructionBase) setConc(conc hasConcInfo) {
	this.conc = conc
}

func (this *InstructionBase) Function() *Function {
	return this.f
}

func (this *InstructionBase) Block() *Block {
	return this.f.blocks[this.b.Index]
}

func (this *InstructionBase) FirstInBlock(b_id int) Instruction {
	return this.Function().Blocks()[b_id].Instrs()[0]
}

func (this *InstructionBase) Next() Instruction {
	instID := this.i_id + 1

	res := this.Block().Instrs()[instID]

	return res
}

func (this *InstructionBase) CanUnblockSimple(elem trace.Element) bool {
	return false
}

func NewSsaPosFunc(f *Function) Instruction {
	b := f.Blocks()[0]
	return b.Instrs()[0]
}

func NewSsaPosFuncBlock(f *Function, blockID int) Instruction {
	b := f.Blocks()[blockID]
	return b.Instrs()[0]
}
