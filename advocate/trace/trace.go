// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/trace.go
// Brief: Functions and structs for the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/utils/control"
	"advocate/utils/log"
	"advocate/utils/types"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// ========================================================
// MARK: Trace
// ========================================================

// Trace is a struct to represent a trace
// Fields:
//   - traces map[int][]TraceElement: the trace element, routineId -> list of elems
//   - hbWasCalc bool: set to true if the vector clock has been calculated for all elements
//   - hasPassedMain bool: true if main func has been passed. Used to determine init
//   - channelWithoutPartner  map[int]map[int]*TraceElementChannel: channel for witch no partner has been found yet, id -> opId -> element
//   - channelIDs map[int]struct{}: all channel ids in the trace
//   - objectAware map[int][]int: for not terminated routines, the blocked objects they can access
//   - blocked map[resource][]int: routines which are blocked and the objects that block it
//   - forks: for each routine store the fork that created it
//   - allocs: allocs
//   - resources: obj id to resource
//   - callGraph: call graph
//   - request bool: if true, requests and commits are separate elements, otherwise they are the same
type Trace struct {
	routines              map[int]*Routine
	numberRoutines        int
	hbWasCalc             bool
	hasPassedMain         bool
	minTraceID            int
	channelWithoutPartner map[int]map[int]*ElementChannel
	channelIDs            map[int]struct{}
	blocked               map[int]Element
	forks                 map[int]*ElementFork
	forkTree              map[int][]int
	allocs                map[int]*ElementAlloc
	resources             map[int]Resource
	callTree              CallTree
	request               bool
}

// NewTrace creates a new empty trace structure
//
// Returns:
//   - Trace: the new trace
func NewTrace() Trace {
	return Trace{
		routines:              make(map[int]*Routine),
		numberRoutines:        0,
		hbWasCalc:             false,
		minTraceID:            0,
		channelWithoutPartner: make(map[int]map[int]*ElementChannel),
		blocked:               make(map[int]Element),
		forks:                 make(map[int]*ElementFork),
		forkTree:              make(map[int][]int),
		allocs:                make(map[int]*ElementAlloc),
		resources:             make(map[int]Resource),
		callTree:              *newCallGraph(),
	}
}

// Clear the trace
func (this *Trace) Clear() {
	this.routines = make(map[int]*Routine)
	this.hbWasCalc = false
	this.numberRoutines = 0
	this.minTraceID = 0
	this.channelWithoutPartner = make(map[int]map[int]*ElementChannel)
	this.blocked = make(map[int]Element)
	this.forks = make(map[int]*ElementFork)
	this.forkTree = make(map[int][]int)
	this.allocs = make(map[int]*ElementAlloc)
	this.resources = make(map[int]Resource)
	this.callTree = *newCallGraph()
}

// AddElement adds an element to the trace
//
// Parameter:
//   - elem TraceElement: Element to add
func (this *Trace) AddElement(elem Element) {
	routine := elem.RoutineID()

	if !elem.Committed() {
		this.blocked[routine] = elem
	}

	switch e := elem.(type) {
	case *ElementFork:
		this.forks[e.objId] = e
		if _, ok := this.forkTree[routine]; !ok {
			this.forkTree[routine] = make([]int, 0)
		}
		this.forkTree[routine] = append(this.forkTree[routine], e.objId)
	}

	this.AddRoutine(routine)

	this.AddResource(elem)

	this.routines[routine].addElement(elem)
}

// AddElement adds a resouce if elem has a resource and if it is not created yet
func (this *Trace) AddResource(elem Element) {
	switch elem.(type) {
	case *ElementFork, *ElementFunc, *ElementReturn, *ElementRoutineEnd, *ElementControllFlow, *ElementSelect:
		return
	}

	id := elem.ResourceID()

	if _, ok := this.resources[id]; ok {
		return
	}

	alloc, ok := this.allocs[id]
	if !ok {
		this.AddTraceElementAllocFromElem(elem)
		alloc, ok = this.allocs[id]
	}
	this.resources[id] = NewResource(id, alloc)
}

// AddRoutine adds an empty routine if not exists
//
// Parameter:
//   - routine int: The routine
func (this *Trace) AddRoutine(routine int) *Routine {
	if _, ok := this.routines[routine]; !ok {
		this.routines[routine] = NewRoutine(routine)
		this.numberRoutines = len(this.routines)
	}

	return this.routines[routine]
}

// GetTraces returns the traces
//
// Returns:
//   - map[int][]traceElement: The traces
func (this *Trace) GetTraces() map[int]*Routine {
	return this.routines
}

// GetTraceSize returns the number of TraceElement with cap and len
func (this *Trace) GetTraceSize() (int, int) {
	capTot := 0
	lenTot := 0
	for _, rout := range this.routines {
		cap, len := rout.size()
		capTot += cap
		lenTot += len
	}
	return capTot, lenTot
}

// GetRoutineTrace returns the trace of the given routine
//
// Parameter:
//   - id int: The id of the routine
//
// Returns:
//   - []Element: The trace of the routine
func (this *Trace) GetRoutineTrace(id int) *Routine {
	if r, ok := this.routines[id]; ok {
		return r
	}

	return this.AddRoutine(id)
}

// GetRoutineTrace returns the trace of the given routine
//
// Parameter:
//   - id int: The id of the routine
//
// Returns:
//   - Element: The last element of the routine or nil if empty
func (this *Trace) GetLastElemInRout(id int) Element {
	return this.routines[id].Last()
}

// GetNumberElements returns the total number of elements in the trace
//
// Returns:
//   - int: total number of elems in trace
func (this *Trace) GetNumberElements() int {
	return this.NumberElemInRoutine(-1)
}

func (this *Trace) GetBlocked() map[int]Element {
	return this.blocked
}

func (this *Trace) GetNotReturned(onlyNonBlocked bool) []int {
	res := make([]int, 0)

	for id, rout := range this.routines {
		if rout.Empty() {
			res = append(res, id)
			continue
		}
		last := rout.Last()
		if last.Type(false) != End && (!onlyNonBlocked || last.Committed()) {
			res = append(res, id)
		}
	}

	return res
}

// GetTraceElementFromBugArg returns the element in the trace,
// given the bug info from the machine readable result file.
//
// Parameter:
//   - bugArg string: The bug info from the machine readable result file
//
// Returns:
//   - *TraceElement: The element
//   - error: An error if the element does not exist
func (this *Trace) GetTraceElementFromBugArg(bugArg string) (Element, error) {
	splitArg := strings.Split(bugArg, "#")

	if splitArg[0] != "T" {
		return nil, errors.New("Bug argument is not a trace element (does not start with T): " + bugArg)
	}

	if len(splitArg) != 7 {
		return nil, errors.New("Bug argument is not a trace element (incorrect number of arguments): " + bugArg)
	}

	routine, err := strconv.Atoi(splitArg[1])
	if err != nil {
		return nil, errors.New("Could not parse routine from bug argument: " + bugArg)
	}

	tPre, err := strconv.Atoi(splitArg[3])
	if err != nil {
		return nil, errors.New("Could not parse tPre from bug argument: " + bugArg)
	}

	if _, ok := this.routines[routine]; !ok {
		return nil, errors.New(fmt.Sprintf("Could not get routine information for routine ID %d", routine))
	}

	for index, elem := range this.routines[routine].elems {
		if elem.T(Request) == tPre {
			return this.routines[routine].At(index), nil
		}
	}

	for id, rout := range this.routines {
		for index, elem := range rout.elems {
			if elem.T(Request) == tPre {
				return this.routines[id].At(index), nil
			}
		}
	}

	return nil, fmt.Errorf("Element %s not in trace", bugArg)
}

// GetNoRoutines returns the number of routines
//
// Returns:
//   - int: The number of routines
func (this *Trace) GetNoRoutines() int {
	return len(this.routines)
}

// GetTraceLength returns the number of element in a given routine
//
// Parameter:
//   - routine int: the routine id
//
// Returns:
//   - int: number of elements in the routine.
func (this *Trace) GetTraceLength(routine int) int {
	return this.routines[routine].Len()
}

// GetTraceLengths returns a slice containing the number of elements in the
// routines
//
// Returns:
//   - []int: number of elements in routines.
func (this *Trace) GetTraceLengths() []int {
	l := make([]int, this.GetNoRoutines()+1)
	for i, rout := range this.routines {
		l[i] = rout.Len()
	}

	return l
}

// NumberElemInRoutine returns the number of elements in the trace.
//
// Parameter:
//   - routine: return the number of elements in this routine, if -1, return the number of all elements
//
// Returns:
//   - int: the number of element in a routine or the complete trace
func (this *Trace) NumberElemInRoutine(routine int) int {
	if routine == -1 {
		total := 0
		for _, rout := range this.routines {
			total += rout.Len()
		}
		return total
	}

	if _, ok := this.routines[routine]; ok {
		return 0
	}

	return this.routines[routine].Len()
}

// SetHBWasCalc sets the hwWasCalc value of the trace
//
// Parameter:
//   - wasCalc bool: the new hbWasCalc value
func (this *Trace) SetHBWasCalc(wasCalc bool) {
	this.hbWasCalc = wasCalc
}

// GetHBWasCalc returns whether the hb clocks have been calculated
//
// Returns:
//   - bool: hbWasCalc
func (this *Trace) GetHBWasCalc() bool {
	return this.hbWasCalc
}

// GetNrAddDoneBeforeTime returns the number of add and done operations that were
// executed on a given wait group, before a given time.
//
// Parameter:
//   - wgID int: The id of the wait group
//   - waitTime int: The time to check
//
// Returns:
//   - int: The number of add operations
//   - int: The number of done operations
func (this *Trace) GetNrAddDoneBeforeTime(wgID int, waitTime int) (int, int) {
	nrAdd := 0
	nrDone := 0

	for _, routine := range this.routines {
		for _, elem := range routine.elems {
			switch e := elem.(type) {
			case *ElementWait:
				if e.ResourceID() == wgID {
					if e.T(Request) < waitTime {
						delta := e.GetDelta()
						if delta > 0 {
							nrAdd++
						} else if delta < 0 {
							nrDone++
						}
					}
				}
			}
		}
	}

	return nrAdd, nrDone
}

// PrintTrace prints the trace sorted by tPost
func (this *Trace) PrintTrace() {
	this.PrintTraceArgs([]string{}, false)
}

// PrintTraceArgs print the elements of given types sorted by tPost
//
// Parameter:
//   - ty: types of the elements to print. If empty, all elements will be printed
//   - clocks: if true, the clocks will be printed
func (this *Trace) PrintTraceArgs(ty []string, clocks bool) {
	elements := make([]struct {
		string
		time   int
		thread int
		vc     *a_clock.VectorClock
		wVc    *a_clock.VectorClock
	}, 0)
	for _, rout := range this.routines {
		for _, elem := range rout.elems {
			elemStr := elem.String()
			if len(ty) == 0 || types.Contains(ty, elemStr[0:1]) {
				elements = append(elements, struct {
					string
					time   int
					thread int
					vc     *a_clock.VectorClock
					wVc    *a_clock.VectorClock
				}{elemStr, elem.T(Commit), elem.RoutineID(), elem.GetVC(a_clock.Strong), elem.GetVC(a_clock.Weak)})
			}
		}
	}

	// sort elements by timestamp
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].time < elements[j].time
	})

	if len(elements) == 0 {
		log.Info("Trace contains no elements")
	} else {
		log.Infof("Trace contains %d elements", len(elements))
	}

	for _, elem := range elements {
		if clocks {
			fmt.Println(elem.thread, elem.string, elem.vc.String(), elem.wVc.String())
		} else {
			fmt.Println(elem.thread, elem.string)
		}
	}
}

// GetConcurrentWaitGroups returns all to element concurrent wait, broadcast
// and signal operations on the same condition variable
//
// Parameter:
//   - element traceElement: The element
//   - filter []string: The types of the elements to return
//
// Returns:
//   - []*traceElement: The concurrent elements
func (this *Trace) GetConcurrentWaitGroups(element Element) map[string][]Element {
	res := make(map[string][]Element)
	res["broadcast"] = make([]Element, 0)
	res["signal"] = make([]Element, 0)
	res["wait"] = make([]Element, 0)
	for _, rout := range this.routines {
		for _, elem := range rout.elems {
			switch elem.(type) {
			case *ElementCond:
			default:
				continue
			}

			if !element.IsSameElement(elem) {
				continue
			}

			e := elem.(*ElementCond)

			if a_clock.GetHappensBefore(element.GetVC(a_clock.Strong), e.GetVC(a_clock.Strong)) == a_hb.Concurrent {
				e := elem.(*ElementCond)
				switch e.op {
				case CondSignal:
					res["signal"] = append(res["signal"], elem)
				case CondBroadcast:
					res["broadcast"] = append(res["broadcast"], elem)
				case CondWait:
					res["wait"] = append(res["wait"], elem)
				}
			}
		}
	}
	return res
}

func (this *Trace) IsRoutTerm(id int) (bool, error) {
	if rout, ok := this.routines[id]; ok {
		return rout.IsTerminated(), nil
	}

	return false, fmt.Errorf("No Routine with ID %d", id)
}

// SetTSortAtIndex sets the tSort for an element given by its index
//
// Parameter:
//   - tSort int: the new tSort
//   - routine int: the routine of the element
//   - index int: the index of the element in its routine
func (this *Trace) SetTSortAtIndex(tPost, routine, index int) {
	this.routines[routine].SetTSortAtIndex(tPost, index)
}

func (this *Trace) CallTree() *CallTree {
	return &this.callTree
}

func (this *Trace) ForkTree() map[int][]int {
	return this.forkTree
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy returns a deep copy a trace
//
// Parameter:
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - Trace: The copy of the trace
//   - error
func (this *Trace) Copy(keep bool) (Trace, error) {
	mapping := make(map[int]Element)

	newTrace := NewTrace()

	for _, rout := range this.routines {
		newTrace.AddRoutine(rout.id)
		for _, elem := range rout.elems {
			newTrace.AddElement(elem.Copy(&newTrace, mapping, keep))
			if control.WasCanceled() {
				return Trace{}, fmt.Errorf("Analysis was canceled due to insufficient RAM")
			}
		}
	}

	return newTrace, nil
}

// ========================================================
// MARK: Sort
// ========================================================

// Helper functions to sort the trace by tSort
type sortByTSort []Element

// len function required for sorting
func (a sortByTSort) Len() int { return len(a) }

// swap function required for sorting
func (a sortByTSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// order function required for sorting
func (a sortByTSort) Less(i, j int) bool {
	return a[i].T(Sorting) < a[j].T(Sorting)
}

// Sort each routine of the trace by tPost
func (this *Trace) Sort() {
	for routine, _ := range this.routines {
		this.routines[routine].Sort()
	}
}

// ========================================================
// MARK: Modify Trace
// ========================================================

// ShortenTrace shortens the trace by removing all elements after the given time
//
// Parameter:
//   - time int: The time to shorten the trace to
//   - incl bool: True if an element with the same time should stay included in the trace
func (this *Trace) ShortenTrace(time int, incl bool) {
	for id, rout := range this.routines {
		for index, elem := range rout.elems {
			if incl && elem.T(Sorting) > time {
				this.routines[id].shortenIndex(index)
				break
			}
			if !incl && elem.T(Sorting) >= time {
				this.routines[id].shortenIndex(index)
				break
			}
		}
	}
}

// RemoveElementFromTrace removes the element with the given id from the trace
//
// Parameter:
//   - elem Element: element to remove
func (this *Trace) RemoveElementFromTrace(elem Element) {
	for id, rout := range this.routines {
		for index, e := range rout.elems {
			if e.ID() == elem.ID() {
				this.routines[id].removeAtIndex(index)
				break
			}
		}
	}
}

// ShortenRoutine shorten the trace of the given routine by removing all
// elements after and equal the given time
//
// Parameter:
//   - routine int: The routine to shorten
//   - time int: The time to shorten the trace to
func (this *Trace) ShortenRoutine(routine int, time int) {
	this.routines[routine].shortenTime(time)
}

// ShortenRoutineIndex shorten a given a routine to index
//
// Parameter:
//   - routine int: the routine to shorten
//   - index int: the index to which it should be shortened
//   - incl bool: if true, the value a index will remain in the routine, otherwise it will be removed
func (this *Trace) ShortenRoutineIndex(routine, index int, incl bool) {
	if incl {
		this.routines[routine].shortenIndex(index + 1)
	} else {
		this.routines[routine].shortenIndex(index)
	}
}

// ShiftTrace shifts all elements with time greater or equal to startTSort by shift
// Only shift forward
//
// Parameter:
//   - startTPre int: The time to start shifting
//   - shift int: The shift
func (this *Trace) ShiftTrace(startTPre int, shift int) bool {
	if shift <= 0 {
		return false
	}

	for _, routine := range this.routines {
		for index, elem := range routine.elems {
			if elem.T(Request) >= startTPre {
				routine.At(index).SetTWithoutNotExecuted(elem.T(Sorting) + shift)
			}
		}
	}

	return true
}

// ShiftConcurrentOrAfterToAfter shifts all elements that are concurrent or
// HB-later than the element such that they are after the element without
// changing the order of these elements
//
// Parameter:
//   - element traceElement: The element
func (this *Trace) ShiftConcurrentOrAfterToAfter(element Element) {
	elemsToShift := make([]Element, 0)
	minTime := -1

	for _, rout := range this.routines {
		for _, elem := range rout.elems {
			if elem.ID() == element.ID() {
				continue
			}

			if !(a_clock.GetHappensBefore(elem.GetVC(a_clock.Strong), element.GetVC(a_clock.Strong)) == a_hb.Before) {
				elemsToShift = append(elemsToShift, elem)
				if minTime == -1 || elem.T(Request) < minTime {
					minTime = elem.T(Request)
				}
			}
		}
	}

	distance := element.T(Request) - minTime + 1

	for _, elem := range elemsToShift {
		tSort := elem.T(Request)
		elem.SetT(Both, tSort+distance)
	}
}

// ShiftConcurrentOrAfterToAfterStartingFromElement shifts all elements that
// are concurrent or HB-later than the element such
// that they are after the element without changing the order of these elements
// Only shift elements that are after start
//
// Parameter:
//   - element traceElement: The element
//   - start traceElement: The time to start shifting (not including)
func (this *Trace) ShiftConcurrentOrAfterToAfterStartingFromElement(element Element, start int) {
	elemsToShift := make([]Element, 0)
	minTime := -1
	maxNotMoved := 0

	for _, rout := range this.routines {
		for _, elem := range rout.elems {
			if elem.ID() == element.ID() {
				continue
			}

			if !(a_clock.GetHappensBefore(elem.GetVC(a_clock.Strong), element.GetVC(a_clock.Strong)) == a_hb.Before) {
				if elem.T(Request) <= start {
					continue
				}

				elemsToShift = append(elemsToShift, elem)
				if minTime == -1 || elem.T(Request) < minTime {
					minTime = elem.T(Request)
				}
			} else {
				if maxNotMoved == 0 || elem.T(Request) > maxNotMoved {
					maxNotMoved = elem.T(Request)
				}
			}
		}
	}

	if !element.Committed() {
		element.SetT(Both, maxNotMoved+1)
	}

	distance := element.T(Request) - minTime + 1

	for _, elem := range elemsToShift {
		tSort := elem.T(Request)
		elem.SetT(Both, tSort+distance)
	}

}

// ShiftConcurrentToBefore shifts the element to be after all elements, that
// are concurrent to it
//
// Parameter:
//   - element traceElement: The element
func (this *Trace) ShiftConcurrentToBefore(element Element) {
	this.ShiftConcurrentOrAfterToAfterStartingFromElement(element, 0)
}

// RemoveConcurrent removes all elements that are concurrent to the element
// and have time greater or equal to tMin
//
// Parameter:
//   - element traceElement: The element
func (this *Trace) RemoveConcurrent(element Element, tMin int) {
	for routine, rout := range this.routines {
		result := make([]Element, 0)
		for _, elem := range rout.elems {
			if elem.T(Sorting) < tMin {
				result = append(result, elem)
				continue
			}

			if elem.ID() == element.ID() {
				result = append(result, elem)
				continue
			}

			if a_clock.GetHappensBefore(elem.GetVC(a_clock.Strong), element.GetVC(a_clock.Strong)) != a_hb.Concurrent {
				result = append(result, elem)
			}
		}
		this.routines[routine].elems = result
	}
}

// RemoveConcurrentOrAfter removes all elements that are concurrent to the
// element or must happen after the element
//
// Parameter:
//   - element traceElement: The element
func (this *Trace) RemoveConcurrentOrAfter(element Element, tMin int) {
	for routine, rout := range this.routines {
		result := make([]Element, 0)
		for _, elem := range rout.elems {
			if elem.T(Sorting) < tMin {
				result = append(result, elem)
				continue
			}

			if elem.ID() == element.ID() {
				result = append(result, elem)
				continue
			}

			if a_clock.GetHappensBefore(elem.GetVC(a_clock.Strong), element.GetVC(a_clock.Strong)) != a_hb.Before {
				result = append(result, elem)
			}
		}
		this.routines[routine].elems = result
	}
}

// GetConcurrentEarliest returns the earliest element that is concurrent to the element
//
// Parameter:
//   - element traceElement: The element
//
// Returns:
//   - map[int]traceElement: The earliest concurrent element for each routine
func (this *Trace) GetConcurrentEarliest(element Element) map[int]Element {
	concurrent := make(map[int]Element)
	for routine, rout := range this.routines {
		for _, elem := range rout.elems {
			if elem.ID() == element.ID() {
				continue
			}

			if a_clock.GetHappensBefore(element.GetVC(a_clock.Strong), elem.GetVC(a_clock.Strong)) == a_hb.Concurrent {
				concurrent[routine] = elem
			}
		}
	}
	return concurrent
}

// RemoveLater removes all elements that have a later tPost that the given tPost
//
// Parameter:
//   - tPost int: Remove elements after tPost
func (this *Trace) RemoveLater(tPost int) {
	mapping := make(map[int]Element)
	for routine, rout := range this.routines {
		newElems := make([]Element, 0)
		for _, elem := range rout.elems {
			if elem.T(Commit) > tPost {
				newElems = append(newElems, elem.Copy(this, mapping, true))
			}
		}
		this.routines[routine].elems = newElems
	}
}

// ShiftRoutine shifts all elements in a routine with time greater or equal to
// startTSort by shift. Only shift back (shift > 0).
//
// Parameter:
//   - routine int: The routine to shift
//   - startTSort int: The time to start shifting
//   - shift int: The shift, must be > 0
//
// Returns:
//   - bool: True if the shift was successful, false otherwise (shift <= 0)
func (this *Trace) ShiftRoutine(routine int, startTSort int, shift int) bool {
	if shift <= 0 {
		return false
	}

	this.routines[routine].shift(startTSort, shift)

	return true
}

// GetPartialTrace returns the partial trace of all element between startTime
// and endTime inclusive.
//
// Parameter:
//   - startTime int: The start time
//   - endTime int: The end time
//
// Returns:
//   - map[int][]TraceElement: The partial trace
func (this *Trace) GetPartialTrace(startTime int, endTime int) map[int][]Element {
	result := make(map[int][]Element)
	for routine, trace := range this.routines {
		for index, elem := range trace.elems {
			if _, ok := result[routine]; !ok {
				result[routine] = make([]Element, 0)
			}
			time := elem.T(Sorting)
			if time >= startTime && time <= endTime {
				result[routine] = append(result[routine], this.routines[routine].At(index))
			}
		}
	}

	return result
}

// split all blocking elements into request and commit
func (this *Trace) AsRequestCommit() int {
	this.request = true

	mapping := make(map[int]Element)

	newTr := make(map[int]*Routine)

	elemCounter := 0

	for rout, tr := range this.routines {
		newRout := Routine{id: tr.id, resources: tr.resources}
		for _, elem := range tr.elems {
			if elem.CanBeRequest() {
				newElem := elem.Copy(this, mapping, true)
				newElem.SetRequest(true)
				newRout.addElement(newElem)
				elemCounter++
			}
			if !elem.CanBeRequest() || elem.T(Commit) != 0 {
				newRout.addElement(elem)
				elemCounter++
			}
		}
		newTr[rout] = &newRout
	}

	this.routines = newTr

	return elemCounter
}

// Make the times consecutive. Only works if trace is request commit trace
func (this *Trace) NormalizeRequestCommit() {
	if !this.request {
		return
	}

	traceIter := this.AsIterator()
	i := 0

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		elem.SetT(Commit, i)
		i++
	}
}

// TODO: print in color
// TODO: print select with case
func (this *Trace) PrintAsTab() {
	colSize := "%-30.30s"

	trIter := this.AsIterator()
	numberRout := this.numberRoutines

	printCol := func(rout int, val string) string {
		res := "| "
		for i := 0; i < rout-1; i++ {
			res += fmt.Sprintf(colSize, "") + "| "
		}
		res += fmt.Sprintf(colSize, val)
		for i := rout - 1; i < numberRout; i++ {
			res += "| " + fmt.Sprintf(colSize, "")
		}

		return res
	}

	header := "| "

	for i := 0; i < numberRout; i++ {
		header += fmt.Sprintf(colSize, "G"+fmt.Sprintf("%d", i)) + "| "
	}
	header += "\n"
	for i := 0; i < numberRout+1; i++ {
		header += "| " + fmt.Sprintf(colSize, "")
	}
	println(header)

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		println(printCol(elem.RoutineID(), elem.StringMin()))
	}

}

// ========================================================
// MARK: Iterator
// ========================================================

// Iterator is an iterator to iterate over the element in the trace
// sorted by tSort
type Iterator struct {
	t            *Trace
	currentIndex map[int]int
}

// AsIterator returns a new iterator for a trace
//
// Returns:
//   - the iterator
func (this *Trace) AsIterator() Iterator {
	return Iterator{this, make(map[int]int)}
}

// GetTraceSection returns a copy of a section of the trace given by index
//
// Parameter:
//   - start int: start index
//   - end int: end index
//
// Returns:
//   - []trace.Element: the elements in the trace between start and end (including)
//     if start >= end, the result is empty. If start < 0, start is set to 0,
//     if end > len(trace), end is set to len(trace)
func (this *Trace) GetTraceSection(start, end int) []Element {
	if end <= start {
		return make([]Element, 0)
	}

	start = max(0, start)
	end = min(end, this.GetNumberElements()-1)
	numElems := end - start

	res := make([]Element, numElems)

	traceIter := this.AsIterator()

	counter := 0
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		if counter >= start {
			res = append(res, elem)
		}
		if counter >= end {
			return res
		}
		counter++
	}
	return res
}

// Next returns the next element from the iterator. If all elements have been returned
// already, return nul
//
// Returns:
//   - TraceElement: the next element, or nil if no element are left
func (this *Iterator) Next() Element {
	// find the local trace, where the element on which currentIndex points to
	// has the smallest tPost
	minTSort := -1
	minRoutine := -1
	for routine, rout := range this.t.routines {
		// no more elements in the routine trace
		if this.currentIndex[routine] == -1 {
			continue
		}

		// ignore empty routines
		if rout.Empty() {
			this.currentIndex[routine] = -1
			continue
		}

		// ignore non executed operations
		tSort := rout.At(this.currentIndex[routine]).T(Sorting)
		if tSort == 0 || tSort == math.MaxInt {
			continue
		}
		if minTSort == -1 || rout.At(this.currentIndex[routine]).T(Sorting) < minTSort {
			minTSort = rout.At(this.currentIndex[routine]).T(Sorting)
			minRoutine = routine
		}
	}

	// all executed elements have been processed
	// check for elements with just a pre but no post
	if minRoutine == -1 {
		for routine := range this.t.routines {
			if this.currentIndex[routine] == -1 {
				continue
			}

			element := this.t.routines[routine].At(this.currentIndex[routine])
			this.IncreaseIndex(routine)

			return element
		}

		// all elements have been processed
		return nil
	}

	// return the element and increase the index
	element := this.t.routines[minRoutine].At(this.currentIndex[minRoutine])
	this.IncreaseIndex(minRoutine)

	return element
}

// Reset resets the iterator
func (this *Iterator) Reset() {
	this.currentIndex = make(map[int]int)
}

// IncreaseIndex the currentIndex value of a trace iterator for a routine
//
// Parameter:
//   - routine int: the routine to update
func (this *Iterator) IncreaseIndex(routine int) {
	if this.currentIndex[routine] == -1 {
		log.Error("Tried to increase index of -1 at routine ", routine)
	}
	this.currentIndex[routine]++
	if this.currentIndex[routine] >= this.t.routines[routine].Len() {
		this.currentIndex[routine] = -1
	}
}

// ========================================================
// MARK: Resources
// ========================================================

func (this *Trace) GetResourcesPerRout(routID int) []Resource {
	return this.routines[routID].Resources()
}

func (this *Trace) GetResourcesRout() map[int][]Resource {
	res := make(map[int][]Resource)

	for id, rout := range this.routines {
		res[id] = rout.Resources()
	}

	return res
}

func (this *Trace) Resources() map[int]Resource {
	return this.resources
}

// GetAlloc returns the alloc of an element.
// For an alloc the element is returned.
// For elements without alloc, nil is returned
func (this *Trace) GetResources(elem Element) []Resource {
	res := make([]Resource, 0)

	switch elem := elem.(type) {
	case *ElementFork, *ElementFunc, *ElementReturn, *ElementRoutineEnd:
	case *ElementSelect:
		for _, c := range elem.GetCases() {
			r, ok := this.resources[c.ResourceID()]
			if !ok || r.id == 0 {
				panic("Invalid Resource")
			}
			res = append(res, r)
		}
	default:
		r, ok := this.resources[elem.ResourceID()]
		if !ok || r.id == 0 {
			panic("Invalid Resource")
		}
		res = append(res, r)
	}

	return res
}

// ========================================================
// MARK: CallTree
// ========================================================

type CallTree struct {
	tree map[*ElementFunc][]*ElementFunc
	root *ElementFunc
}

func newCallGraph() *CallTree {
	return &CallTree{
		tree: make(map[*ElementFunc][]*ElementFunc),
		root: nil,
	}
}

func (this *CallTree) AddElem(parent, child *ElementFunc) {
	if parent != nil {
		this.tree[parent] = append(this.tree[parent], child)
	} else {
		this.root = child
	}

	this.tree[child] = make([]*ElementFunc, 0)
}

func (this *CallTree) GetTree() map[*ElementFunc][]*ElementFunc {
	return this.tree
}

func (this *CallTree) String() string {
	if this == nil || this.root == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintln(&b, "CALL TREE")

	var walk func(*ElementFunc, int)
	walk = func(fn *ElementFunc, depth int) {
		b.WriteString(strings.Repeat("    ", depth))
		fmt.Fprintln(&b, fn.GetSSAName())

		for _, child := range this.tree[fn] {
			walk(child, depth+1)
		}
	}

	walk(this.root, 0)
	return b.String()
}
