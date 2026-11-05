// Copyright (c) 2026 Erik Kassubek
//
// File: instruction.go
// Brief: One instructions and instruction paths
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ssa"
)

type path []*instructionWithInfo

func (p path) last() *instructionWithInfo {
	return p[len(p)-1]
}

type instructionWithInfo struct {
	Inst      s_ssa.Instruction
	Variable  string
	Resource  map[int]map[trace.Resource]bool               // index (for field, return/extrace, ...) ->  resource
	Reference map[int]map[*instructionWithInfo]map[int]bool // index (for field, return/extract, ...) -> instruction -> index in instruction
}

func newIWI(inst s_ssa.Instruction, res map[int]map[trace.Resource]bool, par map[int]map[*instructionWithInfo]map[int]bool) *instructionWithInfo {
	return &instructionWithInfo{inst, inst.Variable(), res, par}
}

func newIwiFromIwi(inst s_ssa.Instruction, iwi *instructionWithInfo) *instructionWithInfo {
	return &instructionWithInfo{inst, inst.Variable(), iwi.Resource, iwi.Reference}
}

func newIwiFromIwiIndex(inst s_ssa.Instruction, iwi *instructionWithInfo, index int) *instructionWithInfo {
	newRes := make(map[int]map[trace.Resource]bool)
	newRes[0] = iwi.Resource[index]

	newRef := make(map[int]map[*instructionWithInfo]map[int]bool)
	newRef[0] = iwi.Reference[index]

	return &instructionWithInfo{inst, inst.Variable(), newRes, newRef}
}

func newIWI1(inst s_ssa.Instruction, res map[int]map[trace.Resource]bool) *instructionWithInfo {
	return &instructionWithInfo{inst, inst.Variable(), res, make(map[int]map[*instructionWithInfo]map[int]bool)}
}

func newIWI2(inst s_ssa.Instruction) *instructionWithInfo {
	return &instructionWithInfo{inst, inst.Variable(), make(map[int]map[trace.Resource]bool), make(map[int]map[*instructionWithInfo]map[int]bool)}
}

func newIWI3(inst s_ssa.Instruction, v string, res map[int]map[trace.Resource]bool) *instructionWithInfo {
	return &instructionWithInfo{inst, v, res, make(map[int]map[*instructionWithInfo]map[int]bool)}
}

func newIWI4(inst s_ssa.Instruction, n int) *instructionWithInfo {
	log.Debug("IWI4: ", inst, n)

	resources := make(map[int]map[trace.Resource]bool)
	references := make(map[int]map[*instructionWithInfo]map[int]bool)

	for i := 0; i < n; i++ {
		resources[i] = map[trace.Resource]bool{}
		references[i] = make(map[*instructionWithInfo]map[int]bool)
	}

	return &instructionWithInfo{inst, inst.Variable(), resources, references}
}

func (self *instructionWithInfo) Merge(other *instructionWithInfo) *instructionWithInfo {
	maxLen := max(len(self.Resource), len(other.Resource))
	new_resources := make(map[int]map[trace.Resource]bool, maxLen)

	for i, resource := range self.Resource {
		if _, ok := new_resources[i]; !ok {
			new_resources[i] = make(map[trace.Resource]bool)
		}

		for res := range resource {
			new_resources[i][res] = true
		}
	}

	for i, resource := range other.Resource {
		if _, ok := new_resources[i]; !ok {
			new_resources[i] = make(map[trace.Resource]bool)
		}

		for res := range resource {
			new_resources[i][res] = true
		}
	}

	new_reference := make(map[int]map[*instructionWithInfo]map[int]bool)

	for i, ref := range self.Reference {
		new_reference[i] = ref
	}

	for i, ref := range other.Reference {
		if _, ok := new_reference[i]; !ok {
			new_reference[i] = make(map[*instructionWithInfo]map[int]bool)
		}

		for r, val := range ref {
			new_reference[i][r] = val
		}
	}

	return &instructionWithInfo{self.Inst, self.Variable, new_resources, new_reference}
}

func (self *instructionWithInfo) GetResources() map[trace.Resource]bool {
	visited := make(map[*instructionWithInfo]bool)
	return self.GetResourcesRec(visited)
}

func (self *instructionWithInfo) GetResourcesRec(visited map[*instructionWithInfo]bool) map[trace.Resource]bool {
	result := make(map[trace.Resource]bool)

	visited[self] = true

	for _, resMap := range self.Resource {
		for res := range resMap {
			result[res] = true
		}
	}

	for _, parents := range self.Reference {
		for parent := range parents {
			if _, ok := visited[parent]; ok {
				continue
			}
			for res := range parent.GetResourcesRec(visited) {
				result[res] = true
			}
		}
	}

	return result
}

func (self *instructionWithInfo) GetResourcesIndex(index int) map[trace.Resource]bool {
	result := self.Resource[index]

	for ind := range self.Reference[index] {
		for res := range ind.GetResources() {
			result[res] = true
		}
	}

	return result
}

func (self *instructionWithInfo) GetResourcesMap() map[int]map[trace.Resource]bool {
	if self == nil {
		return map[int]map[trace.Resource]bool{}
	}
	res := make(map[int]map[trace.Resource]bool)

	for i := range self.Resource {
		res[i] = self.GetResourcesIndex(i)
	}

	return res
}

func compatible(iwi *instructionWithInfo, elem trace.Element) (bool, *trace.Resource) {
	if iwi == nil || iwi.Resource == nil {
		return false, nil
	}

	for _, res := range iwi.Resource { // should be only one element, but better to be sure
		if _, ok := res[elem.Resource()]; !ok { // not the same object
			return false, nil
		}

		r := elem.Resource()

		switch elem := elem.(type) {
		case *trace.ElementChannel:
			switch elem.Type(true) {
			case trace.ChannelRecv:
				return iwi.Inst.Class() == s_ssa.Ic_send, &r
			case trace.ChannelSend:
				if i, ok := iwi.Inst.Inst().(*ssa.UnOp); ok {
					return i.Op == token.ARROW, &r
				}
			}
		case *trace.ElementSelect:
			for _, c := range elem.GetCases() {
				if _, ok := res[c.Resource()]; !ok { // not the same object
					continue
				}
				switch c.Type(true) {
				case trace.ChannelRecv:
					return iwi.Inst.Class() == s_ssa.Ic_send, &r
				case trace.ChannelSend:
					if i, ok := iwi.Inst.Inst().(*ssa.UnOp); ok {
						return i.Op == token.ARROW, &r
					}
				}
			}
		case *trace.ElementMutex:
			if iwi.Inst.HasMutex() {
				if elem.Type(true) == trace.MutexLock && (strings.Contains(iwi.Inst.Term(), "(*sync.Mutex).Unlock(") || strings.Contains(iwi.Inst.Term(), "(*sync.RWMutex).Unlock(")) {
					return true, &r
				} else if elem.Type(true) == trace.MutexRLock && strings.Contains(iwi.Inst.Term(), "(*sync.RWMutex).RUnlock(") {
					return true, &r
				}
			}
		case *trace.ElementCond:
			return iwi.Inst.HasCond() && (strings.Contains(iwi.Inst.Term(), "(*sync.Cond).Signal(") || strings.Contains(iwi.Inst.Term(), "(*sync.Cond).Broadcast(")), &r
		case *trace.ElementWait:
			return iwi.Inst.HasWG() && strings.Contains(iwi.Inst.Term(), "(*sync.WaitGroup).Done("), &r
		}

	}

	return false, nil
}

func (self *instructionWithInfo) String() string {
	if self == nil {
		return ">IWI NIL <"
	} else if self.Inst == nil {
		return "> IWI.INST NIL <"
	}
	res := "> " + self.Inst.String() + " | "

	resSlice := self.GetResourcesMap()

	if len(resSlice) == 0 {
		res += "NO RES"
	}

	for index, resources := range resSlice {
		res += fmt.Sprintf(" %d # ", index)
		for resource := range resources {
			res += fmt.Sprint(resource.Id(), " ")
		}

		if len(resources) == 0 {
			res += "-"
		}

		res += " # "
		references := self.Reference[index]
		for ref := range references {
			res += fmt.Sprint(ref.Variable, " ")
		}

		if len(references) == 0 {
			res += "-"
		}

		if index != len(self.Resource) {
			res += " | "
		}
	}

	res += " <"

	return res
}

func fmtInstRes(resource map[trace.Resource]bool) map[int]map[trace.Resource]bool {
	if resource == nil {
		return make(map[int]map[trace.Resource]bool)
	}

	res := make(map[int]map[trace.Resource]bool)
	res[0] = resource

	return res
}

func (self *instructionWithInfo) sameResource(inst *instructionWithInfo) bool {
	if len(self.Resource) != len(inst.Resource) {
		return false
	}

	for i := 0; i < len(self.Resource); i++ {
		res1 := self.Resource[i]
		res2 := inst.Resource[i]

		if len(res1) != len(res2) {
			return false
		}

		for k := range res1 {
			if _, ok := res2[k]; !ok {
				return false
			}
		}

	}
	return true

}

func (self *instructionWithInfo) addReferenceIndex(iwi *instructionWithInfo, index1, index2 int) {
	if _, ok := self.Reference[index1]; !ok {
		self.Reference[index1] = make(map[*instructionWithInfo]map[int]bool)
	}
	if _, ok := self.Reference[index1][iwi]; !ok {
		self.Reference[index1][iwi] = make(map[int]bool)
	}
	if _, ok := self.Reference[index1][self]; !ok {
		self.Reference[index1][self] = make(map[int]bool)
	}
	if _, ok := self.Reference[index2]; !ok {
		self.Reference[index2] = make(map[*instructionWithInfo]map[int]bool)
	}
	if _, ok := self.Reference[index2][iwi]; !ok {
		self.Reference[index2][iwi] = make(map[int]bool)
	}
	if _, ok := self.Reference[index2][self]; !ok {
		self.Reference[index2][self] = make(map[int]bool)
	}

	self.Reference[index1][iwi][index2] = true
	for ref := range iwi.Reference[index2] {
		if _, ok := self.Reference[index1][ref]; !ok {
			self.Reference[index1][ref] = make(map[int]bool)
		}
		self.Reference[index1][ref][index2] = true
	}

	iwi.Reference[index2][self][index1] = true
	for ref := range self.Reference[index1] {
		if _, ok := self.Reference[index1][ref]; !ok {
			self.Reference[index1][ref] = make(map[int]bool)
		}
		iwi.Reference[index2][ref][index1] = true
	}
}

func (self *instructionWithInfo) addReferenceAll(iwi *instructionWithInfo) {
	indices := make(map[int]bool)

	for index := range self.Reference {
		indices[index] = true
	}
	for index := range iwi.Reference {
		indices[index] = true
	}

	for index := range indices {
		if _, ok := self.Reference[index]; !ok {
			self.Reference[index] = make(map[*instructionWithInfo]map[int]bool)
		}
		if _, ok := self.Reference[index][iwi]; !ok {
			self.Reference[index][iwi] = make(map[int]bool)
		}
		if _, ok := iwi.Reference[index]; !ok {
			iwi.Reference[index] = make(map[*instructionWithInfo]map[int]bool)
		}
		if _, ok := iwi.Reference[index][self]; !ok {
			iwi.Reference[index][self] = make(map[int]bool)
		}

		self.Reference[index][iwi][index] = true
		for ref := range iwi.Reference[index] {
			if self == ref {
				continue
			}
			if index == 1 {
				log.Debug("MAKE: ", ref.Inst)
			}
			if _, ok := self.Reference[index][ref]; !ok {
				self.Reference[index][ref] = make(map[int]bool)
			}
			self.Reference[index][ref][index] = true
		}

		iwi.Reference[index][self][index] = true
		for ref := range self.Reference[index] {
			if iwi == ref {
				continue
			}
			if index == 1 {
				log.Debug("MAKE: ", ref.Inst)
			}
			if _, ok := iwi.Reference[index][ref]; !ok {
				iwi.Reference[index][ref] = make(map[int]bool)
			}
			iwi.Reference[index][ref][index] = true
		}
	}
}
