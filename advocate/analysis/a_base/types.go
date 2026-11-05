// advocate/analysis/baseA/types.go

// Copyright (c) 2025 Erik Kassubek
//
// File: types.go
// Brief: Types used in analysis data
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_base

import (
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
	"strconv"
	"strings"
)

// ElemWithVc is a helper element for an element with an additional vector clock
type ElemWithVc struct {
	Vc   *a_clock.VectorClock
	Elem trace.Element
}

// VectorClockTID2 is a helper to store the relevant elements of a
// trace element without needing to store the element itself
type VectorClockTID2 struct { // TODO: replace
	Routine  int
	ID       int
	Vc       *a_clock.VectorClock
	TypeVal  trace.OperationType
	Val      int
	Buffered bool
	Sel      bool
	SelID    int
}

// ElemWithVcVal is a helper element for an element with an additional vector clock
// and an additional int val
type ElemWithVcVal struct {
	Elem trace.Element
	Vc   *a_clock.VectorClock
	Val  int
}

// AllSelectCase is a helper element to store individual references to all
// select cases in a trace
type AllSelectCase struct {
	Sel          *trace.ElementSelect // the select
	ChanID       int                  // channel id
	Elem         ElemWithVc           // vector clock and tID
	Send         bool                 // true: send, false: receive
	Buffered     bool                 // true: buffered, false: unbuffered
	PartnerFound bool                 // true: partner found, false: no partner found
	Partner      []ElemWithVcVal      // the potential partner
	Exec         bool                 // true: the case was executed, false: otherwise
	Casi         int                  // internal index for the case in the select
}

// ConcurrentEntryType is an enum type used in ConcurrentEntry
type ConcurrentEntryType int

// ConcurrentEntry is a helper element to store elements relevant for
// flow fuzzing
type ConcurrentEntry struct {
	Elem    trace.Element
	Counter int
	Type    ConcurrentEntryType
}

// Possible values for ConcurrentEntryType
const (
	CEOnce ConcurrentEntryType = iota
	CEMutex
	CESend
	CERecv
)

// BufferedVC are elements for buffered channel internal vector clock
type BufferedVC struct {
	Occupied bool
	Send     *trace.ElementChannel
}

// HoldObj can temporarily hold an channel operations with additional information
// it is used in the case that for a synchronous communication, the recv is
// recorded before the send
type HoldObj struct {
	Ch  *trace.ElementChannel
	Vc  map[int]*a_clock.VectorClock
	WVc map[int]*a_clock.VectorClock
}

// State for resource deadlock detection
type State struct {
	Routines map[int]Thread // Recording lock dependencies in phase 1
	Cycles   []Cycle        // Computing cycles in phase 2
	Failed   bool           // Analysis failed (encountered unsupported lock action)
}

// LockDependency represents a Lock dependencies for resource deadlock detection
type LockDependency struct {
	Thread   int
	Lock     LockID
	Lockset  Lockset
	Requests []trace.Element
}

// Cycle is a set of mutex operations, that could constitute a resource deadlock
type Cycle []LockDependency

// Thread represents a routine
// Lock dependencies are computed thread-local. We make use of the following structures.
type Thread struct {
	LockDependencies map[LockID][]Dependency
	CurrentLockset   Lockset        // The thread's current lockset.
	ReaderCounter    map[LockID]int // Store how many readers a readlock has
}

// Dependency represents a set of lock dependencies
// Unfortunately, we can't use double-indexed map of the following form in Go.
// type Deps map[Lock]map[Lockset][]Event
// Hence, we introduce some intermediate structure.
type Dependency struct {
	Lockset  Lockset
	Requests []trace.Element
}

// Representation of vector clocks, events, threads, lock and lockset.

// LockEvent represents a lock operations
// type LockEvent struct {
// 	ThreadID    ThreadID
// 	TraceID     int
// 	LockID      int
// 	VectorClock *a_clock.VectorClock
// }

// LockID represents a lock operation
type LockID struct {
	ID       int
	ReadLock bool
}

// Lockset implements a lockset
type Lockset map[LockID]trace.Element

func (this Lockset) Elements() []trace.Element {
	values := make([]trace.Element, 0, len(this))

	for _, value := range this {
		values = append(values, value)
	}

	return values
}

// Lock Dependency methods.

// Clone creates a copy of a lock dependency
//
// Returns:
//   - LockDependency: The copy
func (this LockDependency) Clone() LockDependency {
	reqs := make([]trace.Element, len(this.Requests))
	copy(reqs, this.Requests)
	return LockDependency{
		Thread:   this.Thread,
		Lock:     this.Lock,
		Lockset:  this.Lockset.Clone(),
		Requests: reqs,
	}
}

// Lock methods.

// IsRead checks if the lock is a reader lock
//
// Returns:
//   - bool: true if it is a reader lock, false otherwise
func (this LockID) IsRead() bool {
	return this.ReadLock
}

// IsWrite checks if the lock is not a reader lock
//
// Returns:
//   - bool: false if it is a reader lock, true otherwise
func (this LockID) IsWrite() bool {
	return !this.ReadLock
}

// AddReader increases the reader counter of a thread for the lock id
//
// Parameter:
//   - s Thread: the thread to increase the ReaderCounter for
func (this LockID) AddReader(s Thread) {
	s.ReaderCounter[this]++
}

// RemoveReader decreases the reader counter of a thread for the lock id
// if it has readers at a given thread
// If the counter gets zero, the reader counter is removed
//
// Parameter:
//   - s Thread: the thread to increase the ReaderCounter for
func (this LockID) RemoveReader(s Thread) {
	if !this.HasReaders(s) {
		return
	}
	s.ReaderCounter[this]--
	if s.ReaderCounter[this] <= 0 {
		delete(s.ReaderCounter, this)
	}
}

// HasReaders checks if a lock has reader given a thread
//
// Parameter:
//   - s Thread: the thread
//
// Returns:
//   - bool: true if it has reader, false otherwise
func (this LockID) HasReaders(s Thread) bool {
	if _, exists := s.ReaderCounter[this]; !exists {
		return false
	}
	return s.ReaderCounter[this] > 0
}

// EqualsIgnoreRW checks if two locks are equal ignoring whether they are read or write locks.
//
// Parameter:
//   - other LockID: the other lock
//
// Returns:
//   - bool: true if two locks are equal ignoring whether they are read or write locks, false otherwise
func (this LockID) EqualsIgnoreRW(other LockID) bool {
	return this.ID == other.ID
}

// EqualsCouldBlock checks if two locks are the same and at least one of them is a write lock.
//
// Parameter:
//   - other LockID: the other lock
//
// Returns:
//   - true if l and other are the same and at least one of them is a write lock.
func (this LockID) EqualsCouldBlock(other LockID) bool {
	if !this.EqualsIgnoreRW(other) {
		return false
	}
	return this.IsWrite() || other.IsWrite()
}

// Lockset methods.

// Empty returns if a lockset is empty
//
// Returns:
//   - bool: if ls is empty
func (this Lockset) Empty() bool {
	return len(this) == 0

}

// Add adds a lockID to the lockset
//
// Parameter:
//   - x LockID: the lock id to add
func (this Lockset) Add(x LockID, event trace.Element) {
	this[x] = event
}

// Remove removes a lock id from a lockset
//
// Parameter:
//   - x LockID: the id to remove
//
// Returns:
//   - bool: true if x was in the lockset, false otherwise
func (this Lockset) Remove(x LockID) bool {
	if _, contains := this[x]; !contains {
		return false
	}
	delete(this, x)
	return true
}

// Clone creates a copy of a lockset
//
// Returns:
//   - Lockset: the copy
func (this Lockset) Clone() Lockset {
	clone := make(Lockset, 0)
	for l := range this {
		clone[l] = this[l]
	}
	return clone
}

// String returns a string representation of a lockset
//
// Returns:
//   - string: the string representation of ls
func (this Lockset) String() string {
	b := strings.Builder{}
	b.WriteString("Lockset{")
	for l := range this {
		b.WriteString(strconv.Itoa(int(l.ID)))
	}
	b.WriteString("}")
	return b.String()
}

// Equal checks if the lockset is equal to another lockset
//
// Parameter:
//   - ls2 Lockset: the lockset to compare to
//
// Returns:
//   - bool: if ls and ls2 contain the same elements
func (this Lockset) Equal(ls2 Lockset) bool {
	if len(this) != len(ls2) {
		return false
	}

	for l := range this {
		if _, contains := ls2[l]; !contains {
			return false
		}
	}
	return true
}

// Disjoint checks if no element is in both ls and ls2
//
// Parameter:
//   - ls2 Lockset: the second lockset
//
// Returns:
//   - bool: true if ls and ls2 are disjoint, false otherwise
func (this Lockset) Disjoint(ls2 Lockset) bool {
	for l := range this {
		if _, contains := ls2[l]; contains {
			return false
		}
	}
	return true
}

// DisjointCouldBlock checks for no pair of elements in ls and ls2
// if they are the same and at least one of them is a write lock
//
// Parameter:
//   - ls2 Lockset: the other lockset
//
// Returns:
//   - bool: false if there is a pair of elements in ls and ls2 such that they
//     are the same and at least one of them is a write lock, true otherwise
func (this Lockset) DisjointCouldBlock(ls2 Lockset) bool {
	for l := range this {
		for l2 := range ls2 {
			if l.EqualsCouldBlock(l2) {
				return false
			}
		}
	}
	return true
}
