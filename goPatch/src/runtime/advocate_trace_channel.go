// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_trace_channel.go
// Brief: Functionality for the channel
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

import "unsafe"

var unbufferedChannelComSend = make(map[uint64]int64) // id -> tpost
var unbufferedChannelComRecv = make(map[uint64]int64) // id -> tpost
var unbufferedChannelComSendMutex mutex
var unbufferedChannelComRecvMutex mutex

// Struct to store an operation on a channel variable
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - res AdvocateTraceResource: the resource the op is applied to
//   - op Operation: operation type
//   - cl bool: true if the recv was executed because the channel is closed
//   - oId uint64: operation id, communicating send and recv have the same oID
//   - qSize uint: size of the channel buffer
//   - qCount uint: number of element in the buffer after the operation finished
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
//   - isNil bool: true if the channel is nil
//   - pan bool: true if it caused a panic
type AdvocateTraceChannel struct {
	tReq   int64
	tCom   int64
	res    AdvocateTraceResource
	op     Operation
	cl     bool
	oId    uint64
	qSize  uint
	qCount uint
	file   string
	line   int
	isNil  bool
	pan    bool
}

// AdvocateChanReq adds a channel send/receive to the trace.
//
// Parameters:
//   - mem unsafe.Pointer: memory address
//   - op Operation: operation send/recv
//   - isNil bool: true if the channel is nil
//
// Returns:
//   - int: index of the operation in the trace, return -1 if it is a atomic operation
func AdvocateChanReq(c *hchan, op Operation, isNil bool) int {
	if AdvocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipChanSendRecv)

	if AdvocateIgnore(file) {
		return -1
	}

	res := AdvocateTraceResource{id: c.id, addr: unsafe.Pointer(c)}

	elem := AdvocateTraceChannel{
		tReq:  timer,
		res:   res,
		op:    op,
		oId:   0,
		qSize: c.dataqsiz,
		file:  file,
		line:  line,
		isNil: isNil,
	}

	return insertIntoTrace(elem)
}

// AdvocateChanClose adds a channel close to the trace
//
// Parameter:
//   - c *hchan
//
// Returns:
//   - index of the operation in the trace
func AdvocateChanClose(c *hchan) int {
	if AdvocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipChanClose)
	if AdvocateIgnore(file) {
		return -1
	}

	res := AdvocateTraceResource{id: c.id, addr: unsafe.Pointer(c)}

	elem := AdvocateTraceChannel{
		tReq:   timer,
		tCom:   timer,
		res:    res,
		op:     OperationChannelClose,
		qSize:  c.dataqsiz,
		qCount: c.qcount,
		file:   file,
		line:   line,
	}

	return insertIntoTrace(elem)
}

// AdvocateChanCom sets the operation as successfully finished
//
// Parameters:
//   - index: index of the operation in the trace
//   - c: the channel
//   - op: the operation
func AdvocateChanCom(index int, c *hchan, op Operation) {
	if AdvocateTracingDisabled {
		return
	}

	time := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceChannel)

	set := false

	if elem.qSize == 0 { // unbuffered channel
		if elem.op == OperationChannelSend {
			lock(&unbufferedChannelComRecvMutex)
			if tpost, ok := unbufferedChannelComRecv[elem.res.id]; ok {
				elem.tCom = tpost - 1
				delete(unbufferedChannelComRecv, elem.res.id)
			} else {
				elem.tCom = time
				lock(&unbufferedChannelComSendMutex)
				unbufferedChannelComSend[elem.res.id] = time
				unlock(&unbufferedChannelComSendMutex)
			}
			unlock(&unbufferedChannelComRecvMutex)
			set = true
		} else if elem.op == OperationChannelRecv {
			lock(&unbufferedChannelComSendMutex)
			if tpost, ok := unbufferedChannelComSend[elem.res.id]; ok {
				elem.tCom = tpost + 1
				delete(unbufferedChannelComSend, elem.res.id)
			} else {
				elem.tCom = time
				unbufferedChannelComRecv[elem.res.id] = time
			}
			unlock(&unbufferedChannelComSendMutex)
			set = true
		}
	}

	if !set {
		elem.tCom = time
	}
	elem.qCount = c.qcount

	if c != nil {
		if op == OperationChannelSend {
			c.numberSend++
			elem.oId = c.numberSend
		} else if op == OperationChannelRecv {
			c.numberRecv++
			elem.oId = c.numberRecv
		}
	}

	currentGoRoutineInfo().updateElement(index, elem)
}

// AdvocateChanPostCausedByClose sets the operation as successfully finished
// Args:
//   - index: index of the operation in the trace
func AdvocateChanPostCausedByClose(index int) {
	if AdvocateTracingDisabled {
		return
	}

	time := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceChannel)

	elem.tCom = time
	elem.cl = true

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     C,[tPre],[tPost],[id],[operation],[cl],[oId],[qSize],[qCount],[file],[line]
func (self AdvocateTraceChannel) toString() string {
	opStr := ""
	switch self.op {
	case OperationChannelSend:
		opStr = "S"
	case OperationChannelRecv:
		opStr = "R"
	case OperationChannelClose:
		opStr = "C"
	}

	idStr := "*"
	if !self.isNil {
		idStr = uint64ToString(self.res.id)
	}

	panStr := "-"
	if self.pan {
		panStr = "P"
	}

	return buildTraceElemString("C", self.tReq, self.tCom, idStr, opStr, self.cl, self.oId, self.qSize, self.qCount, posToString(self.file, self.line), panStr)
}

// Get a string representation for the channel if it is used as a select case
//
// Returns:
//   - string: the string representation of the form
//     C,[id].[operation].[cl].[oId].[qSize].[qCount]
func (self AdvocateTraceChannel) toStringForSelect() string {
	opStr := ""
	switch self.op {
	case OperationChannelSend:
		opStr = "S"
	case OperationChannelRecv:
		opStr = "R"
	case OperationChannelClose:
		opStr = "C"
	}

	idStr := "*"
	if !self.isNil {
		idStr = uint64ToString(self.res.id)
	}

	return buildTraceElemStringSep(".", "C", idStr, opStr, self.cl, self.oId, self.qSize, self.qCount)
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceChannel) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceChannel) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceChannel) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{self.res}
}

func (self AdvocateTraceChannel) causedPanic() traceElem {
	self.pan = true
	return self
}
