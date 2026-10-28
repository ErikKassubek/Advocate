// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/fork.go
// Brief: Struct and functions for fork operations in the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"advocate/utils/flags"
	"advocate/utils/types"
	"errors"
	"fmt"
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementFork is a trace element for a go statement
// Fields:
//   - t int: The timestamp at the end of the event
//   - objId int: The id of the new go routine
//   - pos position: code position
//   - ci *concInfo: concurrency info
//   - function *ElementFunc: the function the operation is in
type ElementFork struct {
	ElementBase

	t        int
	objId    int
	pos      Position
	ci       *concInfo
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementFork adds a new go statement element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the new routine
//   - pos string: The position of the trace element in the file
func (this *Trace) AddTraceElementFork(routine int, tPost string, id string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementFork{
		ElementBase: this.newElementBase(this, routine),
		t:           tPostInt,
		objId:       idInt,
		pos:         NewPosition(file, line),
		ci:          newConcInfo(),
		function:    getLastCall(routine),
	}

	if _, ok := lastCall[idInt]; !ok {
		lastCall[idInt] = types.NewStack[*ElementFunc]()
	}
	lastCall[idInt].Push(getLastCall(routine))

	// add element to created routine
	newRoutine := this.GetRoutine(elem.objId)

	newRoutine.setFork(&elem)

	this.AddElement(&elem)
	return nil
}

// ========================================================
// MARK: ID
// ========================================================

// ResourceID returns the ID of the newly created routine
//
// Returns:
//   - int: The id of the new routine
func (this *ElementFork) ResourceID() int {
	return this.objId
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementFork) Resource() Resource {
	return NewResource(-1, nil)
}

// ========================================================
// MARK: Timestamps
// ========================================================

// T returns the t of the element.
//
// Returns:
//   - int: The tPre of the element
func (this *ElementFork) T(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementFork) SetT(_ timeType, time int) {
	this.t = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementFork) SetTWithoutNotExecuted(tSort int) {
	if this.t != 0 {
		this.t = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementFork) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementFork) Pos() Position {
	return this.pos
}

// File returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementFork) File() string {
	return this.pos.file
}

// Line returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementFork) Line() int {
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
func (this *ElementFork) Type(operation bool) OperationType {
	if !operation {
		return Rout
	}
	return RoutSpawn
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
func (this *ElementFork) IsEqual(elem Element) bool {
	return this.objId == elem.ResourceID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same. For fork, all forks are  considered
// to be on the same element
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same channel
func (this *ElementFork) IsSameElement(elem Element) bool {
	return elem.Type(false) == Rout
}

// ========================================================
// MARK: String
// ========================================================

// String returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementFork) String() string {
	return "G" + "," + strconv.Itoa(this.t) + "," + strconv.Itoa(this.objId) +
		"," + this.Pos().String()
}

func (this *ElementFork) StringLocal() string {
	return "G" + "," + strconv.Itoa(this.t) + "," + strconv.Itoa(this.objId) +
		"," + this.Pos().Short()
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementFork) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

// StringGui returns the simple gui representation of the element
//
// Returns:
//   - string: The simple gui representation of the element
func (this *ElementFork) StringGui() string {
	return "G" + "," + strconv.Itoa(this.objId) +
		"\n" + this.Pos().Short()
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementFork) StringMin() string {
	pos := this.Pos().Short()
	if !flags.ShowFile {
		pos = fmt.Sprintf("%d", this.Pos().Line())
	}

	return fmt.Sprintf("%s,%d,%s", RoutSpawn, this.objId, pos)
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementFork) Function() *ElementFunc {
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
func (this *ElementFork) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementFork) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementFork) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementFork) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementFork) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, this.pos.file, this.pos.line)
}

// Replayable returns if element can be part of replay
func (this *ElementFork) Replayable() bool {
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
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementFork) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	if !keep {
		return &ElementFork{
			ElementBase: this.ElementBase.Copy(trace),
			t:           0,
			objId:       this.objId,
			pos:         this.pos.copy(),
			ci:          newConcInfo(),
			function:    this.function.CopyFunc(trace, mapping, keep),
		}
	}

	return &ElementFork{
		ElementBase: this.ElementBase.Copy(trace),
		t:           this.t,
		objId:       this.objId,
		pos:         this.pos.copy(),
		ci:          this.ci.copy(),
		function:    this.function.CopyFunc(trace, mapping, keep),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementFork) IsValid() bool {
	return this != nil
}
