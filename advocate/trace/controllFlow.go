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
	"errors"
	"fmt"
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementControllFlow is a trace element for if/switch
// Fields:
//   - t int: The timestamp of the int
//   - numCases int: Number Cases
//   - chosenCase int: Chosen case index (0 based)
//   - ci *concInfo: concurrency info
//   - function *ElementFunc: the function the operation is in
//   - pos position: code position
//   - function *ElementFunc: the function the element is in
type ElementControllFlow struct {
	ElementBase

	t          int
	numCases   int
	chosenCase int
	op         OperationType
	pos        Position
	function   *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementAlloc adds a make trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - t string: The timestamp at the event
//   - op string: I/S
//   - numCases string: Number of cases
//   - chosenCase string: Chosen case (0 based)
//   - pos string: position
func (this *Trace) AddTraceElementControllFlow(routine int, t, op, numCases, chosenCase, pos string) error {
	tInt, err := strconv.Atoi(t)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	nc, err := strconv.Atoi(numCases)
	if err != nil {
		return errors.New("numCases is not an integer")
	}

	cc, err := strconv.Atoi(chosenCase)
	if err != nil {
		return errors.New("chosenCase is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	o := UnknownOperation
	switch op {
	case "I":
		o = ControllIf
	case "S":
		o = ControllSwitch
	case "L":
		o = ControllLoop
	}

	elem := ElementControllFlow{
		ElementBase: this.newElementBase(this, routine),
		t:           tInt,
		numCases:    nc,
		chosenCase:  cc,
		op:          o,
		pos:         NewPosition(file, line),
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
func (this *ElementControllFlow) ResourceID() int {
	return -1
}

// setObjId sets the object id
//
// Parameter:
//   - id int: the object id
func (this *ElementControllFlow) setObjId(id int) {
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementControllFlow) Resource() Resource {
	return NewResource(-1, nil)
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
func (this *ElementControllFlow) Type(operation bool) OperationType {
	if !operation {
		return Controll
	}

	return this.op
}

// ========================================================
// MARK: Timestamps
// ========================================================

// T returns the timestamp of the element
//
// Returns:
//   - int: The tPre of the element
func (this *ElementControllFlow) T(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementControllFlow) SetT(_ timeType, tSort int) {
	this.t = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementControllFlow) SetTWithoutNotExecuted(tSort int) {
	if this.t == 0 {
		return
	}
	this.t = tSort
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementControllFlow) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementControllFlow) Pos() Position {
	return this.pos
}

// File returns the file where the operation represented by the element was executed
//
// Returns:
//   - int: The file of the element
func (this *ElementControllFlow) File() string {
	return this.pos.file
}

// GetTraceID sets the file
//
// Parameter:
//   - f string: the file
func (this *ElementControllFlow) setFile(f string) {
	this.pos.file = f
}

// Line returns the line where the operation represented by the element was executed
//
// Returns:
//   - int: The line of the element
func (this *ElementControllFlow) Line() int {
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
func (this *ElementControllFlow) IsEqual(elem Element) bool {
	return this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: always false
func (this *ElementControllFlow) IsSameElement(elem Element) bool {
	return this.IsEqual(elem)
}

// ========================================================
// MARK: String
// ========================================================

// String returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementControllFlow) String() string {
	opStr := ""
	switch this.op {
	case ControllIf:
		opStr = "I"
	case ControllSwitch:
		opStr = "S"
	case ControllLoop:
		opStr = "L"
	default:
		panic("Invalid op in Controll Flow Element")
	}
	return fmt.Sprintf("I,%d,%s,%d,%d,%s", this.t, opStr, this.numCases, this.chosenCase, this.Pos())
}

func (this *ElementControllFlow) StringLocal() string {
	opStr := ""
	switch this.op {
	case ControllIf:
		opStr = "I"
	case ControllSwitch:
		opStr = "S"
	case ControllLoop:
		opStr = "L"
	default:
		panic("Invalid op in Controll Flow Element")
	}
	return fmt.Sprintf("I,%d,%s,%d,%d,%s", this.t, opStr, this.numCases, this.chosenCase, this.Pos().Short())
}

// StringGui returns the simple gui representation of the element
//
// Returns:
//   - string: The simple gui representation of the element
func (this *ElementControllFlow) StringGui() string {
	opStr := ""
	switch this.op {
	case ControllIf:
		opStr = "I"
	case ControllSwitch:
		opStr = "S"
	case ControllLoop:
		opStr = "L"
	default:
		panic("Invalid op in Controll Flow Element")
	}
	return fmt.Sprintf("I,%s,%d,%d\n%s", opStr, this.numCases, this.chosenCase, this.Pos().Short())
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementControllFlow) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementControllFlow) StringMin() string {
	return fmt.Sprintf("%s,-,%s", string(this.op), this.Pos().Short())
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementControllFlow) Function() *ElementFunc {
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
func (this *ElementControllFlow) Vc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementControllFlow) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
	return &a_clock.VectorClock{}
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
func (this *ElementControllFlow) NumberConcurrent(_, _ bool) int {
	return -1
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementControllFlow) SetNumberConcurrent(_ int, _, _ bool) {

}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay ID of the element
//
// Returns:
//   - int: The replayId of the element
func (this *ElementControllFlow) ReplayID() string {
	return ""
}

// Replayable returns if element can be part of replay
func (this *ElementControllFlow) Replayable() bool {
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
func (this *ElementControllFlow) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	return &ElementControllFlow{
		ElementBase: this.ElementBase.Copy(trace),
		t:           this.t,
		numCases:    this.numCases,
		chosenCase:  this.chosenCase,
		op:          this.op,
		pos:         this.pos.copy(),
		function:    this.function.CopyFunc(trace, mapping, keep),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementControllFlow) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

func (this *ElementControllFlow) ChosenCase() int {
	return this.chosenCase
}

func (this *ElementControllFlow) NumberCases() int {
	return this.numCases
}
