// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/mutex.go
// Brief: Struct and functions for mutex operations in the trace
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

// ElementMutex is a trace element for a mutex
//
// Fields:
//   - objId int: The id of the mutex
//   - tReq int: The timestamp at the start of the event
//   - tCom int: The timestamp at the end of the event
//   - pos position: code position
//   - ci *concInfo: concurrency info
//   - rw bool: Whether the mutex is a read-write mutex
//   - op ObjectType: The operation on the mutex
//   - suc bool: Whether the operation was successful (only for trylock else always true)
//   - function *ElementFunc: the function the operation is in
//   - pan bool: caused panic
type ElementMutex struct {
	ElementBase

	objId    int
	tReq     int
	tCom     int
	pos      Position
	ci       *concInfo
	rw       bool
	op       OperationType
	suc      bool
	function *ElementFunc
	pan      bool
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementMutex adds a new mutex element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tReq string: The timestamp at the start of the event
//   - tCom string: The timestamp at the end of the event
//   - id string: The id of the mutex
//   - rw string: Whether the mutex is a read-write mutex
//   - opM string: The operation on the mutex
//   - suc string: Whether the operation was successful (only for trylock else always true)
//   - pos string: The position of the mutex operation in the code
//   - pan string: Cased a panic
func (this *Trace) AddTraceElementMutex(routine int, tReq string,
	tCom string, id string, rw string, opM string, suc string,
	pos string, pan string) error {
	tReqInt, err := strconv.Atoi(tReq)
	if err != nil {
		return errors.New("tReq is not an integer")
	}

	tComInt, err := strconv.Atoi(tCom)
	if err != nil {
		return errors.New("tCom is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	rwBool := false
	if rw == "t" {
		rwBool = true
	}

	var opMInt OperationType
	switch opM {
	case "L":
		opMInt = MutexLock
	case "R":
		opMInt = MutexRLock
	case "T":
		opMInt = MutexTryLock
	case "Y":
		opMInt = MutexTryRLock
	case "U":
		opMInt = MutexUnlock
	case "N":
		opMInt = MutexRUnlock
	default:
		return errors.New("opM is not a valid operation")
	}

	sucBool, err := strconv.ParseBool(suc)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	panBool := false
	if pan == "P" {
		panBool = true
	}

	elem := ElementMutex{
		ElementBase: this.newElementBase(this, routine),
		tReq:        tReqInt,
		tCom:        tComInt,
		objId:       idInt,
		rw:          rwBool,
		op:          opMInt,
		suc:         sucBool,
		pos:         NewPosition(file, line),
		ci:          newConcInfo(),
		function:    getLastCall(routine),
		pan:         panBool,
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
func (this *ElementMutex) ResourceID() int {
	return this.objId
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementMutex) Resource() Resource {
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
func (this *ElementMutex) T(t timeType) int {
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
//   - t timeType: type of time to set
//   - time int: The tPre and tPost of the element
func (this *ElementMutex) SetT(t timeType, time int) {
	switch t {
	case Request:
		this.tReq = time
		if this.tCom != 0 && this.tCom < time {
			this.tCom = time
		}
	case Commit:
		this.tCom = time
		if time != 0 && this.tReq > time {
			this.tReq = time
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
func (this *ElementMutex) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.tCom != 0 {
		this.tCom = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementMutex) Committed() bool {
	return this.tCom != 0
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementMutex) Pos() Position {
	return this.pos
}

// File returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementMutex) File() string {
	return this.pos.file
}

// Line returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementMutex) Line() int {
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
func (this *ElementMutex) Type(operation bool) OperationType {
	if !operation {
		return Mutex
	}

	return this.op
}

// ========================================================
// MARK: Request (gui)
// ========================================================

func (this *ElementMutex) CanBeRequest() bool {
	switch this.op {
	case MutexLock, MutexRLock:
		return true
	}
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
func (this *ElementMutex) IsEqual(elem Element) bool {
	return this.objId == elem.ResourceID() && this.id == elem.ID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same mutex
func (this *ElementMutex) IsSameElement(elem Element) bool {
	if elem.Type(false) != Mutex {
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
func (this *ElementMutex) String() string {
	res := "M,"
	res += strconv.Itoa(this.tReq) + "," + strconv.Itoa(this.tCom) + ","
	res += strconv.Itoa(this.objId) + ","

	if this.rw {
		res += "R,"
	} else {
		res += "-,"
	}

	res += string(string(this.op)[1])

	if this.suc {
		res += ",t"
	} else {
		res += ",f"
	}
	res += "," + this.Pos().String()
	return res
}

func (this *ElementMutex) StringLocal() string {
	res := "M,"
	res += strconv.Itoa(this.tReq) + "," + strconv.Itoa(this.tCom) + ","
	res += strconv.Itoa(this.objId) + ","

	if this.rw {
		res += "R,"
	} else {
		res += "-,"
	}

	res += string(string(this.op)[1])

	if this.suc {
		res += ",t"
	} else {
		res += ",f"
	}
	res += "," + this.Pos().Short()
	return res
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementMutex) StringDebug() string {
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

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The gui string representation of the element
func (this *ElementMutex) StringGui() string {
	res := "M,"
	res += strconv.Itoa(this.objId) + ","

	if this.rw {
		res += "R,"
	} else {
		res += "-,"
	}

	res += string(string(this.op)[1])

	if this.suc {
		res += ",t"
	} else {
		res += ",f"
	}
	res += "\n" + this.Pos().Short()

	if this.pan {
		res += ",🗲"
	}

	return res
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementMutex) StringMin() string {
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

func (this *ElementMutex) Function() *ElementFunc {
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
func (this *ElementMutex) Vc(weak a_clock.VcType, cl *a_clock.VectorClock) {
	this.ci.setVC(weak, cl)
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool: get the weak
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementMutex) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementMutex) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementMutex) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementMutex) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, this.pos.file, this.pos.line)
}

// Replayable returns if element can be part of replay
func (this *ElementMutex) Replayable() bool {
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
func (this *ElementMutex) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	if !keep {
		return &ElementMutex{
			ElementBase: this.ElementBase.Copy(trace),
			tReq:        0,
			tCom:        0,
			objId:       this.objId,
			rw:          this.rw,
			op:          this.op,
			suc:         true,
			pos:         this.pos.copy(),
			ci:          newConcInfo(),
			function:    this.function.CopyFunc(trace, mapping, keep),
		}
	}

	return &ElementMutex{
		ElementBase: this.ElementBase.Copy(trace),
		tReq:        this.tReq,
		tCom:        this.tCom,
		objId:       this.objId,
		rw:          this.rw,
		op:          this.op,
		suc:         this.suc,
		pos:         this.pos.copy(),
		ci:          this.ci.copy(),
		function:    this.function.CopyFunc(trace, mapping, keep),
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementMutex) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

// IsLock returns if the element is a lock operation
//
// Returns:
//   - bool: If the element is a lock operation
func (this *ElementMutex) IsLock() bool {
	return this.op == MutexLock || this.op == MutexRLock || this.op == MutexTryLock || this.op == MutexTryRLock
}

// IsSuc returns whether the locking was successful of the element
//
// Returns:
//   - For trylock wether it was successful, otherwise always true
func (this *ElementMutex) IsSuc() bool {
	return this.suc
}

// IsSuc sets whether the locking was successful of the element
//
// Parameter:
//   - s bool: For trylock wether it was successful, otherwise always true
func (this *ElementMutex) SetSuc(s bool) {
	this.suc = s
}
