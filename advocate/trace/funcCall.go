// Copyright (c) 2024 Erik Kassubek
//
// File: funcCall.go
// Brief: Records the call of a function
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
	"strings"
)

var lastCall = make(map[int]*types.Stack[*ElementFunc])

func getLastCall(rout int) *ElementFunc {
	rStack, ok := lastCall[rout]
	if !ok {
		return nil
	}

	if rStack.IsEmpty() {
		return nil
	}

	return rStack.Peek()
}

// ========================================================
// MARK: Data
// ========================================================

// ElementFunc is a struct to save a function call in the trace
// Fields:
//
//   - name: name of the called function
//   - ssaName: ssa name of the called function
//   - t int: The timestamp of the event
//   - posDef position: code position of the function declaration
//   - posCall position: code position of the function call
//   - function *ElementFunc: the function the operation is in
type ElementFunc struct {
	ElementBase

	name     string
	ssaName  string
	t        int
	posDef   Position
	posCall  Position
	function *ElementFunc
}

// ========================================================
// MARK: Constructor
// ========================================================

func (this *Trace) AddTaceElementFunc(routine int, t string, name string, posDef, posCall string) error {
	tInt, err := strconv.Atoi(t)
	if err != nil {
		return errors.New("t is not an integer")
	}

	fileDef, lineDef, err := PosFromPosString(posDef)
	if err != nil {
		return err
	}

	fileCall, lineCall, err := PosFromPosString(posCall)
	if err != nil {
		return err
	}

	if name == "main.main" { // main func
		this.hasPassedMain = true
	}

	elem := ElementFunc{
		ElementBase: this.newElementBase(this, routine),
		name:        name,
		ssaName:     funcNameToSSANane(name),
		t:           tInt,
		posDef:      NewPosition(fileDef, lineDef),
		posCall:     NewPosition(fileCall, lineCall),
		function:    getLastCall(routine),
	}

	if _, ok := lastCall[routine]; !ok {
		lastCall[routine] = types.NewStack[*ElementFunc]()
	}

	lc := lastCall[routine].Peek()

	this.callTree.AddElem(lc, &elem)

	lastCall[routine].Push(&elem)

	this.AddElement(&elem)
	return nil
}

// ========================================================
// MARK: ID
// ========================================================

func (this *ElementFunc) ResourceID() int {
	return -1
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementFunc) Resource() Resource {
	return NewResource(-1, nil)
}

// ========================================================
// MARK: Timestamps
// ========================================================

func (this *ElementFunc) T(_ timeType) int {
	return this.t
}

func (this *ElementFunc) SetT(_ timeType, time int) {
	this.t = time
}

func (this *ElementFunc) SetTWithoutNotExecuted(t int) {
	this.t = t
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementFunc) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

func (this *ElementFunc) Pos() Position {
	return this.posCall
}

func (this *ElementFunc) File() string {
	return this.posCall.file
}

func (this *ElementFunc) Line() int {
	return this.posCall.line
}

func (this *ElementFunc) PosDef() Position {
	return this.posDef
}

// ========================================================
// MARK: Operation
// ========================================================

func (this *ElementFunc) Type(operation bool) OperationType {
	if operation {
		return FuncCall
	}

	return Func
}

// ========================================================
// MARK: Equal
// ========================================================

func (this *ElementFunc) IsEqual(elem Element) bool {
	return this.id == elem.ID()
}

func (this *ElementFunc) IsSameElement(elem Element) bool {
	switch e := elem.(type) {
	case *ElementFunc:
		return this.ssaName == e.ssaName
	}

	return false
}

// ========================================================
// MARK: String
// ========================================================

func (this *ElementFunc) String() string {
	return fmt.Sprintf("F,%d,%s,%s,%s", this.t, this.name, this.PosDef(), this.Pos())
}

func (this *ElementFunc) StringLocal() string {
	return fmt.Sprintf("F,%d,%s,%s,%s", this.t, this.name, this.PosDef().Short(), this.Pos().Short())
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementFunc) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.RoutineID())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

func (this *ElementFunc) StringGui() string {
	return fmt.Sprintf("F,%s,%s\n%s", this.name, this.PosDef().Short(), this.Pos().Short())
}

// StringGui returns the gui string representation of the element
//
// Returns:
//   - string: The minimal string representation of the element
func (this *ElementFunc) StringMin() string {
	return fmt.Sprintf("%s,-,%s", string(FuncCall), this.Pos().Short())
}

// ========================================================
// MARK: VC
// ========================================================

func (this *ElementFunc) Vc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

func (this *ElementFunc) GetVC(a_clock.VcType) *a_clock.VectorClock {
	return &a_clock.VectorClock{}
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementFunc) Function() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

func (this *ElementFunc) NumberConcurrent(_, _ bool) int {
	return 0
}

func (this *ElementFunc) SetNumberConcurrent(_ int, _, _ bool) {
}

// ========================================================
// MARK: Replay
// ========================================================

func (this *ElementFunc) ReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routineId, this.posCall.file, this.posCall.line)
}

// Replayable returns if element can be part of replay
func (this *ElementFunc) Replayable() bool {
	return false
}

// ========================================================
// MARK: Copy
// ========================================================

func (this *ElementFunc) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	if this == nil {
		return nil
	}

	id := this.ID()

	if existing, ok := mapping[id]; ok {
		return existing
	}

	elem := &ElementFunc{
		ElementBase: this.ElementBase.Copy(trace),
		t:           this.t,
		name:        this.name,
		ssaName:     this.ssaName,
		posDef:      this.posDef.copy(),
		posCall:     this.posCall.copy(),
	}

	mapping[id] = elem

	return elem
}

func (this *ElementFunc) CopyFunc(trace *Trace, mapping map[int]Element, keep bool) *ElementFunc {
	if this == nil {
		return nil
	}

	var funcCopy *ElementFunc

	if fc, ok := this.function.Copy(trace, mapping, keep).(*ElementFunc); ok {
		funcCopy = fc
	}

	return funcCopy
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementFunc) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

func (this *ElementFunc) Name() string {
	return this.name
}

func (this *ElementFunc) GetSSAName() string {
	return this.ssaName
}

func funcNameToSSANane(name string) string {
	if name == "" {
		return name
	}

	fields := strings.Split(name, ".")

	if fields[0] == "main" {
		fields[0] = flags.ExecName
	}

	if len(fields) >= 3 {
		l := len(fields) - 1
		if strings.HasPrefix(fields[l], "func") {
			fields[l-1] = fields[l-1] + "$" + strings.TrimPrefix(fields[l], "func")
			fields = fields[:l]
		}
	}

	return strings.Join(fields, ".")

}
