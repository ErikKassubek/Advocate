// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/element.go
// Brief: Interface for all trace element types
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"advocate/utils/consts"
	"advocate/utils/flags"
	"fmt"
	"path/filepath"
	"strings"
)

// ========================================================
// MARK: Element
// ========================================================

// Element is an interface for the elements in a trace
type Element interface {
	ID() int
	setID(ID int)
	ResourceID() int
	Resource() Resource

	T(t timeType) int
	SetT(t timeType, time int)
	SetTWithoutNotExecuted(tSort int)
	Committed() bool

	Pos() Position
	File() string
	Line() int

	RoutineID() int
	Routine() *Routine
	TraceIndex() (int, int)

	Type(operation bool) OperationType

	IsEqual(elem Element) bool
	IsSameElement(elem Element) bool

	Trace() *Trace
	setTrace(trace *Trace)

	Request() bool
	CanBeRequest() bool
	SetRequest(req bool)

	String() string
	StringMin() string
	StringLocal() string
	StringDebug() string
	StringGui() string

	Function() *ElementFunc

	Vc(weak a_clock.VcType, vc *a_clock.VectorClock)
	GetVC(weak a_clock.VcType) *a_clock.VectorClock

	NumberConcurrent(weak, sameElem bool) int
	SetNumberConcurrent(c int, weak, sameElem bool)

	ReplayID() string

	Copy(trace *Trace, mapping map[int]Element, keep bool) Element

	IsValid() bool

	InInit() bool

	Replayable() bool
}

func IsOp(elem Element) bool {
	switch elem.(type) {
	case *ElementAlloc, *ElementRoutineEnd:
		return false
	}

	return true
}

func isReqStr(elem Element) string {
	if elem.Request() {
		return "R"
	}
	return "C"
}

// ========================================================
// MARK: Base
// ========================================================

type ElementBase struct {
	id        int
	index     int
	routineId int
	routine   *Routine

	trace *Trace

	init bool

	request bool // can only be true if trace.request is true
}

func (this *Trace) newElementBase(trace *Trace, routID int) ElementBase {
	this.minTraceID++
	rout := this.routines[routID]
	return ElementBase{id: this.minTraceID, routineId: routID, index: this.NumberElemInRoutine(routID), routine: rout, trace: trace, init: !this.hasPassedMain}
}

// ID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementBase) ID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementBase) setID(ID int) {
	this.id = ID
}

func (this *ElementBase) Routine() *Routine {
	return this.routine
}

func (this *ElementBase) RoutineID() int {
	return this.routine.id
}

func (this *ElementBase) TraceIndex() (int, int) {
	return this.routineId, this.index
}

func (this *ElementBase) Request() bool {
	return this.request
}

func (this *ElementBase) SetRequest(req bool) {
	this.request = req
}

func (this *ElementBase) CanBeRequest() bool {
	return false
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (e ElementBase) Copy(trace *Trace) ElementBase { // TODO: fix copy, especially alloc
	new_e := e
	new_e.setTrace(trace)
	return new_e
}

func (e ElementBase) InInit() bool {
	return e.init
}

func (e ElementBase) Trace() *Trace {
	return e.trace
}

func (e ElementBase) setTrace(trace *Trace) {
	e.trace = trace
}

// ========================================================
// MARK: Operation
// ========================================================

// Values for possible primitive types and functions
type OperationType string

const (
	None OperationType = ""

	Atomic            OperationType = "A"
	AtomicLoad        OperationType = "AL"
	AtomicStore       OperationType = "AS"
	AtomicAdd         OperationType = "AA"
	AtomicAnd         OperationType = "AN"
	AtomicOr          OperationType = "AO"
	AtomicSwap        OperationType = "AW"
	AtomicCompAndSwap OperationType = "AC"

	Channel      OperationType = "C"
	ChannelSend  OperationType = "CS"
	ChannelRecv  OperationType = "CR"
	ChannelClose OperationType = "CC"

	Cond          OperationType = "D"
	CondWait      OperationType = "DW"
	CondSignal    OperationType = "DS"
	CondBroadcast OperationType = "DB"

	Fork   OperationType = "G"
	ForkOp OperationType = "GG"

	End        OperationType = "E"
	EndRoutine OperationType = "EG"

	Mutex         OperationType = "M"
	MutexLock     OperationType = "ML"
	MutexRLock    OperationType = "MR"
	MutexTryLock  OperationType = "MT"
	MutexTryRLock OperationType = "MY"
	MutexUnlock   OperationType = "MU"
	MutexRUnlock  OperationType = "MN"

	New        OperationType = "N"
	NewAtomic  OperationType = "NA"
	NewChannel OperationType = "NC"
	NewCond    OperationType = "ND"
	NewMutex   OperationType = "NM"
	NewOnce    OperationType = "NO"
	NewWait    OperationType = "NW"

	Once     OperationType = "O"
	OnceSuc  OperationType = "OS"
	OnceFail OperationType = "OF"

	Replay   OperationType = "R"
	ReplayOP OperationType = "RR"

	Select   OperationType = "S"
	SelectOp OperationType = "SS"

	Wait     OperationType = "W"
	WaitAdd  OperationType = "WA"
	WaitDone OperationType = "WD"
	WaitWait OperationType = "WW"

	Func       OperationType = "F"
	FuncCall   OperationType = "FC"
	FuncReturn OperationType = "FR"

	Controll       OperationType = "I"
	ControllIf     OperationType = "II"
	ControllSwitch OperationType = "IS"
	ControllLoop   OperationType = "IL"

	UnknownOperation OperationType = "XX"
)

// GetElemTypeFromObjectType returns the object type from the operation type
//
// Parameter:
//   - ob OperationType: the operation or object type
//
// Returns:
//   - OperationType: the corresponding object type
func GetElemTypeFromObjectType(ob OperationType) OperationType {
	switch ob {
	case Atomic, AtomicLoad, AtomicStore, AtomicAdd, AtomicAnd, AtomicOr, AtomicSwap, AtomicCompAndSwap:
		return Atomic
	case Channel, ChannelSend, ChannelRecv, ChannelClose:
		return Channel
	case Cond, CondWait, CondSignal, CondBroadcast:
		return Cond
	case Fork, ForkOp:
		return Fork
	case End, EndRoutine:
		return End
	case Mutex, MutexLock, MutexRLock, MutexTryLock, MutexUnlock, MutexRUnlock:
		return Mutex
	case Once, OnceSuc, OnceFail:
		return Once
	case Replay, ReplayOP:
		return Replay
	case Select, SelectOp:
		return Select
	case Wait, WaitAdd, WaitDone, WaitWait:
		return Wait
	case Func, FuncCall, FuncReturn:
		return Func
	default:
		return None
	}
}

// ========================================================
// MARK: Time
// ========================================================

type timeType int

const (
	Request timeType = iota
	Commit
	Both = iota
	Sorting
)

// ========================================================
// MARK: Concurrent
// ========================================================

type concInfo struct {
	vc                       *a_clock.VectorClock
	wVc                      *a_clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

func newConcInfo() *concInfo {
	return &concInfo{nil, nil, -1, -1, -1, -1}
}

func (this *concInfo) copy() *concInfo {
	return &concInfo{
		this.vc.Copy(),
		this.wVc.Copy(),
		this.numberConcurrent,
		this.numberConcurrentWeak,
		this.numberConcurrentSame,
		this.numberConcurrentWeakSame,
	}
}

func (this *concInfo) getVC(weak a_clock.VcType) *a_clock.VectorClock {
	if weak == a_clock.Weak {
		return this.wVc
	}
	return this.vc
}

func (this *concInfo) setVC(weak a_clock.VcType, vc *a_clock.VectorClock) {
	if weak == a_clock.Weak {
		this.wVc = vc
		return
	}
	this.vc = vc
}

func (this *concInfo) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return this.numberConcurrentWeakSame
		}
		return this.numberConcurrentWeak
	}
	if sameElem {
		return this.numberConcurrentSame
	}
	return this.numberConcurrent
}

func (this *concInfo) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			this.numberConcurrentWeakSame = c
		} else {
			this.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			this.numberConcurrentSame = c
		} else {
			this.numberConcurrent = c
		}
	}
}

// ========================================================
// MARK: Position
// ========================================================

type Position struct {
	file string
	line int
}

func NewPosition(file string, line int) Position {
	return Position{file, line}
}

func (this Position) File() string {
	return this.file
}

func (this Position) Line() int {
	return this.line
}

func (this Position) copy() Position {
	return Position{
		this.file, this.line,
	}
}

func (this Position) String() string {
	return fmt.Sprintf("%s%s%d", this.file, consts.PosSep, this.line)
}

func (this Position) Short() string {
	pathPre := flags.ProgPath
	if pathPre == "" {
		pathPre = strings.Split(flags.TracePath, "/advocateResult/")[0]
	}

	pp := filepath.Dir(pathPre)
	if strings.HasSuffix(pathPre, ".go") {
		pp = filepath.Dir(pp)
	}
	shortFile := strings.TrimPrefix(this.file, pp)

	if shortFile != this.file {
		shortFile = "." + shortFile
	} else if strings.Contains(shortFile, "/goPatch/") {
		shortFile = "goPatch/" + strings.Split(shortFile, "/goPatch/")[1]
	}
	res := fmt.Sprintf("%s%s%d", shortFile, consts.PosSep, this.line)

	return res
}
