package s_ssa

import (
	"advocate/trace"

	"golang.org/x/tools/go/ssa"
)

type Instruction interface {
	Variable() string
	Term() string
	VariableGlobal() bool
	TermGlobal() bool
	String() string
	StringInfo() string
	Inst() ssa.Instruction
	InTrace() bool

	Conc() hasConcInfo
	HasChannel() bool
	HasMutex() bool
	HasCond() bool
	HasWG() bool

	Class() InstClass

	CanUnblockSimple(elem trace.Element) bool

	setVariable(name string, global bool)
	setTerm(term string, global bool)
	setInst(inst ssa.Instruction)
	setConc(conc hasConcInfo)
	setRelevant(data *Data)

	Function() *Function
	Block() *Block
	Next() Instruction
	FirstInBlock(b_id int) Instruction
}
