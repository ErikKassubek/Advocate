// Copyright (c) 2026 
//
// File: gosc_constraints.go
// Brief: Enforce constraints for replay
//
//
// License: BSD-3-Clause

package runtime

var (
	// constraints for channel
	constraintsChanW = make(map[string]map[*constraint]struct{})
	constraintsChanR = make(map[string]map[*constraint]struct{})
	snd              = make(map[int]chan struct{})
	rcv              = make(map[int]chan struct{})
)

type constraintMode int

const (
	base constraintMode = iota
	rcvMode
	sndMode
)

// struct to store a constraint w -> r or w -/-> r
//
// Fields:
//   - w string: pos of write
//   - e string: pos of read
//   - pos bool: true for positive constraints, false for negative
type constraint struct {
	w        string
	r        string
	pos      bool
	mode     constraintMode
	waitList []chan struct{}
}

func AddConstraint(pos bool, t, w, r string) {
	// needed to stop c from escaping to the heap, which is not allowed in the runtime
	c := new(constraint)

	*c = constraint{
		w:        w,
		r:        r,
		pos:      pos,
		mode:     base,
		waitList: make([]chan struct{}, 0),
	}

	if t == "C" { // channel
		if _, ok := constraintsChanR[r]; !ok {
			constraintsChanR[r] = make(map[*constraint]struct{})
		}
		if _, ok := constraintsChanW[w]; !ok {
			constraintsChanW[w] = make(map[*constraint]struct{})
		}
		constraintsChanR[r][c] = struct{}{}
		constraintsChanW[w][c] = struct{}{}
	}
}

// TODO: when are elements in waiting list released?
func ExecuteConstraint(id int, write bool, skip int) chan struct{} {
	chWait := make(chan struct{}, 1)

	pos := posFromCaller(skip)

	done := make([]*constraint, 0)

	if write {
		for ccw := range constraintsChanW[pos] {
			switch ccw.mode {
			case base:
				if ccw.w != pos {
					ccw.waitList = append(ccw.waitList, chWait)
				} else {
					snd[id] = chWait
					ccw.mode = rcvMode
				}
			case rcvMode:
				ccw.waitList = append(ccw.waitList, chWait)
			case sndMode:
				if ccw.w == pos {
					rcv[id] <- struct{}{}
					chWait <- struct{}{}
					done = append(done, ccw)
				} else {
					ccw.waitList = append(ccw.waitList, chWait)
				}
			}
		}
	} else {
		for ccr := range constraintsChanR[pos] {
			switch ccr.mode {
			case base:
				if ccr.r != pos {
					ccr.waitList = append(ccr.waitList, chWait)
				} else {
					rcv[id] = chWait
					ccr.mode = sndMode
				}
			case rcvMode:
				if ccr.r == pos {
					snd[id] <- struct{}{}
					chWait <- struct{}{}
					done = append(done, ccr)
				} else {
					ccr.waitList = append(ccr.waitList, chWait)
				}
			case sndMode:
				ccr.waitList = append(ccr.waitList, chWait)
			}
		}
	}

	// delete finished constraints
	for _, elem := range done {
		delete(constraintsChanR[elem.r], elem)
		delete(constraintsChanW[elem.w], elem)

		if len(constraintsChanR[elem.r]) == 0 {
			delete(constraintsChanR, elem.r)
		}
		if len(constraintsChanW[elem.w]) == 0 {
			delete(constraintsChanW, elem.w)
		}
	}

	return chWait

}
