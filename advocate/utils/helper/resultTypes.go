// Copyright (c) 2025 Erik Kassubek
//
// File: constants.go
// Brief: List of global constants
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

// ResultType is an ID for a type of result

package helper

// ResultType is an enum for the type of found bug
type ResultType string

// possible values for ResultType
const (
	Empty ResultType = ""

	// actual
	ASendOnClosed           ResultType = "A01"
	ACloseOnClosed          ResultType = "A02"
	ACloseOnNilChannel      ResultType = "A03"
	ANegWG                  ResultType = "A04"
	AUnlockOfNotLockedMutex ResultType = "A05"
	ABlocking               ResultType = "A06"
	ADeadlock               ResultType = "A07"
	AConcurrentRecv         ResultType = "A08"
	AMixedDeadlock          ResultType = "A09"

	// possible
	PSendOnClosed     ResultType = "P01"
	PNegWG            ResultType = "P02"
	PUnlockBeforeLock ResultType = "P03"
	PCyclicDeadlock   ResultType = "P04"
	PMixedDeadlock    ResultType = "P05"

	// leaks
	LUnknown   ResultType = "L00"
	LChan      ResultType = "L01"
	LNilChan   ResultType = "L02"
	LSelect    ResultType = "L03"
	LMutex     ResultType = "L04"
	LWaitGroup ResultType = "L05"
	LCond      ResultType = "L06"

	// recording
	RUnknownPanic ResultType = "R01"
	RTimeout      ResultType = "R02"
	RReplayFail   ResultType = "R03"

	// not executed select
	// SNotExecutedWithPartner = "S00"
)

var ResultTypes = []ResultType{
	ASendOnClosed,
	ACloseOnClosed,
	ACloseOnNilChannel,
	ANegWG,
	AUnlockOfNotLockedMutex,
	ABlocking,
	ADeadlock,
	AConcurrentRecv,
	AMixedDeadlock,
	PSendOnClosed,
	PNegWG,
	PUnlockBeforeLock,
	PCyclicDeadlock,
	PMixedDeadlock,
	LUnknown,
	LChan,
	LNilChan,
	LSelect,
	LMutex,
	LWaitGroup,
	LCond,
}

var ResultTypesActual = []ResultType{
	ASendOnClosed,
	ACloseOnClosed,
	ACloseOnNilChannel,
	ANegWG,
	AUnlockOfNotLockedMutex,
	ABlocking,
	ADeadlock,
	AConcurrentRecv,
	AMixedDeadlock,
}

var ResultTypesPotential = []ResultType{
	PSendOnClosed,
	PNegWG,
	PUnlockBeforeLock,
	PCyclicDeadlock,
	PMixedDeadlock,
}

var ResultTypesLeak = []ResultType{
	LUnknown,
	LChan,
	LNilChan,
	LSelect,
	LMutex,
	LWaitGroup,
	LCond,
}

var ResultTypesRecording = []ResultType{
	RUnknownPanic,
	RTimeout,
	RReplayFail,
}

func ResultTypeFromString(code string) ResultType {
	switch code {
	case "A01":
		return ASendOnClosed
	case "A02":
		return ACloseOnClosed
	case "A03":
		return ACloseOnNilChannel
	case "A04":
		return ANegWG
	case "A05":
		return AUnlockOfNotLockedMutex
	case "A06":
		return ABlocking
	case "A07":
		return ADeadlock
	case "A08":
		return AConcurrentRecv
	case "A09":
		return AMixedDeadlock
	case "P01":
		return PSendOnClosed
	case "P02":
		return PNegWG
	case "P03":
		return PUnlockBeforeLock
	case "P04":
		return PCyclicDeadlock
	case "P05":
		return PMixedDeadlock
	case "L00":
		return LUnknown
	case "L01":
		return LChan
	case "L02":
		return LNilChan
	case "L03":
		return LSelect
	case "L04":
		return LMutex
	case "L05":
		return LWaitGroup
	case "L06":
		return LCond
	case "R01":
		return RUnknownPanic
	case "R02":
		return RTimeout
	case "R03":
		return RReplayFail
	default:
		return Empty // Return Empty for codes not found
	}
}

// Values for the possible program exit codes
const (
	ExitCodeNone             = -1
	ExitCodePanic            = 3
	ExitCodeTimeout          = 10
	ExitCodeLeakUnbuf        = 20
	ExitCodeLeakBuf          = 21
	ExitCodeLeakMutex        = 22
	ExitCodeLeakCond         = 23
	ExitCodeLeakWG           = 24
	ExitCodeSendClose        = 30
	ExitCodeRecvClose        = 31
	ExitCodeCloseClose       = 32
	ExitCodeCloseNil         = 33
	ExitCodeNegativeWG       = 34
	ExitCodeUnlockBeforeLock = 35
	ExitCodeCyclic           = 41
	ExitCodeMixedDeadlock    = 42
)

// MinExitCodeSuc is the minimum exit code for successful replay
const MinExitCodeSuc = ExitCodeLeakUnbuf

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
