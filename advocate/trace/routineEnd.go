// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/routineEnd.go
// Brief: Struct and functions for fork operations in the trace
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

// ElementRoutineEnd is a trace element for the termination of a routine end
// Fields:
//   - t int: The timestamp at the end of the event
//   - ci *concInfo: concurrency info
type ElementRoutineEnd struct {
	ElementBase

	t  int
	ci *concInfo
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementRoutineEnd add a routine and element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the new routine
//   - pos string: The position of the trace element in the file
func (this *Trace) AddTraceElementRoutineEnd(routine int, tPost string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	elem := ElementRoutineEnd{
		ElementBase: this.newElementBase(this, routine),
		t:           tPostInt,
		ci:          newConcInfo(),
	}

	this.AddElement(&elem)

	return nil
}

// ========================================================
// MARK: ID
// ========================================================

// ID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementRoutineEnd) ID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementRoutineEnd) setID(ID int) {
	this.id = ID
}

// ResourceID is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (this *ElementRoutineEnd) ResourceID() int {
	return 0
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementRoutineEnd) Resource() Resource {
	return NewResource(-1, nil)
}

// ========================================================
// MARK: Timestamps
// ========================================================

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPre of the element
func (this *ElementRoutineEnd) T(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementRoutineEnd) SetT(_ timeType, time int) {
	this.t = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementRoutineEnd) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.t != 0 {
		this.t = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementRoutineEnd) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// Pos is a dummy function to implement the traceElement interface
//
// Returns:
//   - position: the position
func (this *ElementRoutineEnd) Pos() Position {
	return NewPosition("", 0)
}

// File is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (this *ElementRoutineEnd) File() string {
	return ""
}

// Line is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (this *ElementRoutineEnd) Line() int {
	return 0
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
//   - string: the object type
func (this *ElementRoutineEnd) Type(operation bool) OperationType {
	if !operation {
		return Rout
	}
	return RoutEnd
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
func (this *ElementRoutineEnd) IsEqual(elem Element) bool {
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
func (this *ElementRoutineEnd) IsSameElement(elem Element) bool {
	return false
}

// ========================================================
// MARK: String
// ========================================================

// String returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementRoutineEnd) String() string {
	return "E" + "," + strconv.Itoa(this.t)
}

func (this *ElementRoutineEnd) StringLocal() string {
	return "E" + "," + strconv.Itoa(this.t)
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementRoutineEnd) StringDebug() string {
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
func (this *ElementRoutineEnd) StringGui() string {
	return string(RoutEnd)
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementRoutineEnd) StringMin() string {
	return fmt.Sprintf("%s,-,-", RoutEnd)
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementRoutineEnd) Function() *ElementFunc {
	return nil
}

// ========================================================
// MARK: Concurrent
// ========================================================

// Vc sets the vector clock
//
// Parameter:
//   - weak bool
//   - vc *clock.VectorClock: the vector clock
func (this *ElementRoutineEnd) Vc(weak a_clock.VcType, vc *a_clock.VectorClock) {
	this.ci.setVC(weak, vc)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementRoutineEnd) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
	return this.ci.getVC(weak)
}

// NumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
func (this *ElementRoutineEnd) NumberConcurrent(_, _ bool) int {
	return -1
}

// SetNumberConcurrent sets the number of concurrent elements
func (this *ElementRoutineEnd) SetNumberConcurrent(_ int, _, _ bool) {}

// GetConcurrent returns the elements that are concurrent to the element
func (this *ElementRoutineEnd) GetConcurrent(_, _ bool) []Element {
	return []Element{}
}

// SetConcurrent sets the concurrent elements
func (this *ElementRoutineEnd) SetConcurrent(_ []Element, _, _ bool) {
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (this *ElementRoutineEnd) ReplayID() string {
	return ""
}

// Replayable returns if element can be part of replay
func (this *ElementRoutineEnd) Replayable() bool {
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
func (this *ElementRoutineEnd) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	if !keep {
		return &ElementRoutineEnd{
			ElementBase: this.ElementBase.Copy(trace),
			t:           0,
			ci:          newConcInfo(),
		}
	}

	return &ElementRoutineEnd{
		ElementBase: this.ElementBase.Copy(trace),
		t:           this.t,
		ci:          this.ci.copy(),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementRoutineEnd) IsValid() bool {
	return this != nil
}
