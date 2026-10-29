// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/atomic.go
// Brief: Struct and functions for atomic operations in the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"advocate/utils/flags"
	"advocate/utils/log"
	"errors"
	"fmt"
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementAtomic is a struct to save an atomic event in the trace
// Fields:
//
//   - objId int: The id of the atomic variable
//   - op ObjectType: The operation on the atomic variable
//   - t int: The timestamp of the event
//   - pos position: code position
//   - ci *concInfo: concurrency info
//   - function *ElementFunc: the function the operation is in
type ElementAtomic struct {
	ElementBase

	objId    int
	op       OperationType
	t        int
	pos      Position
	ci       *concInfo
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementAtomic adds a new atomic trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp of the event
//   - id string: The id of the atomic variable
//   - operation string: The operation on the atomic variable
//   - pos string: The position of the atomic
func (this *Trace) AddTraceElementAtomic(routine int, tPost string,
	id string, operation string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	var opAInt OperationType
	switch operation {
	case "L":
		opAInt = AtomicLoad
	case "S":
		opAInt = AtomicStore
	case "A":
		opAInt = AtomicAdd
	case "W":
		opAInt = AtomicSwap
	case "C":
		opAInt = AtomicCompAndSwap
	case "N":
		opAInt = AtomicAnd
	case "O":
		opAInt = AtomicOr
	default:
		return fmt.Errorf("Atomic operation '%s' is not a valid operation", operation)
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		log.Error("Cannot read pos string ", pos)
		return err
	}

	elem := ElementAtomic{
		ElementBase: this.newElementBase(this, routine),
		t:           tPostInt,
		objId:       idInt,
		op:          opAInt,
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
func (this *ElementAtomic) ResourceID() int {
	return this.objId
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementAtomic) Resource() Resource {
	return this.trace.resources[this.objId]
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
func (this *ElementAtomic) Type(operation bool) OperationType {
	if !operation {
		return Atomic
	}

	return this.op
}

// ========================================================
// MARK: Timestamps
// ========================================================

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (this *ElementAtomic) T(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementAtomic) SetT(_ timeType, time int) {
	this.t = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementAtomic) SetTWithoutNotExecuted(tSort int) {
	if this.t != 0 {
		this.t = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementAtomic) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementAtomic) Pos() Position {
	return this.pos
}

// File returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementAtomic) File() string {
	return this.pos.file
}

// Line returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementAtomic) Line() int {
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
func (this *ElementAtomic) IsEqual(elem Element) bool {
	return this.objId == elem.ResourceID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same atomic variable
func (this *ElementAtomic) IsSameElement(elem Element) bool {
	return this.objId == elem.ResourceID()
}

// ========================================================
// MARK: String
// ========================================================

// String returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementAtomic) String() string {
	opString := string(string(this.op)[1])

	return fmt.Sprintf("A,%d,%d,%s,%s", this.t, this.objId, opString, this.Pos())
}

func (this *ElementAtomic) StringLocal() string {
	opString := string(string(this.op)[1])

	return fmt.Sprintf("A,%d,%d,%s,%s", this.t, this.objId, opString, this.Pos().Short())
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementAtomic) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

// StringGui returns the gui string representation of the element.
//
// Returns:
//   - string: The simple gui representation of the element
func (this *ElementAtomic) StringGui() string {
	opString := string(string(this.op)[1])

	return fmt.Sprintf("A,%d,%s\n%s", this.objId, opString, this.Pos().Short())
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementAtomic) StringMin() string {
	pos := this.Pos().Short()
	if !flags.ShowFile {
		pos = fmt.Sprintf("%d", this.Pos().Line())
	}
	return fmt.Sprintf("%s,%d,%s", string(this.op), this.Resource().shortID, pos)

}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementAtomic) Function() *ElementFunc {
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
func (this *ElementAtomic) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementAtomic) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementAtomic) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementAtomic) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementAtomic) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, this.pos.file, this.pos.line)
}

// Replayable returns if element can be part of replay
func (this *ElementAtomic) Replayable() bool {
	return false
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy the atomic element
//
// Parameter:
//   - trace *Trace: the new trace
//   - mapping map[int]Element: map containing all already copied elements, if nil ignore all vc based values.
//     since atomics do not contain reference to other elements and no other
//     elements contain referents to atomics, this is not used
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementAtomic) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {

	if !keep {
		return &ElementAtomic{
			ElementBase: this.ElementBase.Copy(trace),
			t:           0,
			objId:       this.objId,
			op:          this.op,
			pos:         this.pos.copy(),
			ci:          newConcInfo(),
			function:    this.function.CopyFunc(trace, mapping, keep),
		}
	}

	return &ElementAtomic{
		ElementBase: this.ElementBase.Copy(trace),
		t:           this.t,
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

func (this *ElementAtomic) IsValid() bool {
	return this != nil
}
