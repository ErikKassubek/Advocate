// Copyright (c) 2026 
//
// File: gosc_partial_deadlock.go
// Brief: Detect partial deadlocks while running
//
//
// License: BSD-3-Clause

package runtime

var blockedConcurrencyReasons = []waitReason{
	waitReasonChanReceiveNilChan,
	waitReasonChanSendNilChan,
	waitReasonSelect,
	waitReasonSelectNoCases,
	waitReasonChanReceive,
	waitReasonChanSend,
	waitReasonSyncCondWait,
	waitReasonSyncMutexLock,
	waitReasonSyncRWMutexRLock,
	waitReasonSyncRWMutexLock,
	waitReasonSyncWaitGroupWait,
}

type GoscG struct {
	rout *g
}

func (self *GoscG) parkForeverReplay() bool {
	return self.rout.goscRoutineInfo.parkForeverReplay
}

func (self *GoscG) isNil() bool {
	return self.rout.goscRoutineInfo == nil
}

func (self *GoscG) id() uint64 {
	return self.rout.goscRoutineInfo.id
}

func (self *GoscG) goId() uint64 {
	return self.rout.goid
}

func (self *GoscG) getForkPos() string {
	return self.rout.goscRoutineInfo.GetForkPos()
}

func (self *GoscG) setOAT(oat []uint64) {
	self.rout.goscRoutineInfo.oat = oat
}

func ForEachGoscG(fn func(adGp *GoscG)) {
	forEachG(func(gp *g) {
		fn(&GoscG{gp})
	})
}

// GetWaitingReasonString takes a waitReason of a routine and returns a
// string representation
//
// Parameter:
//   - wr waitReason: the wait reason enum value
//
// Returns:
//   - string: the string representation of wr
func GetWaitingReasonString(wr waitReason) string {
	switch wr {
	case waitReasonChanReceiveNilChan:
		return "chan:recvOnNil"
	case waitReasonChanSendNilChan:
		return "chan:sendOnNil"
	case waitReasonSelect:
		return "select:select"
	case waitReasonSelectNoCases:
		return "select:withoutCases"
	case waitReasonChanReceive:
		return "chan:revc"
	case waitReasonChanSend:
		return "chan:send"
	case waitReasonSyncCondWait:
		return "cond:wait"
	case waitReasonSyncMutexLock:
		return "mutex:lock"
	case waitReasonSyncRWMutexRLock:
		return "rwmutex:rlock"
	case waitReasonSyncRWMutexLock:
		return "rwmutex:lock"
	case waitReasonSyncWaitGroupWait:
		return "waitGroup:wait"
	}
	return "unknown:unknown"
}

// isRoutineWaitingOnConcurrency determines if an gosc routine is blocked on a concurreny operations
// Returns:
//   - bool: true blocked on concurrency op, false otherwise
func (self *GoscG) LastTraceElem() traceElem {
	if self.isNil() {
		return nil
	}

	l := len(self.rout.goscRoutineInfo.Trace)

	if l == 0 {
		return nil
	}

	return self.rout.goscRoutineInfo.Trace[l-1]
}
