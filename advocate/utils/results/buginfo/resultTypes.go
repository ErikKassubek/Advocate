// Copyright (c) 2025 Erik Kassubek
//
// File: ResultType.go
// Brief: Result argument types
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package buginfo

import (
	"advocate/utils/log"
)

// ResultType is an enum for the type of found bug
type ResultType string

func (rt ResultType) IsLeak() bool {
	return string(rt)[0] == 'L'
}

func (rt ResultType) IsBlocking() bool {
	return rt == ABlocking || rt == ADeadlock
}

func (rt ResultType) IsPos() bool {
	return string(rt)[0] == 'P'
}

func (rt ResultType) IsActual() bool {
	return string(rt)[0] == 'A'
}

func (rt ResultType) Args() (string, string) {
	switch rt {
	case ASendOnClosed:
		return "send", "close"
	case ACloseOnClosed:
		return "panic", "close"
	case ACloseOnNilChannel:
		return "close", ""
	case ANegWG:
		return "done", ""
	case AUnlockOfNotLockedMutex:
		return "unlock", ""
	case ABlocking:
		return "blocking", ""
	case ADeadlock:
		return "blocking", "" // TODO: add holding
	case AConcurrentRecv:
		return "recv", "recv"
	case AMixedDeadlock:
		return "", ""

	// possible
	case PSendOnClosed:
		return "send", "close"
	case PNegWG:
		return "done", "add"
	case PUnlockBeforeLock:
		return "unlock", "lock"
	case PCyclicDeadlock:
		return "blocking", "acquired"
	case PMixedDeadlock:
		return "send", "recv"

	// leaks
	case LUnknown:
		return "leak", ""
	case LChan:
		return "leak", ""
	case LNilChan:
		return "leak", ""
	case LSelect:
		return "leak", ""
	case LMutex:
		return "leak", ""
	case LWaitGroup:
		return "leak", ""
	case LCond:
		return "leak", ""

	// recording
	case RUnknownPanic:
		return "panic", ""
	case RTimeout:
		return "", ""
	case RReplayFail:
		return "", ""
	}

	return "", ""
}

func (rt ResultType) String() string {
	switch rt {
	case RUnknownPanic:
		return "Unknown Panic:"
	case RTimeout:
		return "Timeout"
	case RReplayFail:
		return "ReplayFail"
	case ASendOnClosed:
		return "Actual Send on Closed Channel:"
	case ACloseOnClosed:
		return "Actual Close on Closed Channel:"
	case ACloseOnNilChannel:
		return "Actual close on nil channel:"
	case AConcurrentRecv:
		return "Concurrent Receive:"
	case ABlocking:
		return "Blocking routine:"
	case ADeadlock:
		return "Deadlock:"
	case ANegWG:
		return "Actual negative Wait Group:"
	case AUnlockOfNotLockedMutex:
		return "Actual unlock of not locked mutex:"
	case PSendOnClosed:
		return "Possible send on closed channel:"
	case PNegWG:
		return "Possible negative waitgroup counter:"
	case PUnlockBeforeLock:
		return "Possible unlock of a not locked mutex:"
	case PCyclicDeadlock:
		return "Possible cyclic deadlock:"
	case PMixedDeadlock:
		return "Possible Mixed Deadlock:"
	case LUnknown:
		return "Leak on routine"
	case LChan:
		return "Leak on a channel:"
	case LNilChan:
		return "Leak on nil channel:"
	case LSelect:
		return "Leak on select:"
	case LMutex:
		return "Leak on mutex:"
	case LWaitGroup:
		return "Leak on wait group:"
	case LCond:
		return "Leak on conditional variable:"
	default:
		log.Error("Unknown bug type in toString: " + string(rt))
		return ""
	}
}
