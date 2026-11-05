// advocate/analysis/analysis/scenarios/mixedDeadlock.go

// Copyright (c) 2024 Erik Kassubek
//
// File: mixedDeadlock.go
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/results/buginfo"
	"advocate/utils/results/results"
	"advocate/utils/timer"
	"fmt"
)

// ---------------------------------------------------------------------------
// Data structures
// ---------------------------------------------------------------------------

// mdLockRef associates a lock ID and its CS/PCS status with the concrete RD
type mdLockRef struct {
	LockID a_base.LockID
	IsCS   bool
	RD     *mdRDNode
}

// mdRDNode represents a single lock-acquire event in a goroutine
type mdRDNode struct {
	Thread   int
	Lock     a_base.LockID
	Requests []*trace.ElementMutex
	Elem     *trace.ElementMutex
}

// mdCDNode represents a channel operation that has at least one lock context
type mdCDNode struct {
	Thread     int
	ChanID     int
	OpType     trace.OperationType // ChannelSend | ChannelRecv | ChannelClose
	Buffered   bool
	AssocRDs   []mdLockRef           // CS and PCS lock contexts at time of op
	Elem       *trace.ElementChannel // concrete trace element
	Depth      int                   // total lock stack depth at channel op
	ReadDepth  int                   // read lock depth at channel op (for RWMutex)
	WriteDepth int                   // write lock depth at channel op
}

// mdThreadState holds per-goroutine online recording state
type mdThreadState struct {
	CurrentLockset a_base.Lockset
	ActiveRDs      map[a_base.LockID][]*mdRDNode // Stack for nested locks
	MostRecentRD   map[a_base.LockID]*mdRDNode
	ReadLockCount  map[a_base.LockID]int
	LockDepth      int // Total locks currently held
	ReadDepth      int // Read locks currently held (for RWMutex)
	WriteDepth     int // Write locks currently held
}

// mdState as global analysis state for mixed-deadlock detection
type mdState struct {
	Threads map[int]*mdThreadState
	AllCDs  []*mdCDNode
}

var currentMDState mdState

// ---------------------------------------------------------------------------
// Phase 1: Online event recording
// ---------------------------------------------------------------------------

// ResetMixedDeadlockState resets all state before a new analysis run.
func ResetMixedDeadlockState() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)
	currentMDState = mdState{
		Threads: make(map[int]*mdThreadState),
	}
}

// getOrCreateMDThread returns the per-goroutine state, creating it if absent.
func getOrCreateMDThread(tid int) *mdThreadState {
	if t, ok := currentMDState.Threads[tid]; ok {
		return t
	}
	t := &mdThreadState{
		CurrentLockset: make(a_base.Lockset),
		ActiveRDs:      make(map[a_base.LockID][]*mdRDNode),
		MostRecentRD:   make(map[a_base.LockID]*mdRDNode),
		ReadLockCount:  make(map[a_base.LockID]int),
		LockDepth:      0,
		ReadDepth:      0,
		WriteDepth:     0,
	}
	currentMDState.Threads[tid] = t
	return t
}

// HandleMutexEventForMixedDeadlock processes one mutex trace event
func HandleMutexEventForMixedDeadlock(element *trace.ElementMutex) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	tid := element.RoutineID()
	t := getOrCreateMDThread(tid)

	isReadLock := false
	isReadUnlock := false
	switch element.Type(true) {
	case trace.MutexRLock, trace.MutexTryRLock:
		isReadLock = true
	case trace.MutexRUnlock:
		isReadUnlock = true
	}

	lockID := a_base.LockID{ID: element.ResourceID(), ReadLock: isReadLock || isReadUnlock}

	switch element.Type(true) {

	// --------- WRITE LOCK ---------
	case trace.MutexLock, trace.MutexTryLock:
		mdPushRD(t, tid, lockID, element, element)
		t.CurrentLockset.Add(lockID, element)
		t.LockDepth++
		t.WriteDepth++

	// --------- READ LOCK (RWMutex RLock) ---------
	case trace.MutexRLock, trace.MutexTryRLock:
		if _, ok := t.ReadLockCount[lockID]; !ok {
			t.ReadLockCount[lockID] = 0
		}
		t.ReadLockCount[lockID]++

		if t.ReadLockCount[lockID] == 1 {
			mdPushRD(t, tid, lockID, element, element)
			t.CurrentLockset.Add(lockID, element)
			t.LockDepth++
		}
		t.ReadDepth++

	// --------- WRITE UNLOCK ---------
	case trace.MutexUnlock:
		if stack, ok := t.ActiveRDs[lockID]; ok && len(stack) > 0 {
			rd := stack[len(stack)-1]
			t.ActiveRDs[lockID] = stack[:len(stack)-1]
			t.MostRecentRD[lockID] = rd
			if len(t.ActiveRDs[lockID]) == 0 {
				delete(t.ActiveRDs, lockID)
			}
			t.LockDepth--
			t.WriteDepth--
		}
		t.CurrentLockset.Remove(lockID)

	// --------- READ UNLOCK (RWMutex RUnlock) ---------
	case trace.MutexRUnlock:
		if _, ok := t.ReadLockCount[lockID]; !ok {
			break
		}

		t.ReadLockCount[lockID]--
		t.ReadDepth--

		if t.ReadLockCount[lockID] == 0 {
			if stack, ok := t.ActiveRDs[lockID]; ok && len(stack) > 0 {
				rd := stack[len(stack)-1]
				t.ActiveRDs[lockID] = stack[:len(stack)-1]
				t.MostRecentRD[lockID] = rd
				if len(t.ActiveRDs[lockID]) == 0 {
					delete(t.ActiveRDs, lockID)
				}
				t.LockDepth--
			}
			t.CurrentLockset.Remove(lockID)
			delete(t.ReadLockCount, lockID)
		}

	default:
		log.Error(fmt.Sprintf("MD phase1: unknown mutex operation: %s", element.String()))
	}
}

// mdPushRD pushes a new RD node onto the stack for the given lock
func mdPushRD(
	t *mdThreadState,
	tid int,
	lockID a_base.LockID,
	event *trace.ElementMutex,
	element *trace.ElementMutex,
) {
	rd := &mdRDNode{
		Thread:   tid,
		Lock:     lockID,
		Requests: []*trace.ElementMutex{event},
		Elem:     element,
	}
	t.ActiveRDs[lockID] = append(t.ActiveRDs[lockID], rd)
}

// HandleChannelEventForMixedDeadlock processes one channel trace event
func HandleChannelEventForMixedDeadlock(element *trace.ElementChannel) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	opType := element.Type(true)
	switch opType {
	case trace.ChannelSend, trace.ChannelRecv, trace.ChannelClose:
	default:
		return
	}

	tid := element.RoutineID()
	t := getOrCreateMDThread(tid)

	var assocRDs []mdLockRef

	// CS locks: currently held locks at time of channel op
	for lockID := range t.CurrentLockset {
		if stack, ok := t.ActiveRDs[lockID]; ok && len(stack) > 0 {
			topRD := stack[len(stack)-1]
			fmt.Printf("DEBUG: CS lock for T%d chan: lock=%d, tPre=%d\n",
				tid, lockID.ID, topRD.Elem.T(trace.Request))
			assocRDs = append(assocRDs, mdLockRef{LockID: lockID, IsCS: true, RD: topRD})
		}
	}

	// PCS locks: released before the channel op
	for lockID, rd := range t.MostRecentRD {
		if _, held := t.CurrentLockset[lockID]; held {
			continue
		}
		assocRDs = append(assocRDs, mdLockRef{LockID: lockID, IsCS: false, RD: rd})
	}

	if len(assocRDs) == 0 {
		return
	}

	cd := &mdCDNode{
		Thread:     tid,
		ChanID:     element.ResourceID(),
		OpType:     opType,
		Buffered:   element.IsBuffered(),
		AssocRDs:   assocRDs,
		Elem:       element,
		Depth:      t.LockDepth,
		ReadDepth:  t.ReadDepth,
		WriteDepth: t.WriteDepth,
	}
	currentMDState.AllCDs = append(currentMDState.AllCDs, cd)
}

// ---------------------------------------------------------------------------
// Phase 2: Offline CD-CD partner matching
// ---------------------------------------------------------------------------

// CheckForMixedDeadlock as entry point
func CheckForMixedDeadlock() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	cdByChan := make(map[int][]*mdCDNode, len(currentMDState.AllCDs))
	for _, cd := range currentMDState.AllCDs {
		cdByChan[cd.ChanID] = append(cdByChan[cd.ChanID], cd)
	}

	reported := make(map[[2]*trace.ElementChannel]bool)

	for _, cd := range currentMDState.AllCDs {
		switch cd.OpType {

		case trace.ChannelRecv:
			partner := mdFindPartnerForRecv(cd, cdByChan)
			if partner != nil {
				mdCheckAndReport(cd, partner, reported)
			}
			closePartner := mdFindCloseCDNode(cd, cdByChan)
			if closePartner != nil {
				mdCheckAndReport(cd, closePartner, reported)
			}

		case trace.ChannelSend:
			if cd.Buffered {
				continue
			}
			partner := mdFindPartnerCDByElem(cd.Elem.GetPartner(), cdByChan[cd.ChanID])
			if partner == nil {
				continue
			}
			mdCheckAndReport(cd, partner, reported)

		default:
		}
	}
}

// ---------------------------------------------------------------------------
// Partner lookup helpers
// ---------------------------------------------------------------------------

func mdFindPartnerForRecv(recvCD *mdCDNode, cdByChan map[int][]*mdCDNode) *mdCDNode {
	elem := recvCD.Elem
	if elem.GetClosed() {
		return mdFindCloseCDNode(recvCD, cdByChan)
	}
	return mdFindPartnerCDByElem(elem.GetPartner(), cdByChan[recvCD.ChanID])
}

func mdFindPartnerCDByElem(partnerElem *trace.ElementChannel, candidates []*mdCDNode) *mdCDNode {
	if partnerElem == nil {
		return nil
	}
	for _, c := range candidates {
		if c.Elem == partnerElem {
			return c
		}
	}
	return nil
}

func mdFindCloseCDNode(recvCD *mdCDNode, cdByChan map[int][]*mdCDNode) *mdCDNode {
	for _, c := range cdByChan[recvCD.ChanID] {
		if c.OpType != trace.ChannelClose {
			continue
		}
		if c.Thread == recvCD.Thread {
			continue
		}
		return c
	}
	return nil
}

// ---------------------------------------------------------------------------
// MD condition checks and reporting
// ---------------------------------------------------------------------------

func mdCheckAndReport(cdA, cdB *mdCDNode, reported map[[2]*trace.ElementChannel]bool) {
	key := mdPairKey(cdA.Elem, cdB.Elem)
	if reported[key] {
		return
	}

	for _, refA := range cdA.AssocRDs {
		for _, refB := range cdB.AssocRDs {

			if !refA.LockID.EqualsCouldBlock(refB.LockID) {
				continue
			}

			if !refA.IsCS && !refB.IsCS {
				continue
			}

			if refA.IsCS && refB.IsCS && !cdA.Buffered {
				if cdA.OpType != trace.ChannelClose && cdB.OpType != trace.ChannelClose {
					continue
				}
			}

			if !mdLockAcqAreConcurrent(refA.RD, refB.RD) {
				continue
			}

			holderCD, holderRef, waiterCD, waiterRef := mdDetermineRoles(cdA, refA, cdB, refB)

			mdReportCandidate(holderCD, holderRef, waiterCD, waiterRef)
			reported[key] = true
			return
		}
	}
}

func mdLockAcqAreConcurrent(rdA, rdB *mdRDNode) bool {
	if rdA == nil || rdB == nil {
		return false
	}
	for _, reqA := range rdA.Requests {
		if reqA.VC(a_clock.Strong) == nil {
			continue
		}
		for _, reqB := range rdB.Requests {
			if reqB.VC(a_clock.Strong) == nil {
				continue
			}
			if a_clock.GetHappensBefore(reqA.VC(a_clock.Strong), reqB.VC(a_clock.Strong)) == a_hb.Concurrent {
				return true
			}
		}
	}
	return false
}

// mdDetermineRoles assigns holder and waiter roles
func mdDetermineRoles(
	cdA *mdCDNode, refA mdLockRef,
	cdB *mdCDNode, refB mdLockRef,
) (holderCD *mdCDNode, holderRef mdLockRef, waiterCD *mdCDNode, waiterRef mdLockRef) {

	// Asymmetric: CS side is the holder
	if refA.IsCS && !refB.IsCS {
		return cdA, refA, cdB, refB
	}
	if refB.IsCS && !refA.IsCS {
		return cdB, refB, cdA, refA
	}

	// Both CS (MD2-1)
	if refA.IsCS && refB.IsCS {
		// For RWMutex: reader with read lock should be holder
		if cdA.ReadDepth > 0 && cdB.ReadDepth == 0 {
			return cdA, refA, cdB, refB
		}
		if cdB.ReadDepth > 0 && cdA.ReadDepth == 0 {
			return cdB, refB, cdA, refA
		}
		// Otherwise use total depth (deeper stack = inner CS = holder)
		if cdA.Depth > cdB.Depth {
			return cdA, refA, cdB, refB
		}
		if cdB.Depth > cdA.Depth {
			return cdB, refB, cdA, refA
		}
	}

	// Both CS with same depth VC order
	if len(refA.RD.Requests) > 0 && len(refB.RD.Requests) > 0 {
		vcA := refA.RD.Requests[0].VC(a_clock.Strong)
		vcB := refB.RD.Requests[0].VC(a_clock.Strong)
		if vcA != nil && vcB != nil {
			switch a_clock.GetHappensBefore(vcA, vcB) {
			case a_hb.Before:
				return cdB, refB, cdA, refA
			case a_hb.After:
				return cdA, refA, cdB, refB
			}
		}
	}

	// Tie-break: larger tPre = acquired later = holder
	tPreA, tPreB := 0, 0
	if refA.RD.Elem != nil {
		tPreA = refA.RD.Elem.T(trace.Request)
	}
	if refB.RD.Elem != nil {
		tPreB = refB.RD.Elem.T(trace.Request)
	}
	if tPreA >= tPreB {
		return cdA, refA, cdB, refB
	}
	return cdB, refB, cdA, refA
}

// ---------------------------------------------------------------------------
// Reporting
// ---------------------------------------------------------------------------

func mdReportCandidate(
	holderCD *mdCDNode, holderRef mdLockRef,
	waiterCD *mdCDNode, waiterRef mdLockRef,
) {
	if holderCD.Elem == nil || holderRef.RD == nil || holderRef.RD.Elem == nil ||
		waiterCD.Elem == nil || waiterRef.RD == nil || waiterRef.RD.Elem == nil {
		log.Error("MD report: nil element pointer in candidate — skipping")
		return
	}

	// holder's channel element
	holderChanRes := results.TraceElementResult{
		RoutineID:  int(holderCD.Thread),
		ResourceID: holderCD.ChanID,
		TRequest:   holderCD.Elem.T(trace.Request),
		ObjType:    holderCD.OpType,
		File:       holderCD.Elem.File(),
		Line:       holderCD.Elem.Line(),
	}

	// holder's lock acquire element
	holderLockRes := results.TraceElementResult{
		RoutineID:  int(holderRef.RD.Thread),
		ResourceID: holderRef.LockID.ID,
		TRequest:   holderCD.Elem.T(trace.Request),
		ObjType:    "DC",
		File:       holderCD.Elem.File(),
		Line:       holderCD.Elem.Line(),
	}

	// waiter's channel element
	waiterChanRes := results.TraceElementResult{
		RoutineID:  int(waiterCD.Thread),
		ResourceID: waiterCD.ChanID,
		TRequest:   waiterCD.Elem.T(trace.Request),
		ObjType:    waiterCD.OpType,
		File:       waiterCD.Elem.File(),
		Line:       waiterCD.Elem.Line(),
	}

	// waiter's lock acquire element (stuck element)
	waiterLockReq := waiterRef.RD.Requests[0]
	waiterLockRes := results.TraceElementResult{
		RoutineID:  int(waiterRef.RD.Thread),
		ResourceID: waiterRef.LockID.ID,
		TRequest:   waiterLockReq.T(trace.Request),
		ObjType:    "DC",
		File:       waiterLockReq.File(),
		Line:       waiterLockReq.Line(),
	}

	stuckElement := waiterLockRes
	stuckElement.ObjType = "DH"

	cycleElements := []results.ResultElem{
		holderChanRes,
		holderLockRes,
		waiterChanRes,
		waiterLockRes,
	}

	results.Result(
		results.CRITICAL,
		buginfo.PMixedDeadlock,
		[]results.ResultElem{stuckElement}, cycleElements,
	)
}

// mdPairKey returns a canonical key for a channel-element pair (de-duplication)
func mdPairKey(a, b *trace.ElementChannel) [2]*trace.ElementChannel {
	if a.T(trace.Request) <= b.T(trace.Request) {
		return [2]*trace.ElementChannel{a, b}
	}
	return [2]*trace.ElementChannel{b, a}
}
