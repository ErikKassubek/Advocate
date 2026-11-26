// Copyright (c) 2026 Erik Kassubek
//
// File: ssaData.go
// Brief: Blocking Data
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static"
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
	"strings"
)

var data *static.Data = nil
var blocking = newBlockData()

type BlockingData struct {
	nextPerRout           map[int]s_ssa.Instruction
	jumpBackPos           map[int]*types.Stack[s_ssa.Instruction]
	closures              map[string]*s_ssa.Function
	lastClosure           map[int][]*instructionWithInfo
	globalVars            map[string]*instructionWithInfo
	pathPerRoutine        map[int]*types.Stack[path]
	lastBlockIdPerRoutine map[int]int
	chanBuffer            map[int]*types.Stack[*instructionWithInfo]
	returnVariables       map[int]*types.Stack[*s_ssa.InstructionCall]

	funcCallToSSAFunc map[*trace.ElementFunc]*s_ssa.Function
	blocked           map[trace.Element][]trace.Resource
	blockedResources  map[int]trace.Resource

	maxRoutId int
}

func newBlockData() *BlockingData {
	return &BlockingData{
		nextPerRout:           make(map[int]s_ssa.Instruction),
		jumpBackPos:           make(map[int]*types.Stack[s_ssa.Instruction]),
		closures:              make(map[string]*s_ssa.Function),
		lastClosure:           make(map[int][]*instructionWithInfo),
		globalVars:            make(map[string]*instructionWithInfo),
		pathPerRoutine:        make(map[int]*types.Stack[path]),
		funcCallToSSAFunc:     make(map[*trace.ElementFunc]*s_ssa.Function),
		blocked:               make(map[trace.Element][]trace.Resource),
		blockedResources:      make(map[int]trace.Resource),
		lastBlockIdPerRoutine: make(map[int]int),
		chanBuffer:            make(map[int]*types.Stack[*instructionWithInfo]),
		returnVariables:       make(map[int]*types.Stack[*s_ssa.InstructionCall]),
		maxRoutId:             0,
	}
}

func (self *BlockingData) NewFuncStack(rout int, inst *s_ssa.InstructionCall) {
	res := make([]*instructionWithInfo, 0)
	if inst != nil && inst.Variable() != "" {
		res = append(res, &instructionWithInfo{Variable: inst.Variable()})
	}

	self.pathPerRoutine[rout].Push(res)
	self.returnVariables[rout].Push(inst)
}

func (self *BlockingData) ReturnStack(rout int) *s_ssa.InstructionCall {
	self.pathPerRoutine[rout].Pop()
	return self.returnVariables[rout].Pop()
}

func (self *BlockingData) NewPathPerRoutine(rout int) {
	self.pathPerRoutine[rout] = types.NewStack[path]()
	self.returnVariables[rout] = &types.Stack[*s_ssa.InstructionCall]{}
}

func addPathInstr(rout int, iwi *instructionWithInfo) *instructionWithInfo {
	if _, ok := blocking.pathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	top := blocking.pathPerRoutine[rout].Pop()
	top = append(top, iwi)
	blocking.pathPerRoutine[rout].Push(top)

	return iwi
}

func addPathParam(rout int, v string, iwi *instructionWithInfo, f *s_ssa.Function) *instructionWithInfo {
	if _, ok := blocking.pathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	var res map[int]map[trace.Resource]bool
	var par map[int]map[*instructionWithInfo]map[int]bool

	if iwi != nil {
		res = iwi.Resource
		par = iwi.Reference
	}

	newElem := &instructionWithInfo{s_ssa.NewParameter(f), v, res, par}

	top := blocking.pathPerRoutine[rout].Pop()
	top = append(top, newElem)
	blocking.pathPerRoutine[rout].Push(top)

	return newElem
}

func getDecOfSSAVar(rout int, v string) *instructionWithInfo {
	if strings.Contains(v, ":") {
		return &instructionWithInfo{Resource: make(map[int]map[trace.Resource]bool), Reference: make(map[int]map[*instructionWithInfo]map[int]bool)}
	}

	ppr := blocking.pathPerRoutine[rout].Peek()

	v = strings.TrimPrefix(v, "*")
	v = strings.TrimPrefix(v, "<-")

	for i := len(ppr) - 1; i >= 0; i-- {
		if ppr[i].Variable == v {
			return ppr[i]
		}
	}

	// global variables
	if b, ok := blocking.globalVars[v]; ok {
		return b
	}

	log.Errorf("Unable to find declaration of %s in rout %d", v, rout)

	return nil
}
