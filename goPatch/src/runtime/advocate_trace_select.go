// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace_select.go
// Brief: Functionality for selects
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store a spawn
//
// Fields
//   - tReq int64: time when the operation started
//   - tCom int64: time when the operation finished
//   - id uint64: id of the select
//   - cases []AdvocateTraceChannel: the operation for each of the non default cases
//     The elements are sorted the same as the internal sorting in the select,
//     first all send, then all recv
//   - selIndex int: The index of the operation in cases that was executed,
//     if default was executed, this is set to -1
//   - hasDef bool: true if the select has a default case, false otherwise
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
//   - pan bool: caused a panic
type AdvocateTraceSelect struct {
	tReq     int64
	tCom     int64
	id       uint64
	cases    []AdvocateTraceChannel
	selIndex int
	hasDef   bool
	file     string
	line     int
	pan      bool
}

// AdvocateSelectReq adds a select to the trace
//
// Parameter:
//   - cases: cases of the select
//   - nsends: number of send cases
//   - ncases: total number of non default cases
//   - block: true if the select is blocking (has no default), false otherwise
//
// Returns:
//   - index of the operation in the trace
func AdvocateSelectReq(cases *[]scase, nsends int, ncases int, block bool) int {
	if AdvocateTracingDisabled || cases == nil {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipSelect)
	if AdvocateIgnore(file) {
		return -1
	}

	id := GetAdvocateObjectID()
	caseElements := make([]AdvocateTraceChannel, ncases)

	for casi := 0; casi < ncases; casi++ {
		cas := (*cases)[casi]
		c := cas.c

		chanOp := OperationChannelRecv
		if casi < nsends {
			chanOp = OperationChannelSend
		}

		if c == nil { // ignore nil cases
			caseElements[casi] = AdvocateTraceChannel{
				tReq:  timer,
				op:    chanOp,
				isNil: true,
			}
		} else {
			res := AdvocateTraceResource{id: c.id, addr: unsafe.Pointer(c)}
			caseElements[casi] = AdvocateTraceChannel{
				tReq:  timer,
				res:   res,
				op:    chanOp,
				qSize: c.dataqsiz,
			}
		}
	}

	elem := AdvocateTraceSelect{
		tReq:  timer,
		id:    id,
		cases: caseElements,
		file:  file,
		line:  line,
	}

	if !block {
		elem.hasDef = true
	}

	return insertIntoTrace(elem)
}

// AdvocateSelectCom adds a post event for select in case of an non-default case
//
// Parameter:
//   - index: index of the operation in the trace
//   - c: channel of the chosen case
//   - selIndex: index of the chosen case in the select
//   - rClosed: true if the channel was closed at another routine
func AdvocateSelectCom(index int, c *hchan, selIndex int, rClosed bool) {
	if AdvocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceSelect)
	elem.tCom = timer
	elem.selIndex = selIndex

	if selIndex != -1 { // not default case
		// set tpost and cl of chosen case
		chosenCase := elem.cases[selIndex]
		chosenCase.tCom = timer
		if rClosed {
			chosenCase.cl = true
		}

		// set oId
		if chosenCase.op == OperationChannelSend {
			chosenCase.oId = c.numberSend
			c.numberSend++
		} else {
			chosenCase.oId = c.numberRecv
			c.numberRecv++
		}
		chosenCase.qCount = uint(c.numberSend - c.numberRecv)

		elem.cases[selIndex] = chosenCase
	}

	currentGoRoutineInfo().updateElement(index, elem)
}

// AdvocateSelectReqOneNonDef adds a new select element to the trace if the
// select has exactly one non-default case and a default case
//
// Parameter:
//   - c: channel of the non-default case
//   - send: true if the non-default case is a send, false otherwise
//
// Returns:
//   - index of the operation in the trace
func AdvocateSelectReqOneNonDef(c *hchan, send bool) int {
	if AdvocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	id := GetAdvocateObjectID()

	opChan := OperationChannelRecv
	if send {
		opChan = OperationChannelSend
	}

	var caseElem AdvocateTraceChannel

	if c != nil {
		res := AdvocateTraceResource{id: c.id, addr: unsafe.Pointer(c)}
		caseElem = AdvocateTraceChannel{
			tReq:  timer,
			res:   res,
			op:    opChan,
			qSize: c.dataqsiz,
		}
	} else {
		caseElem = AdvocateTraceChannel{
			tReq: timer,
			op:   opChan,
		}
	}

	_, file, line, _ := Caller(CallerSkipSelectOneDef)
	if AdvocateIgnore(file) {
		return -1
	}

	cases := make([]AdvocateTraceChannel, 1)
	cases[0] = caseElem

	elem := AdvocateTraceSelect{
		tReq:   timer,
		id:     id,
		cases:  cases,
		hasDef: true,
		file:   file,
		line:   line,
	}

	return insertIntoTrace(elem)
}

// AdvocateSelectComOneNonDef adds the selected case for a select with one
// non-default and one default case
//
// Parameter:
//   - index: index of the operation in the trace
//   - res: true for channel, false for default
//   - c *hchan: the channel in the select cases
func AdvocateSelectComOneNonDef(index int, res bool, c *hchan) {
	if AdvocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceSelect)

	elem.tCom = timer

	if res { // channel case
		ca := elem.cases[0]
		ca.tCom = timer
		if ca.op == OperationChannelSend {
			c.numberSend++
		} else {
			c.numberRecv++
		}
		ca.qCount = uint(c.numberSend - c.numberRecv)
		elem.cases[0] = ca
		elem.selIndex = 0
	} else { // default case
		elem.selIndex = -1
	}

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     [S],[tPre],[tPost],[id],[cases],[selIndex],[file],[line]
//     where cases consists of the form [case]~[case]~..., followed by a d
//     if the select has a default that was not executed, or D if it was executed.
//     The [case] is build using AdvocateTraceChannel.toStringForSelect()
func (self AdvocateTraceSelect) toString() string {
	p1 := buildTraceElemString("S", self.tReq, self.tCom, self.id)
	p2 := buildTraceElemString(self.selIndex, posToString(self.file, self.line))
	cases := ""
	for i, c := range self.cases {
		if i != 0 {
			cases += "~"
		}
		cases += c.toStringForSelect()
	}
	if self.hasDef {
		if cases != "" {
			cases += "~"
		}
		if self.selIndex == -1 {
			cases += "D"
		} else {
			cases += "d"
		}
	}

	panStr := "-"
	if self.pan {
		panStr = "P"
	}

	return buildTraceElemString(p1, cases, p2, panStr)
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceSelect) getOperation() Operation {
	if self.selIndex == -1 {
		return OperationSelectDefault
	}
	return OperationSelectCase
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceSelect) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceSelect) resource() []AdvocateTraceResource {
	res := make(map[AdvocateTraceResource]struct{})

	for _, c := range self.cases {
		for _, r := range c.resource() {
			res[r] = struct{}{}
		}
	}

	return mapToSlice(res)
}

func (self AdvocateTraceSelect) causedPanic() traceElem {
	self.pan = true
	return self
}
