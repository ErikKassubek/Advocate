// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/select.go
// Brief: Struct and functions for select operations in the trace
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
	"strings"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementSelect is a trace element for a select statement
// Fields:
//   - objId int: The id of the select statement
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - pos position: code position
//   - ci *concInfo: concurrency info
//   - cases []*ElementChannel: The cases of the select statement, ordered by casi starting from 0
//   - chosenCase *ElementChannel: The chosen case, nil if default case chosen
//   - chosenIndex int: The internal index of chosen case
//   - containsDefault bool: Whether the select statement contains a default case
//   - chosenDefault bool: if the default case was chosen
//   - function *ElementFunc: the function the operation is in
//   - pan bool: Caused a panic
type ElementSelect struct {
	ElementBase

	objId               int
	tPre                int
	tPost               int
	pos                 Position
	ci                  *concInfo
	cases               []*ElementChannel
	chosenCase          *ElementChannel
	chosenIndex         int
	containsDefault     bool
	chosenDefault       bool
	casesWithPosPartner []int
	function            *ElementFunc
	pan                 bool
}

// ========================================================
// MARK: Constructor
// ========================================================

// AddTraceElementSelect adds a new select statement element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tReq string: The timestamp at the start of the event
//   - tCom string: The timestamp at the end of the event
//   - id string: The id of the select statement
//   - cases string: The cases of the select statement
//   - chosenIndex string: The internal index of chosen case
//   - pos string: The position of the select statement in the code
//   - pan string: Caused a panic
func (this *Trace) AddTraceElementSelect(routine int, tReq,
	tCom, id, cases, chosenIndex, pos, pan string) error {

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

	chosenIndexInt, err := strconv.Atoi(chosenIndex)
	if err != nil {
		return errors.New("chosenIndex is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	panBool := false
	if pan == "P" {
		panBool = true
	}

	elem := ElementSelect{
		ElementBase:         this.newElementBase(this, routine),
		tPre:                tReqInt,
		tPost:               tComInt,
		objId:               idInt,
		chosenIndex:         chosenIndexInt,
		pos:                 NewPosition(file, line),
		ci:                  newConcInfo(),
		casesWithPosPartner: make([]int, 0),
		function:            getLastCall(routine),
		pan:                 panBool,
	}

	cs := strings.Split(cases, "~")
	casesList := make([]*ElementChannel, 0)
	containsDefault := false
	chosenDefault := false
	for i, c := range cs {
		if c == "" {
			continue
		}

		if c == "d" {
			containsDefault = true
			break
		}
		if c == "D" {
			containsDefault = true
			chosenDefault = true
			break
		}

		// read channel operation
		caseList := strings.Split(c, ".")

		cID := -1
		if caseList[1] != "*" {
			cID, err = strconv.Atoi(caseList[1])
			if err != nil {
				return errors.New("c_id is not an integer")
			}
		}
		var cOpC = ChannelSend
		switch caseList[2] {
		case "R":
			cOpC = ChannelRecv
		case "C":
			return errors.New("Close in select case list")
		}

		cCl, err := strconv.ParseBool(caseList[3])
		if err != nil {
			return errors.New("c_cr is not a boolean")
		}

		cOID, err := strconv.Atoi(caseList[4])
		if err != nil {
			return errors.New("c_oId is not an integer")
		}
		cOSize, err := strconv.Atoi(caseList[5])
		if err != nil {
			return errors.New("c_oSize is not an integer")
		}

		cTPost := 0
		if i == chosenIndexInt {
			cTPost = tComInt
		}

		elemCase := &ElementChannel{
			ElementBase: this.newElementBase(this, routine),
			tReq:        tReqInt,
			tCom:        cTPost,
			objId:       cID,
			op:          cOpC,
			cl:          cCl,
			oID:         cOID,
			qSize:       cOSize,
			sel:         &elem,
			selIndex:    len(caseList),
			pos:         NewPosition(file, line),
			ci:          newConcInfo(),
		}

		casesList = append(casesList, elemCase)
		if elemCase.tCom != 0 {
			elem.chosenCase = elemCase
			elemCase.findPartner(this)
		}
	}

	elem.containsDefault = containsDefault
	elem.chosenDefault = chosenDefault
	elem.cases = casesList

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
func (this *ElementSelect) ResourceID() int {
	return this.objId
}

// ResourceID returns the resource
//
// Returns:
//   - Resource: resource
func (this *ElementSelect) Resource() Resource {
	return NewResource(-1, nil)
}

// ResourceID returns the resources
//
// Returns:
//   - Resource: resource
func (this *ElementSelect) Resources() map[Resource]bool {
	res := make(map[Resource]bool)

	for _, c := range this.cases {
		res[c.Resource()] = true
	}

	return res
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
func (this *ElementSelect) T(t timeType) int {
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
//   - time int: The tPre and tPost of the element
func (this *ElementSelect) SetT(t timeType, time int) {
	switch t {
	case Request:
		this.tPre = time
		if this.tPost != 0 && this.tPost < time {
			this.tPost = time
		}

		for _, c := range this.cases {
			c.SetTPre2(time)
		}
	case Commit:
		this.tPost = time
		this.chosenCase.SetTPost2(time)
	case Sorting:
		this.SetT(Request, Sorting)
		this.tPost = Sorting
	case Both:
		this.tPre = time
		this.tPost = time

		this.chosenCase.tCom = time

		for i := range this.cases {
			this.cases[i].tReq = time
		}
	}

}

// SetTPre2 sets the tPre of the element. It does not update the chosen case
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementSelect) SetTPre2(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}

	for _, c := range this.cases {
		c.SetTPre2(tPre)
	}
}

// SetTPost2 sets the tPost. It does not update the chosen case
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementSelect) SetTPost2(tPost int) {
	this.tPost = tPost
}

// SetTSort2 set the timer, that is used for the sorting of the trace.
// It does not update the chosen case
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementSelect) SetTSort2(tSort int) {
	this.SetTPre2(tSort)
	this.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter: tSort int: The timer of the element
func (this *ElementSelect) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
	this.chosenCase.SetTWithoutNotExecuted2(tSort)
}

// SetTWithoutNotExecuted2 sets the timer, that is used for the sorting of the trace, only if the original
// value was not 0. Do not update the chosen case
//
// Parameter: tSort int: The timer of the element
func (this *ElementSelect) SetTWithoutNotExecuted2(tSort int) {
	this.SetTPre2(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementSelect) Committed() bool {
	return this.tPost != 0
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementSelect) Pos() Position {
	return this.pos
}

// File returns the file where the operation represented by the element was executed
//
// Returns:
//   - string: The file of the element
func (this *ElementSelect) File() string {
	return this.pos.file
}

// Line returns the line where the operation represented by the element was executed
//
// Returns:
//   - string: The line of the element
func (this *ElementSelect) Line() int {
	return this.pos.line
}

// ========================================================
// MARK: Request (gui)
// ========================================================

func (this *ElementSelect) CanBeRequest() bool {
	return true
}

// ========================================================
// MARK: Operation
// ========================================================

// GetCases returns the cases of the select statement
//
// Returns:
//   - []traceElementChannel: The cases of the select statement
func (this *ElementSelect) GetCases() []*ElementChannel {
	return this.cases
}

// Type returns he object type
//
// Parameter:
//   - operations bool: if true, the operation id contains the operations, otherwise just that it is select
//
// Returns:
//   - the object type
func (this *ElementSelect) Type(operation bool) OperationType {
	if !operation {
		return Select
	}

	return SelectOp
}

// Type returns the object type of the cases
//
// Parameter:
//   - operations bool: if true, the operation id contains the operations, otherwise just that it is select
//
// Returns:
//   - the object type
func (this *ElementSelect) Types() []OperationType {
	res := make([]OperationType, 0)

	for _, c := range this.cases {
		res = append(res, c.Type(true))
	}

	return res
}

// ========================================================
// MARK: Panic
// ========================================================

func (this *ElementSelect) Pan() bool {
	return this.pan
}

// ========================================================
// MARK: Equal
// ========================================================

// IsEqual checks if the given element is equal to the select
//
// Parameter:
//   - elem TraceElement: The element
//
// Returns:
//   - bool: true if they are equal, false otherwise
func (this *ElementSelect) IsEqual(elem Element) bool {
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
func (this *ElementSelect) IsSameElement(elem Element) bool {
	return false
}

// ========================================================
// MARK: String
// ========================================================

// String returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementSelect) String() string {
	res := "S" + "," + strconv.Itoa(this.tPre) + "," +
		strconv.Itoa(this.tPost) + "," + strconv.Itoa(this.objId) + ","

	notNil := 0
	for _, ca := range this.cases { // cases
		if ca.tReq != 0 { // ignore nil cases
			if notNil != 0 {
				res += "~"
			}
			res += ca.toStringSep(".", true)
			notNil++
		}
	}

	if this.containsDefault {
		if notNil != 0 {
			res += "~"
		}
		if this.chosenDefault {
			res += "D"
		} else {
			res += "d"
		}
	}
	res += "," + strconv.Itoa(this.chosenIndex)
	res += "," + this.Pos().String()
	return res
}

func (this *ElementSelect) StringLocal() string {
	res := "S" + "," + strconv.Itoa(this.tPre) + "," +
		strconv.Itoa(this.tPost) + "," + strconv.Itoa(this.objId) + ","

	notNil := 0
	for _, ca := range this.cases { // cases
		if ca.tReq != 0 { // ignore nil cases
			if notNil != 0 {
				res += "~"
			}
			res += ca.toStringSep(".", true)
			notNil++
		}
	}

	if this.containsDefault {
		if notNil != 0 {
			res += "~"
		}
		if this.chosenDefault {
			res += "D"
		} else {
			res += "d"
		}
	}
	res += "," + strconv.Itoa(this.chosenIndex)
	res += "," + this.Pos().Short()
	return res
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementSelect) StringDebug() string {
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
func (this *ElementSelect) StringGui() string {
	res := "S" + "," + strconv.Itoa(this.objId) + ","

	notNil := 0
	for _, ca := range this.cases { // cases
		if ca.tReq != 0 { // ignore nil cases
			if notNil != 0 {
				res += "~"
			}
			res += ca.toStringSep(".", true)
			notNil++
		}
	}

	if this.containsDefault {
		if notNil != 0 {
			res += "~"
		}
		if this.chosenDefault {
			res += "D"
		} else {
			res += "d"
		}
	}
	res += "," + strconv.Itoa(this.chosenIndex)
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
func (this *ElementSelect) StringMin() string {
	pos := this.Pos().Short()
	if !flags.ShowFile {
		pos = fmt.Sprintf("%d", this.Pos().Line())
	}

	if this.chosenDefault {
		return fmt.Sprintf("%s,D,%s", SelectOp, pos)
	}

	if this.chosenCase == nil {
		return fmt.Sprintf("%s,-,%s", SelectOp, pos)
	}

	res := fmt.Sprintf("%s,%d,%s", SelectOp, this.chosenCase.Resource().shortID, pos)

	if this.pan {
		res += ",🗲"
	}

	return res
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementSelect) Function() *ElementFunc {
	return this.function
}

// ========================================================
// MARK: Concurrent
// ========================================================

// Vc sets the vector clock
//
// Parameter:
//   - weak bool
//   - vc *clock.VectorClock: the vector clock
func (this *ElementSelect) Vc(weak a_clock.VcType, vc *a_clock.VectorClock) {
	this.ci.setVC(weak, vc)
	if this.chosenCase != nil {
		this.chosenCase.Vc(weak, vc)
	}
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementSelect) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
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
func (this *ElementSelect) NumberConcurrent(weak, sameElem bool) int {
	return this.ci.GetNumberConcurrent(weak, sameElem)
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementSelect) SetNumberConcurrent(c int, weak, sameElem bool) {
	this.ci.SetNumberConcurrent(c, weak, sameElem)

	if this.GetChosenCase() != nil {
		this.GetChosenCase().SetNumberConcurrent(c, weak, sameElem)
	}
}

// ========================================================
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the operations
//
// Returns:
//   - string: The replay id of the element
func (this *ElementSelect) ReplayID() string {
	return fmt.Sprintf("%d:%s#%d", this.routineId, this.pos.file, this.pos.line)
}

// Replayable returns if element can be part of replay
func (this *ElementSelect) Replayable() bool {
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
//     This avoids double copy of referenced elements
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementSelect) Copy(trace *Trace, mapping map[int]Element, keep bool) Element {
	id := this.ID()

	if existing, ok := mapping[id]; ok {
		return existing
	}

	if !keep {
		elem := &ElementSelect{
			ElementBase:     this.ElementBase.Copy(trace),
			tPre:            0,
			tPost:           0,
			objId:           this.objId,
			chosenIndex:     this.chosenIndex,
			containsDefault: this.containsDefault,
			chosenDefault:   this.chosenDefault,
			pos:             this.pos.copy(),
			ci:              newConcInfo(),
			function:        this.function.CopyFunc(trace, mapping, keep),
		}

		mapping[id] = elem

		elem.cases = make([]*ElementChannel, 0)
		for _, c := range this.cases {
			cp := c.Copy(trace, mapping, keep).(*ElementChannel)
			elem.cases = append(elem.cases, cp)
			if cp.Committed() {
				elem.chosenCase = cp
			}
		}

		for _, c := range elem.cases {
			c.sel = elem
		}

		return elem
	}

	elem := &ElementSelect{
		ElementBase:     this.ElementBase.Copy(trace),
		tPre:            this.tPre,
		tPost:           this.tPost,
		objId:           this.objId,
		chosenIndex:     this.chosenIndex,
		containsDefault: this.containsDefault,
		chosenDefault:   this.chosenDefault,
		pos:             this.pos.copy(),
		ci:              this.ci.copy(),
		function:        this.function.CopyFunc(trace, mapping, keep),
	}

	mapping[id] = elem

	elem.cases = make([]*ElementChannel, 0)
	for _, c := range this.cases {
		cp := c.Copy(trace, mapping, keep).(*ElementChannel)
		elem.cases = append(elem.cases, cp)
		if cp.Committed() {
			elem.chosenCase = cp
		}
	}

	for _, c := range elem.cases {
		c.sel = elem
	}

	return elem
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementSelect) IsValid() bool {
	return this != nil
}

// ========================================================
// MARK: Others
// ========================================================

// GetChosenCase returns the chosen case
//
// Returns:
//   - the chosen case
func (this *ElementSelect) GetChosenCase() *ElementChannel {
	if this.chosenDefault || this.tPost == 0 {
		return nil
	}
	return this.chosenCase
}

// GetChosenIndex returns the index of the chosen case in se.cases
//
// Returns:
//   - The internal index of the chosen case
func (this *ElementSelect) GetChosenIndex() int {
	return this.chosenIndex
}

// GetContainsDefault returns whether the select contains a default case
//
// Returns:
//   - bool: true if select contains default, false otherwise
func (this *ElementSelect) GetContainsDefault() bool {
	return this.chosenDefault
}

// GetPartner returns the communication partner of the select. If there is none,
// it returns nil
//
// Returns:
//   - *TraceElementChannel: The communication partner of the select or nil
func (this *ElementSelect) GetPartner() *ElementChannel {
	if this.chosenCase.tCom != 0 && !this.chosenDefault {
		return this.chosenCase.partner
	}
	return nil
}

// GetCasiWithPosPartner returns a list of all internal indices, where the
// corresponding case as a potential partner
//
// Returns:
//   - []int: list of indices
func (this *ElementSelect) GetCasiWithPosPartner() []int {
	return this.casesWithPosPartner
}

// AddCasesWithPosPartner adds an casi to casesWithPosPartner
//
// Parameter:
//   - casi int: the case id to add
func (this *ElementSelect) AddCasesWithPosPartner(casi int) {
	this.casesWithPosPartner = append(this.casesWithPosPartner, casi)
}

// GetCasesWithPosPartner returns casesWithPosPartner
//
// Returns:
//   - []int: list of cases with potential partner
func (this *ElementSelect) GetCasesWithPosPartner() []int {
	return this.casesWithPosPartner
}

// SetChosenCase sets the chosen case of a select
//
// Parameter:
//   - index of the case that should be set as the chosen case
//
// Returns:
//   - error
func (this *ElementSelect) SetChosenCase(index int) error {
	if index >= len(this.cases) {
		return fmt.Errorf("Invalid index %d for size %d", index, len(this.cases))
	}
	this.cases[this.chosenIndex].tCom = 0
	this.chosenIndex = index
	this.cases[index].tCom = this.tPost

	return nil
}

// GetChosenDefault if the default case is the executed case
//
// Returns: bool: true if default case
func (this *ElementSelect) GetChosenDefault() bool {
	return this.chosenDefault
}

// SetCaseByIndex set the case to the case at the given index or default if index = -1
//
// Parameter:
//   - index of the case, -1 for default
//
// Returns:
//   - error
func (this *ElementSelect) SetCaseByIndex(index int) error {
	if index > len(this.cases) {
		return fmt.Errorf("Invalid index for select: %d [%d]", index, len(this.cases))
	}

	for i := range this.cases {
		this.cases[i].SetT(Commit, 0)
	}

	if index < 0 {
		this.chosenDefault = true
		this.chosenIndex = -1
		return nil
	}

	this.cases[index].SetT(Commit, this.T(Commit))
	this.chosenIndex = index
	this.chosenDefault = false
	return nil
}

// SetCase set the case where the channel id and direction is correct as the active one
//
// Parameter:
//   - chanID int: id of the channel in the case, -1 for default
//   - send opChannel: channel operation of case
//
// Returns:
//   - error
func (this *ElementSelect) SetCase(chanID int, op OperationType) error {
	if chanID == -1 {
		if this.containsDefault {
			this.chosenDefault = true
			this.chosenIndex = -1
			for i := range this.cases {
				this.cases[i].SetT(Commit, 0)
			}
			return nil
		}

		return fmt.Errorf("Tried to set select without default to default")
	}

	found := false
	for i, c := range this.cases {
		if c.objId == chanID && c.op == op {
			tPost := this.T(Commit)
			if !this.chosenDefault {
				this.cases[this.chosenIndex].SetT(Commit, 0)
			} else {
				this.chosenDefault = false
			}
			this.cases[i].SetT(Commit, tPost)
			this.chosenIndex = i
			this.chosenDefault = false
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Select case not found")
	}

	return nil
}

// HasCommonChannels returns if the set of cases that are in both the receiver
// select and the argument select is not empty. We do not consider the default case
//
// Parameter:
//   - s *trace.ElementSelect: the other select
//
// Returns:
//   - bool: true if this and s have at least one common channel
func (this *ElementSelect) HasCommonChannel(s *ElementSelect) bool {
	seen := make(map[int]struct{}, len(this.GetCases()))

	for _, v := range this.GetCases() {
		seen[v.objId] = struct{}{}
	}

	for _, v := range s.GetCases() {
		if _, ok := seen[v.objId]; ok {
			return true
		}
	}

	return false
}
