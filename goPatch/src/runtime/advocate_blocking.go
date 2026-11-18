// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_partial_deadlock.go
// Brief: Detect partial deadlocks while running
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

var CollectPartialDeadlockInfo = false
var InfoHaveRef map[uintptr][]bool // pointer to parked operation -> list of routines with reference to this
var haveRefArray []uintptr

// build oat creates the object aware trace.
// It first determines the blocked concurrency objects.
// It then determines which non-terminated routines have access to those objects
func BuildOAT() {
	blocked, maxRoutId := blockeInfo()

	// from blocked obj memory address to slice.
	// Each elem in slice represents one routine, id -> index
	// If true, routine has access, otherwise it does not
	// Done to prevent alloc during GC
	InfoHaveRef = make(map[uintptr][]bool)
	for b := range blocked {
		InfoHaveRef[b] = make([]bool, 2*(maxRoutId + 1))
	}

	haveRefArray = make([]uintptr, len(InfoHaveRef))
	i := 0
	for b := range InfoHaveRef {
		haveRefArray[i] = b
		i++
	}

	CollectPartialDeadlockInfo = true
	GC()
	CollectPartialDeadlockInfo = false

	// invert InfoHaveRef, rout id -> []object id
	ref := make(map[uint64][]uint64)
	for obj, routInfo := range InfoHaveRef {
		for rout, ok := range routInfo {
			if !ok {
				continue
			}

			r := uint64(rout)

			if _, ok := ref[r]; !ok {
				ref[r] = make([]uint64, 0)
			}

			ref[r] = append(ref[r], blocked[obj])
		}
	}

	ForEachAdvocateG(func(adGp *AdvocateG) {
		if adGp.isNil() {
			return
		}

		id := adGp.id()

		if obj, ok := ref[id]; ok {
			adGp.setOAT(obj)
		}
	})
}

// blockeInfo determines all blocked resource
// It additionally determines the max routine id
//
// Returns:
//   - map[uintptr]park: blocked resource, from object memory addres to object info
//   - uint64: max rout id
func blockeInfo() (map[uintptr]uint64, uint64) {
	blocked := make(map[uintptr]uint64, 0)
	var maxID uint64 = 0

	ForEachAdvocateG(func(gp *AdvocateG) {
		if gp.isNil() {
			return
		}

		le := gp.LastTraceElem()

		if le == nil {
			return
		}

		if !le.hasCommit() { // automatically filters out terminated routines
			for _, res := range le.resource() {
				blocked[uintptr(res.addr)] = res.id
			}
		}

		maxID = max(maxID, gp.id())
	})

	return blocked, maxID
}
