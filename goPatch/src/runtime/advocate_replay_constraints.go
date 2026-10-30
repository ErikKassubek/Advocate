// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_replay_constrains.go
// Brief: Functionality to handle the constraints
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

const constraintElemNoInfo = -100

type ConstraintElem struct {
	key   string
	index int
	info  int
	chAck chan struct{}
}

func NewConstraintElem(routine int, file string, line, index, info int) *ConstraintElem {
	return &ConstraintElem{
		key:   intToString(routine) + "$" + file + "#" + intToString(line),
		index: index,
		info:  info,
	}
}

func AddConstraintElemFromString(info string) {
	fields := split(info, ':')
	if len(fields) != 4 {
		return
	}

	newConst := ConstraintElem{
		key:   fields[0] + "$" + fields[1],
		index: stringToInt(fields[2]),
		info:  stringToInt(fields[3]),
	}

	constraintMain.Add(newConst)

}

func AddConstraintInfo(info string) {
	fields := split(info, ':')
	if len(fields) != 2 {
		return
	}

	NumberElemConstraint = stringToInt(fields[0])
	NumberElemClosure = stringToInt(fields[1])
}

type Constraint struct {
	order           []ConstraintElem
	elems           map[string]map[int]bool // key -> index
	elemIndex       map[string]int          // key -> currentIndex
	replayIndex     int
	releasedTimeout map[string]bool

	mu mutex
}

func NewConstraint() {
	constraintMain = Constraint{
		order:       make([]ConstraintElem, 0),
		elems:       make(map[string]map[int]bool),
		elemIndex:   make(map[string]int),
		replayIndex: 0,
		mu:          mutex{},
	}
}

func (this *Constraint) Add(elem ConstraintElem) {
	this.order = append(this.order, elem)

	if printDebug {
		println(elem.key, " -> ", elem.index)
	}

	if _, ok := this.elems[elem.key]; !ok {
		this.elems[elem.key] = make(map[int]bool)
	}
	this.elems[elem.key][elem.index] = true
	this.elemIndex[elem.key] = 1
}

// func (this *Constraint) isNext() *ConstraintElem {
// 	lock(mu)
// 	elem := this.elems[ReplayIndex]
// 	unlock(mu)
// 	return elem
// }

func (this *Constraint) isNext(key string) (bool, int) {
	lock(&this.mu)
	next := this.order[this.replayIndex]
	unlock(&this.mu)
	return next.key == key, next.info
}

func (this *Constraint) getNext() *ConstraintElem {
	if this.replayIndex >= len(this.order) {
		return nil
	}

	lock(&this.mu)
	next := &this.order[this.replayIndex]
	unlock(&this.mu)
	return next
}

func (this *Constraint) isDone() bool {
	return len(this.order) == this.replayIndex
}

func GetReplayIndex() int {
	return constraintMain.replayIndex
}

/*
// Info determines if the operation of the key should be blocked or released
// Returns:
//	   bool: true if it should block, false if it can be released
*/
func (this *Constraint) isInConstraint(key string) bool {
	// lock(&this.mu)
	info, ok := this.elems[key]
	// unlock(&this.mu)

	// is never in constraintMain
	if !ok {
		return false
	}

	// is in constraints with this index
	lock(&this.mu)
	index := this.elemIndex[key]
	if info[index] {
		unlock(&this.mu)
		return true
	}
	unlock(&this.mu)

	// is in constraints but not with this index
	return false
}

func (this *Constraint) alreadyExecuted(key string, remove bool) bool {
	lock(&this.mu)

	_, ok := this.releasedTimeout[key]

	if remove {
		delete(this.releasedTimeout, key)
	}

	unlock(&this.mu)

	return ok
}

func (this *Constraint) update(key string) {
	lock(&this.mu)
	this.elemIndex[key]++
	this.replayIndex++
	unlock(&this.mu)

	this.releaseBlocked()
}

func (this *Constraint) releaseBlocked() {
	next := this.getNext()

	if next == nil {
		return
	}

	if c := waiting.contains(next); c.ch != nil {
		if printDebug {
			println("ReleaseW: ", next.key)
		}

		release(c.ch, next.info)
		waiting.remove(next)
	}
}

func (this *Constraint) releaseAllBlocked() {
	lock(&this.mu)

	for _, ch := range waiting.waiting {
		release(ch.ch, -1)
	}
	waiting.waiting = make(map[string]waitingOp)

	unlock(&this.mu)
}

func (this *Constraint) releaseOldest() {
	lock(&this.mu)

	oldestKey := ""
	var oldestTime int64 = -1

	for key, ch := range waiting.waiting {
		if oldestTime == -1 || ch.t < oldestTime {
			oldestKey = key
			oldestTime = ch.t
		}
	}

	if oldestTime != -1 {
		this.releasedTimeout[oldestKey] = true
		if printDebug {
			println("ReleaseT: ", oldestKey)
		}
		release(waiting.waiting[oldestKey].ch, -1)
	}
	unlock(&this.mu)
}
