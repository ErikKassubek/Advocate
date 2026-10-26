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
	ARecvOnClosed           ResultType = "A02"
	ACloseOnClosed          ResultType = "A03"
	ACloseOnNilChannel      ResultType = "A04"
	ANegWG                  ResultType = "A05"
	AUnlockOfNotLockedMutex ResultType = "A06"
	ABlocking               ResultType = "A07"
	ADeadlock               ResultType = "A08"
	AConcurrentRecv         ResultType = "A09"
	AMixedDeadlock          ResultType = "A10"

	// possible
	PSendOnClosed     ResultType = "P01"
	PRecvOnClosed     ResultType = "P02"
	PNegWG            ResultType = "P03"
	PUnlockBeforeLock ResultType = "P04"
	PCyclicDeadlock   ResultType = "P05"
	PMixedDeadlock    ResultType = "P06"

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

	// not executed select
	// SNotExecutedWithPartner = "S00"
)

var ResultTypes = []ResultType{
	ASendOnClosed,
	ARecvOnClosed,
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
	ARecvOnClosed,
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
	// PRecvOnClosed,
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
}

func ResultTypeFromString(code string) ResultType {
	switch code {
	case "A01":
		return ASendOnClosed
	case "A02":
		return ARecvOnClosed
	case "A03":
		return ACloseOnClosed
	case "A04":
		return ACloseOnNilChannel
	case "A05":
		return ANegWG
	case "A06":
		return AUnlockOfNotLockedMutex
	case "A07":
		return ABlocking
	case "A08":
		return ADeadlock
	case "A09":
		return AConcurrentRecv
	case "P01":
		return PSendOnClosed
	case "P02":
		return PRecvOnClosed
	case "P03":
		return PNegWG
	case "P04":
		return PUnlockBeforeLock
	case "P05":
		return PCyclicDeadlock
	case "P06":
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
	case ARecvOnClosed:
		return "recv", "close"
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
	case PRecvOnClosed:
		return "recv", "close"
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
	}

	return "", ""
}
