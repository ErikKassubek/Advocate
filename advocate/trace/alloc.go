// Copyright (c) 2026 Erik Kassubek
//
// File: alloc.go
// Brief: Trace element to store the creation (new) of relevant operations. For now this is only creates the new for channel. This may be expanded later.
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
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementAlloc is a trace element for the creation of an object / new
// Fields:
//   - objId int: The id of the underlying operation
//   - elemType newOpType: The type of the created object
//   - t int: The timestamp of the new
//   - pos *position: code position
//   - ci *concInfo: concurrency info
//   - num int: variable field for additional information
//   - function *ElementFunc: the function the operation is in
//   - init bool: true if in init
type ElementAlloc struct {
	ElementBase

	objId    int
	elemType OperationType
	t        int
	pos      Position
	ci       *concInfo
	num      int
	function *ElementFunc
	init     bool
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementAlloc adds a make trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - t string: The timestamp at the event
//   - id string: The id of the channel
//   - elemType string: Type of the created primitive
//   - num string: Variable field for additional information
//   - pos string: position
func (this *Trace) AddTraceElementAlloc(routine int, t string, id string, elemType string, num string, pos string) error {
	tInt, err := strconv.Atoi(t)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	numInt, err := strconv.Atoi(num)
	if err != nil {
		return errors.New("num is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	et := None
	switch elemType {
	case "C":
		et = NewChannel
	case "D":
		et = NewCond
	case "M":
		et = NewMutex
	case "W":
		et = NewWait
	default:
		panic("Unknown alloc type: " + elemType)
	}

	elem := ElementAlloc{
		ElementBase: this.newElementBase(this, routine),
		t:           tInt,
		objId:       idInt,
		elemType:    et,
		num:         numInt,
		pos:         NewPosition(file, line),
		ci:          newConcInfo(),
		function:    getLastCall(routine),
	}

	this.allocs[idInt] = &elem

	this.AddElement(&elem)

	return nil
}

func (this *Trace) AddTraceElementAllocFromElem(elem Element) {
	var et OperationType
	switch elem.(type) {
	case *ElementChannel:
		et = NewChannel
	case *ElementCond:
		et = NewCond
	case *ElementMutex:
		et = NewMutex
	case *ElementWait:
		et = NewWait
	case *ElementOnce:
		et = NewOnce
	case *ElementAtomic:
		et = NewAtomic
	default:
		return
	}

	rout := elem.RoutineID()
	id := elem.ResourceID()

	al := ElementAlloc{
		ElementBase: this.newElementBase(this, rout),
		t:           elem.T(Request) - 1,
		objId:       id,
		elemType:    et,
		num:         -1,
		pos:         elem.Pos(),
		ci:          newConcInfo(),
		function:    elem.Function(),
	}

	this.allocs[id] = &al

	this.AddElement(&al)
}

// ========================================================
// MARK: ID
// ========================================================

// ResourceID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementAlloc) ResourceID() int {
	return this.objId
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementAlloc) Resource() Resource {
	return this.trace.resources[this.objId]
}

// setObjId sets the object id
//
// Parameter:
//   - id int: the object id
func (this *ElementAlloc) setObjId(id int) {
	this.objId = id
}

// ========================================================
// MARK: Index
// ========================================================

// RoutineID returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementAlloc) RoutineID() int {
	return this.routineId
}

// TraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementAlloc) TraceIndex() (int, int) {
	return this.routineId, this.index
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
func (this *ElementAlloc) Type(operation bool) OperationType {
	if !operation {
		return New
	}

	return this.elemType
}

// ========================================================
// MARK: Timestamps
// ========================================================

// T returns the timestamp of the element
//
// Returns:
//   - int: The tPre of the element
func (this *ElementAlloc) T(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementAlloc) SetT(_ timeType, tSort int) {
	this.t = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementAlloc) SetTWithoutNotExecuted(tSort int) {
	if this.t == 0 {
		return
	}
	this.t = tSort
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementAlloc) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//
//	position: the position
func (this *ElementAlloc) Pos() Position {
	return this.pos
}

// File returns the file where the operation represented by the element was executed
//
// Returns:
//   - int: The file of the element
func (this *ElementAlloc) File() string {
	return this.pos.file
}

// GetTraceID sets the file
//
// Parameter:
//   - f string: the file
func (this *ElementAlloc) setFile(f string) {
	this.pos.file = f
}

// Line returns the line where the operation represented by the element was executed
//
// Returns:
//   - int: The line of the element
func (this *ElementAlloc) Line() int {
	return this.pos.line
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
func (this *ElementAlloc) IsEqual(elem Element) bool {
	return this.objId == elem.ResourceID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: always false
func (this *ElementAlloc) IsSameElement(elem Element) bool {
	return this.objId == elem.ResourceID()
}

// ========================================================
// MARK: String
// ========================================================

// String returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementAlloc) String() string {
	return fmt.Sprintf("N,%d,%d,%s,%d,%s", this.t, this.objId, string(this.elemType), this.num, this.Pos())
}

func (this *ElementAlloc) StringLocal() string {
	return fmt.Sprintf("N,%d,%d,%s,%d,%s", this.t, this.objId, string(this.elemType), this.num, this.Pos().Short())
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementAlloc) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The gui string representation of the element
func (this *ElementAlloc) StringGui() string {
	return fmt.Sprintf("N,%d,%s,%d\n%s", this.objId, string(this.elemType), this.num, this.Pos().Short())
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementAlloc) StringMin() string {
	pos := this.Pos().Short()
	if flags.HideFile {
		pos = fmt.Sprintf("%d", this.Pos().Line())
	}

	return fmt.Sprintf("%s,%d,%s", string(this.elemType), this.Resource().shortID, pos)
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementAlloc) Function() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

// Vc sets the vector clock
//
// Parameter:
//   - weak bool: set the weak wv
//   - cl *clock.allocVectorClock: the vector clock
func (this *ElementAlloc) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementAlloc) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementAlloc) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementAlloc) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay ID of the element
//
// Returns:
//   - int: The replayId of the element
func (this *ElementAlloc) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, this.pos.file, this.pos.line)
}

// Replayable returns if element can be part of replay
func (this *ElementAlloc) Replayable() bool {
	return false
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy the element
//
// Parameter:
//   - trace *Trace: the new trace
//   - mapping map[string]Element: map containing all already copied elements.
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementAlloc) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	return &ElementAlloc{
		ElementBase: this.ElementBase.Copy(trace),
		t:           0,
		objId:       this.objId,
		elemType:    this.elemType,
		pos:         this.pos.copy(),
		ci:          this.ci.copy(),
		function:    this.function.CopyFunc(trace, mapping, keep),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementAlloc) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

// GetNum returns the num field of the element
//
// Returns:
//   - VectorClock: The num field of the element
func (this *ElementAlloc) GetNum() int {
	return this.num
}

// IsRequest determines if the element is a request
// Returns:
//   - bool: element is request
func (this *ElementAlloc) IsRequest() bool {
	return false
}

// IsRequest determines if the element can be a request
// Returns:
//   - bool: element can be request
func (this *ElementAlloc) CanBeRequest() bool {
	return false
}

// SetRequest set request
// Argument:
//   - bool: element is request
func (this *ElementAlloc) SetRequest(_ bool) {
}

// SetRequest sets the routine id
// Argument:
//   - int: new routine id
func (this *ElementAlloc) SetRoutine(id int) {
	this.routineId = id
}
