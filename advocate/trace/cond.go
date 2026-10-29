// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/cond.go
// Brief: Struct and functions for operations of conditional variables in the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"advocate/utils/flags"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementCond is a trace element for a condition variable
// Fields:
//   - tReq int: The timestamp at the start of the event
//   - tCom int: The timestamp at the end of the event
//   - objId int: The id of the condition variable
//   - op objectType: The operation on the condition variable
//   - pos position: code position
//   - ci *concInfo: concurrency infocalculated
//   - function *ElementFunc: the function the operation is in
type ElementCond struct {
	ElementBase

	tReq     int
	tCom     int
	objId    int
	op       OperationType
	pos      Position
	ci       *concInfo
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementCond adds a new condition variable element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the condition variable
//   - opC string: The operation on the condition variable
//   - pos string: The position of the condition variable operation in the code
func (this *Trace) AddTraceElementCond(routine int, tPre string, tPost string, id string, opN string, pos string) error {
	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tPre is not an integer")
	}
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}
	var op OperationType
	switch opN {
	case "W":
		op = CondWait
	case "S":
		op = CondSignal
	case "B":
		op = CondBroadcast
	default:
		return errors.New("op is not a valid operation")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementCond{
		ElementBase: this.newElementBase(this, routine),
		tReq:        tPreInt,
		tCom:        tPostInt,
		objId:       idInt,
		op:          op,
		pos:         NewPosition(file, line),
		ci:          newConcInfo(),
		function:    getLastCall(routine),
	}

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
func (this *ElementCond) ResourceID() int {
	return this.objId
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementCond) Resource() Resource {
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
func (this *ElementCond) T(t timeType) int {
	switch t {
	case Request:
		return this.tReq
	case Commit:
		return this.tCom
	case Sorting:
		t := this.tReq
		if this.op == CondWait {
			t = this.tCom
		}
		if t == 0 {
			// add at the end of the trace
			return math.MaxInt
		}
		return t
	}

	return this.tCom
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - t timeType: type of time to set
//   - time int: The tPre and tPost of the element
func (this *ElementCond) SetT(t timeType, time int) {
	switch t {
	case Request:
		this.tReq = time
		if this.tCom != 0 && this.tCom < time {
			this.tCom = time
		}
	case Commit, Sorting:
		this.tCom = time
		if time != 0 && this.tReq > time {
			this.tReq = time
		}
	case Both:
		this.SetT(Request, time)
		this.SetT(Commit, time)
	}
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementCond) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.op == CondWait {
		if this.tCom != 0 {
			this.tCom = tSort
		}
		return
	}
	if this.tReq != 0 {
		this.tReq = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementCond) Committed() bool {
	return this.tCom != 0
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementCond) Pos() Position {
	return this.pos
}

// File returns the file of the element
//
// Returns:
//   - The file of the element
func (this *ElementCond) File() string {
	return this.pos.file
}

// Line returns the line of the element
//
// Returns:
//   - The line of the element
func (this *ElementCond) Line() int {
	return this.pos.line
}

// ========================================================
// MARK: Operation
// ========================================================

// Type returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementCond) Type(operation bool) OperationType {
	if !operation {
		return Cond
	}

	return this.op
}

// ========================================================
// MARK: Request (gui)
// ========================================================

func (this *ElementCond) CanBeRequest() bool {
	return this.op == CondWait
}

// ========================================================
// MARK: Panic
// ========================================================

func (this *ElementCond) Pan() bool {
	return false
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
func (this *ElementCond) IsEqual(elem Element) bool {
	return this.objId == elem.ResourceID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same conditional variable
func (this *ElementCond) IsSameElement(elem Element) bool {
	if elem.Type(false) != Cond {
		return false
	}

	return this.objId == elem.ResourceID()
}

// ========================================================
// MARK: String
// ========================================================

// String returns the string representation of the element
//
// Returns:
//   - string: The string representation of the element
func (this *ElementCond) String() string {
	res := "D,"
	res += strconv.Itoa(this.tReq) + "," + strconv.Itoa(this.tCom) + ","
	res += strconv.Itoa(this.objId) + ","
	switch this.op {
	case CondWait:
		res += "W"
	case CondSignal:
		res += "S"
	case CondBroadcast:
		res += "B"
	}
	res += "," + this.Pos().String()
	return res
}

func (this *ElementCond) StringLocal() string {
	res := "D,"
	res += strconv.Itoa(this.tReq) + "," + strconv.Itoa(this.tCom) + ","
	res += strconv.Itoa(this.objId) + ","
	switch this.op {
	case CondWait:
		res += "W"
	case CondSignal:
		res += "S"
	case CondBroadcast:
		res += "B"
	}
	res += "," + this.Pos().Short()
	return res
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementCond) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

// StringGui returns the simple gui representation of the element.
//
// Returns:
//   - string: The simple gui representation of the element
func (this *ElementCond) StringGui() string {
	opString := string(string(this.op)[1])

	return fmt.Sprintf("D,%d,%s\n%s", this.objId, opString, this.Pos().Short())
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementCond) StringMin() string {
	pos := this.Pos().Short()
	if !flags.ShowFile {
		pos = fmt.Sprintf("%d", this.Pos().Line())
	}

	return fmt.Sprintf("%s,%d,%s", string(this.op), this.Resource().shortID, pos)
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementCond) Function() *ElementFunc {
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
func (this *ElementCond) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementCond) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementCond) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementCond) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementCond) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, this.pos.file, this.pos.line)
}

// Replayable returns if element can be part of replay
func (this *ElementCond) Replayable() bool {
	return true
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy the element
//
// Parameter:
//   - trace *Trace: the new trace
//   - mapping map[string]Element: map containing all already copied elements.
//     since conds do not contain reference to other elements and no other
//     elements contain referents to conds, this is not used
//   - keep bool: if true, keep vc and order information

// Returns:
//   - TraceElement: The copy of the element
func (this *ElementCond) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	if !keep {
		return &ElementCond{
			ElementBase: this.ElementBase.Copy(trace),
			tReq:        0,
			tCom:        0,
			objId:       this.objId,
			op:          this.op,
			pos:         this.pos.copy(),
			ci:          newConcInfo(),
			function:    this.function.CopyFunc(trace, mapping, keep),
		}
	}

	return &ElementCond{
		ElementBase: this.ElementBase.Copy(trace),
		tReq:        this.tReq,
		tCom:        this.tCom,
		objId:       this.objId,
		op:          this.op,
		pos:         this.pos.copy(),
		ci:          this.ci.copy(),
		function:    this.function.CopyFunc(trace, mapping, keep),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementCond) IsValid() bool {
	return this != nil
}
