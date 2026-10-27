// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/routineEnd.go
// Brief: Struct and functions for wait group operations in the trace
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

// OpWait enum
type OpWait int

// ========================================================
// MARK: Data
// ========================================================

// ElementWait is a trace element for a wait group statement
//
// Fields:
//   - objId int: The id of the wait group
//   - op OperationType: The operation on the wait group
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - pos position: code position
//   - ci *concInfo: concurrency info
//   - delta int: The delta of the wait group
//   - val int: The value of the wait group
//   - function *ElementFunc: the function the operation is in
type ElementWait struct {
	ElementBase

	objId    int
	op       OperationType
	tPre     int
	tPost    int
	pos      Position
	ci       *concInfo
	delta    int
	val      int
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementWait adds a new wait group element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the wait group
//   - opW string: The operation on the wait group
//   - delta string: The delta of the wait group
//   - val string: The value of the wait group
//   - pos string: The position of the wait group in the code
func (this *Trace) AddTraceElementWait(routine int, tPre,
	tPost, id, opW, delta, val, pos string) error {
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

	deltaInt, err := strconv.Atoi(delta)
	if err != nil {
		return errors.New("delta is not an integer")
	}

	opWOp := None
	if opW == "W" {
		opWOp = WaitWait
	} else if deltaInt > 0 {
		opWOp = WaitAdd
	} else {
		opWOp = WaitDone
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return errors.New("val is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementWait{
		ElementBase: this.newElementBase(this, routine),
		tPre:        tPreInt,
		tPost:       tPostInt,
		objId:       idInt,
		op:          opWOp,
		delta:       deltaInt,
		val:         valInt,
		pos:         NewPosition(file, line),
		ci:          newConcInfo(),
		function:    getLastCall(routine),
	}

	this.AddElement(&elem)

	return nil
}

// Return an empty wait element with an id. Mainly used for source/drain in
// st-graph to detect potential negative wait group
//
// Parameter:
//   - id int: the id of the element
//
// Returns:
//   - ElementWait: the wait element
func EmptyWait(id int) ElementWait {
	return ElementWait{
		ElementBase: ElementBase{id: id},
		objId:       id,
	}
}

// ========================================================
// MARK: ID
// ========================================================

// ResourceID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementWait) ResourceID() int {
	return this.objId
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementWait) Resource() Resource {
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
func (this *ElementWait) T(t timeType) int {
	switch t {
	case Request:
		return this.tPre
	case Commit:
		return this.tPost
	case Sorting:
		if this.tPost == 0 {
			return math.MaxInt
		}
		return this.tPost
	}

	return this.tPost
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - t timeType: type of time to set
//   - time int: The tPre and tPost of the element
func (this *ElementWait) SetT(t timeType, time int) {
	switch t {
	case Request:
		this.tPre = time
		if this.tPost != 0 && this.tPost < time {
			this.tPost = time
		}
	case Commit:
		this.tPost = time
		if time != 0 && this.tPre > time {
			this.tPre = time
		}
	case Sorting, Both:
		this.SetT(Request, time)
		this.SetT(Commit, time)
	}
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementWait) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementWait) Committed() bool {
	return this.tPost != 0
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementWait) Pos() Position {
	return this.pos
}

// File returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementWait) File() string {
	return this.pos.file
}

// Line returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementWait) Line() int {
	return this.pos.line
}

// ========================================================
// MARK: Operation
// ========================================================

// Type returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementWait) Type(operation bool) OperationType {
	if !operation {
		return Wait
	}

	return this.op
}

// IsWait returns if the operation is a wait op
//
// Returns:
//   - bool: True if the operation is a wait op
func (this *ElementWait) IsWait() bool {
	return this.op == WaitWait
}

// ========================================================
// MARK: Request (gui)
// ========================================================

func (this *ElementWait) CanBeRequest() bool {
	return this.op == WaitWait
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
func (this *ElementWait) IsEqual(elem Element) bool {
	return this.objId == elem.ResourceID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same w3ait group
func (this *ElementWait) IsSameElement(elem Element) bool {
	if elem.Type(false) != Wait {
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
func (this *ElementWait) String() string {
	res := "W,"
	res += strconv.Itoa(this.tPre) + "," + strconv.Itoa(this.tPost) + ","
	res += strconv.Itoa(this.objId) + ","
	switch this.op {
	case WaitAdd, WaitDone:
		res += "A,"
	case WaitWait:
		res += "W,"
	}

	res += strconv.Itoa(this.delta) + "," + strconv.Itoa(this.val)
	res += "," + this.Pos().String()
	return res
}

func (this *ElementWait) StringLocal() string {
	res := "W,"
	res += strconv.Itoa(this.tPre) + "," + strconv.Itoa(this.tPost) + ","
	res += strconv.Itoa(this.objId) + ","
	switch this.op {
	case WaitAdd, WaitDone:
		res += "A,"
	case WaitWait:
		res += "W,"
	}

	res += strconv.Itoa(this.delta) + "," + strconv.Itoa(this.val)
	res += "," + this.Pos().Short()
	return res
}

// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementWait) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

// StringGui returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementWait) StringGui() string {
	res := "W,"
	res += strconv.Itoa(this.objId) + ","
	switch this.op {
	case WaitAdd, WaitDone:
		res += "A,"
	case WaitWait:
		res += "W,"
	}

	res += strconv.Itoa(this.delta) + "," + strconv.Itoa(this.val)
	res += "\n" + this.Pos().Short()
	return res
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementWait) StringMin() string {
	pos := this.Pos().Short()
	if flags.HideFile {
		pos = fmt.Sprintf("%d", this.Pos().Line())
	}

	return fmt.Sprintf("%s,%d,%s", string(this.op), this.Resource().shortID, pos)
}

// String returns the simple string representation of the element with leading routine

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementWait) Function() *ElementFunc {
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
func (this *ElementWait) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementWait) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementWait) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementWait) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementWait) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, this.pos.file, this.pos.line)
}

// Replayable returns if element can be part of replay
func (this *ElementWait) Replayable() bool {
	return true
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy the element
//
// Parameter:
//   - mapping map[string]Element: map containing all already copied elements.
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementWait) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	if !keep {
		return &ElementWait{
			ElementBase: this.ElementBase.Copy(trace),
			tPre:        0,
			tPost:       0,
			objId:       this.objId,
			op:          this.op,
			delta:       this.delta,
			val:         0,
			pos:         this.pos.copy(),
			ci:          newConcInfo(),
			function:    this.function.CopyFunc(trace, mapping, keep),
		}
	}

	return &ElementWait{
		ElementBase: this.ElementBase.Copy(trace),
		tPre:        this.tPre,
		tPost:       this.tPost,
		objId:       this.objId,
		op:          this.op,
		delta:       this.delta,
		val:         this.val,
		pos:         this.pos.copy(),
		ci:          this.ci.copy(),
		function:    this.function.CopyFunc(trace, mapping, keep),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementWait) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

// GetDelta returns the delta of the element. The delta is the value by which the counter
// of the wait has been changed. For Add the delta is > 0, for Done it is -1,
// for Wait it is 0
//
// Returns:
//   - int: the delta of the wait element
func (this *ElementWait) GetDelta() int {
	return this.delta
}

// SetVal sets the value of the internal counter
//
// Parameter:
//   - v int: the new value
func (this *ElementWait) SetVal(v int) {
	this.val = v
}
