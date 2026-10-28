// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/channel.go
// Brief: Struct and functions for channel operations in the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"advocate/analysis/hb/a_clock"
	"advocate/utils/flags"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementChannel is a trace element for a channel
//
// Fields:
//   - objId int: The id of the channel
//   - op ObjectType: The operation on the channel
//   - tReq int: The timestamp at the start of the event
//   - tCom int: The timestamp at the end of the event
//   - pos position: code position
//   - ci *concInfo: concurrency info
//   - oID int: The id of the other communication
//   - cl bool: Whether the channel has closed
//   - qSize int: The size of the channel queue
//   - qCount int: The number of elements in the queue after the operation
//   - sel *traceElementSelect: The select operation, if the channel operation
//     is part of a select, otherwise nil
//   - selIndex int: index of the channel in sel.chases if sel != nil, otherwise -1
//   - partner *ElementChannel: The partner of the channel operation
//   - function *ElementFunc: the function the operation is in
//   - pan bool: cause a panic
type ElementChannel struct {
	ElementBase

	objId    int
	op       OperationType
	tReq     int
	tCom     int
	pos      Position
	ci       *concInfo
	oID      int
	cl       bool
	qSize    int
	qCount   int
	sel      *ElementSelect
	selIndex int
	partner  *ElementChannel
	function *ElementFunc
	pan      bool
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementChannel adds a new channel element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tReq string: The timestamp at the start of the event
//   - tCom string: The timestamp at the end of the event
//   - id string: The id of the channel
//   - opC string: The operation on the channel
//   - cl string: Whether the channel was finished because it was closed
//   - oId string: The id of the other communication
//   - qSize string: The size of the channel queue
//   - qCount string: The number of elements in the queue
//   - pos string: The position of the channel operation in the code
//
// Returns:
//   - error
func (this *Trace) AddTraceElementChannel(routine int, tReq string,
	tCom string, id string, opC string, cl string, oID string, qSize string,
	qCount string, pos string, pan string) error {

	tReqInt, err := strconv.Atoi(tReq)
	if err != nil {
		return errors.New("tReq is not an integer")
	}

	tComInt, err := strconv.Atoi(tCom)
	if err != nil {
		return errors.New("tCom is not an integer")
	}

	idInt := -1
	if id != "*" {
		idInt, err = strconv.Atoi(id)
		if err != nil {
			return errors.New("id is not an integer")
		}
	}

	var opCInt OperationType
	switch opC {
	case "S":
		opCInt = ChannelSend
	case "R":
		opCInt = ChannelRecv
	case "C":
		opCInt = ChannelClose
	default:
		return errors.New("opC is not a valid operation")
	}

	clBool, err := strconv.ParseBool(cl)
	if err != nil {
		return errors.New("cl is not a boolean")
	}

	oIDInt, err := strconv.Atoi(oID)
	if err != nil {
		return errors.New("oId is not an integer")
	}

	qSizeInt, err := strconv.Atoi(qSize)
	if err != nil {
		return fmt.Errorf("qSize '%s' is not an integer", qSize)
	}

	qCountInt, err := strconv.Atoi(qCount)
	if err != nil {
		return fmt.Errorf("qCount '%s' is not an integer", qCount)
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	panBool := false
	if pan == "P" {
		panBool = true
	}

	elem := ElementChannel{
		ElementBase: this.newElementBase(this, routine),
		tReq:        tReqInt,
		tCom:        tComInt,
		objId:       idInt,
		op:          opCInt,
		cl:          clBool,
		oID:         oIDInt,
		qSize:       qSizeInt,
		qCount:      qCountInt,
		pos:         NewPosition(file, line),
		selIndex:    -1,
		ci:          newConcInfo(),
		function:    getLastCall(routine),
		pan:         panBool,
	}

	elem.findPartner(this)

	this.AddElement(&elem)
	return nil
}

// ========================================================
// MARK: ID
// ========================================================

// ResourceID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementChannel) ResourceID() int {
	return this.objId
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementChannel) Resource() Resource {
	return this.trace.resources[this.objId]
}

// ========================================================
// MARK: Timestamps
// ========================================================

// T returns the t of the element
//
// Parameter:
//   - t timeType: timer type
//
// Returns:
//   - int: The tPre of the element
func (this *ElementChannel) T(t timeType) int {
	switch t {
	case Request:
		return this.tReq
	case Commit:
		return this.tCom
	case Sorting:
		if this.tCom == 0 {
			return math.MaxInt
		}
		return this.tCom
	}

	return this.tCom
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
// - time int: The tPre and tPost of the element
func (this *ElementChannel) SetT(t timeType, time int) {
	switch t {
	case Request:
		this.tReq = time
		if this.tCom != 0 && this.tCom < time {
			this.tCom = time
		}

		if this.sel != nil {
			this.sel.SetTPre2(time)
		}
	case Commit:
		this.tCom = time
		if this.sel != nil {
			this.sel.SetTPost2(time)
		}
	case Sorting:
		this.SetT(Request, time)
		this.tCom = time

		if this.sel != nil {
			this.sel.SetTSort2(time)
		}
	case Both:
		this.SetT(Request, time)
		this.SetT(Commit, time)
	}
	this.tReq = time
	this.tCom = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementChannel) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.tCom != 0 {
		this.tCom = tSort
	}

	if this.sel != nil {
		this.sel.SetTWithoutNotExecuted2(tSort)
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementChannel) Committed() bool {
	return this.tCom != 0
}

// SetTPre2 sets the tPre of the element. It does not set the tPre of the select operation
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementChannel) SetTPre2(tPre int) {
	this.tReq = tPre
	if this.tCom != 0 && this.tCom < tPre {
		this.tCom = tPre
	}
}

// SetTPost2 sets the tPost of the element. It does not set the tPost of the select operation
//
// Parameter:
//   - tPost int: The tPost of the element
func (this *ElementChannel) SetTPost2(tPost int) {
	this.tCom = tPost
}

// SetTSort2 sets the timer, that is used for the sorting of the trace.
// It does not set the tPost of the select operation
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementChannel) SetTSort2(tPost int) {
	this.SetT(Request, tPost)
	this.tCom = tPost
}

// SetTWithoutNotExecuted2 sets the timer, that is used for the sorting of the trace, only if the original
// value was not 0. Do not set the tPost of the select operation
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementChannel) SetTWithoutNotExecuted2(tSort int) {
	this.SetT(Request, tSort)
	if this.tCom != 0 {
		this.tCom = tSort
	}
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementChannel) Pos() Position {
	return this.pos
}

// File returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementChannel) File() string {
	return this.pos.file
}

// Line returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementChannel) Line() int {
	return this.pos.line
}

// ========================================================
// MARK: Operation
// ========================================================

// GetObjType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementChannel) Type(operation bool) OperationType {
	if !operation {
		return Channel
	}

	return this.op
}

// ========================================================
// MARK: Request (gui)
// ========================================================

func (this *ElementChannel) CanBeRequest() bool {
	return this.op != ChannelClose
}

// ========================================================
// MARK: Equal
// ========================================================

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementChannel) IsEqual(elem Element) bool {
	return this.objId == elem.ResourceID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same channel
func (this *ElementChannel) IsSameElement(elem Element) bool {
	if elem.Type(false) != Channel {
		return false
	}

	return this.objId == elem.ResourceID()
}

// ========================================================
// MARK: String
// ========================================================

// String returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementChannel) String() string {
	return this.toStringSep(",", false)
}

// ToStringSep returns the simple string representation of the element with a
// custom separator
//
// Parameter:
//   - sep string: The separator between the values
//   - sel bool: Whether the channel is part of a select do not add time and pos)
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementChannel) toStringSep(sep string, sel bool) string {
	opFull := string(this.op)
	op := "?"
	if len(opFull) > 1 {
		op = string(opFull[1])
	}

	cl := "f"
	if this.cl {
		cl = "t"
	}

	timeString := ""
	posStr := ""
	if !sel {
		timeString = fmt.Sprintf("%s%d%s%d", sep, this.T(Request), sep, this.T(Commit))
		posStr = sep + this.Pos().String()
	}

	return fmt.Sprintf("C%s%s%d%s%s%s%s%s%d%s%d%s%d%s", timeString, sep, this.objId, sep, op, sep, cl, sep, this.oID, sep, this.qSize, sep, this.qCount, posStr)
}

func (this *ElementChannel) StringLocal() string {
	opFull := string(this.op)
	op := "?"
	if len(opFull) > 1 {
		op = string(opFull[1])
	}

	cl := "f"
	if this.cl {
		cl = "t"
	}

	timeString := ""
	posStr := ""
	timeString = fmt.Sprintf(",%d,%d", this.T(Request), this.T(Commit))
	posStr = "," + this.Pos().Short()

	sep := ","

	return fmt.Sprintf("C%s%s%d%s%s%s%s%s%d%s%d%s%d%s", timeString, sep, this.objId, sep, op, sep, cl, sep, this.oID, sep, this.qSize, sep, this.qCount, posStr)
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementChannel) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	res := fmt.Sprintf("%s@%s", routine, this.String())

	if this.pan {
		res += ",🗲"
	}

	return res
}

// StringGui returns the gui string representation of the element.
//
// Returns:
//   - string: The gui string representation of the element
func (this *ElementChannel) StringGui() string {
	opString := string(string(this.op)[1])

	res := fmt.Sprintf("C,%d,%s\n%s", this.objId, opString, this.Pos().Short())

	if this.pan {
		res += ",🗲"
	}

	return res
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementChannel) StringMin() string {
	pos := this.Pos().Short()
	if !flags.ShowFile {
		pos = fmt.Sprintf("%d", this.Pos().Line())
	}

	res := fmt.Sprintf("%s,%d,%s", string(this.op), this.Resource().shortID, pos)

	if this.pan {
		res += ",🗲"
	}

	return res
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementChannel) Function() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

// Vc sets the vector clock
//
// Parameter:
//   - weak bool: set the weak wv
//   - cl *clock.VectorClock: the vector clock
func (this *ElementChannel) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementChannel) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
	return this.ci.getVC(weak)
}

// NumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Parameter:
//   - weak bool: get number of weak concurrent
//   - sameElem bool: only operation on the same variable
//
// Returns:
//   - number of concurrent element, or -1
func (this *ElementChannel) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementChannel) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementChannel) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, this.pos.file, this.pos.line)
}

// Replayable returns if element can be part of replay
func (this *ElementChannel) Replayable() bool {
	return true
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy creates a copy of the channel element
//
//   - trace *Trace: the new trace
//   - mapping map[string]Element: map containing all already copied elements,
//     Used to avoid double copy of references
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementChannel) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	id := this.ID()
	if existing, ok := mapping[id]; ok {
		return existing
	}

	if !keep {
		newCh := ElementChannel{
			ElementBase: this.ElementBase.Copy(trace),
			tReq:        0,
			tCom:        0,
			objId:       this.objId,
			op:          this.op,
			cl:          false,
			oID:         0,
			qSize:       this.qSize,
			qCount:      0,
			pos:         this.pos.copy(),
			selIndex:    this.selIndex,
			ci:          newConcInfo(),
			function:    this.function.CopyFunc(trace, mapping, keep),
		}

		mapping[id] = &newCh

		var newPartner *ElementChannel
		if this.partner != nil {
			newPartner = this.partner.Copy(trace, mapping, keep).(*ElementChannel)
		}

		var newSelect *ElementSelect
		if this.sel != nil {
			newSelect = this.sel.Copy(trace, mapping, keep).(*ElementSelect)
		}

		newCh.partner = newPartner
		newCh.sel = newSelect

		return &newCh
	}

	newCh := ElementChannel{
		ElementBase: this.ElementBase.Copy(trace),
		tReq:        this.tReq,
		tCom:        this.tCom,
		objId:       this.objId,
		op:          this.op,
		cl:          this.cl,
		oID:         this.oID,
		qSize:       this.qSize,
		qCount:      this.qCount,
		pos:         this.pos.copy(),
		selIndex:    this.selIndex,
		ci:          this.ci.copy(),
		function:    this.function.CopyFunc(trace, mapping, keep),
	}

	mapping[id] = &newCh

	var newPartner *ElementChannel
	if this.partner != nil {
		newPartner = this.partner.Copy(trace, mapping, keep).(*ElementChannel)
	}

	var newSelect *ElementSelect
	if this.sel != nil {
		newSelect = this.sel.Copy(trace, mapping, keep).(*ElementSelect)
	}

	newCh.partner = newPartner
	newCh.sel = newSelect

	return &newCh
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementChannel) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

// GetPartner returns the partner of the channel operation
//
// Returns:
//   - *TraceElementChannel: The partner of the channel operation
func (this *ElementChannel) GetPartner() *ElementChannel {
	return this.partner
}

// GetOID returns the operation ID of the element
//
// Returns:
//   - int: The oID of the element
func (this *ElementChannel) GetOID() int {
	return this.oID
}

// IsBuffered returns if the channel is buffered
//
// Returns:
//   - bool: Whether the channel operation is buffered
func (this *ElementChannel) IsBuffered() bool {
	return this.qSize != 0
}

// GetQCount returns the number of elems in the queue after the operation
//
// Returns:
//   - int: The number of elems in the queue after the operation
func (this *ElementChannel) GetQCount() int {
	return this.qCount
}

// GetQCount sets the number of elems in the queue after the operation
//
// Parameter:
//   - qCount int: The number of elems in the queue after the operation
func (this *ElementChannel) SetQCount(qc int) {
	this.qCount = qc
}

// GetQSize returns the size of the buffer
//
// Returns:
//   - int: the size of the buffer
func (this *ElementChannel) GetQSize() int {
	return this.qSize
}

// GetSelect returns the select the element is a part of, if it is not part
// of a select, it returns nil
//
// Returns:
//   - VectorClock: The select the element is a part of, if not in select it is nil
func (this *ElementChannel) GetSelect() *ElementSelect {
	return this.sel
}

// SetPartner sets the partner of the channel operation
//
// Parameter:
//   - partner *TraceElementChannel: The partner of the channel operation
func (this *ElementChannel) SetPartner(partner *ElementChannel) {
	this.partner = partner
}

// SetClosed sets the cl value to closed
//
// Parameter:
//   - closed bool: the new cl value
func (this *ElementChannel) SetClosed(closed bool) {
	this.cl = closed
}

// GetClosed returns if the channel was closed
//
// Returns:
//   - bool: cl
func (this *ElementChannel) GetClosed() bool {
	return this.cl
}

// SetOID sets the operation ID of the element
//
// Parameter:
//   - oID int: The oID of the element
func (this *ElementChannel) SetOID(oID int) {
	this.oID = oID
}

// Find the partner of the channel operation
//
// Parameter:
//   - tr *Trace: the trace, the element is in
//
// Returns:
//   - *TraceElementChannel: The partner, -1 if not found
func (this *ElementChannel) findPartner(tr *Trace) *ElementChannel {
	id := this.ResourceID()
	oID := this.GetOID()

	// return -1 if closed by channel
	if this.GetClosed() || !this.Committed() {
		return nil
	}

	// find partner has already been applied to the partner and the communication
	// was fund. An repeated search is not necessary
	if this.GetPartner() != nil {
		return this.GetPartner()
	}

	// check if partner has already been processed
	if partner, ok := tr.channelWithoutPartner[id][oID]; ok {
		if this.IsEqual(partner) {
			return nil
		}

		// partner was already processed
		this.SetPartner(partner)
		partner.SetPartner(this)

		delete(tr.channelWithoutPartner[id], oID)

		return partner
	}

	if tr.channelWithoutPartner[id] == nil {
		tr.channelWithoutPartner[id] = make(map[int]*ElementChannel)
	}
	tr.channelWithoutPartner[id][oID] = this

	return nil
}
