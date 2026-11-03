// Copyright (c) 2025 Erik Kassubek
//
// File: bugInfo.go
// Brief: Bug info constants
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

// ResultType is an ID for a type of result

package buginfo

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

var ResultTypeMap = map[ResultType]string{
	ASendOnClosed:           "Actual Send on Closed Channel",
	ACloseOnClosed:          "Actual Close on Closed Channel",
	AConcurrentRecv:         "Concurrent Receive",
	ACloseOnNilChannel:      "Actual close on nil channel",
	ABlocking:               "Actual Non-Cyclic Blocking Bug",
	ADeadlock:               "Actual Cyclic Deadlock",
	ANegWG:                  "Actual negative Wait Group",
	AUnlockOfNotLockedMutex: "Actual unlock of not locked mutex",

	PSendOnClosed:     "Possible send on closed channel",
	PNegWG:            "Possible negative waitgroup counter",
	PUnlockBeforeLock: "Possible unlock of a not locked mutex",
	PCyclicDeadlock:   "Possible cyclic deadlock",
	PMixedDeadlock:    "Possible Mixed Deadlock",

	LUnknown:   "Leak on routine or unknown element",
	LChan:      "Leak on channel",
	LNilChan:   "Leak on nil channel",
	LSelect:    "Leak on select w",
	LMutex:     "Leak on mutex",
	LWaitGroup: "Leak on wait group",
	LCond:      "Leak on conditional variable",

	RUnknownPanic: "Unknown Panic",
	RTimeout:      "Timeout",
	RReplayFail:   "ReplayFailed",

	// SNotExecutedWithPartner: "Not executed select with potential partner",
}

// MinExitCodeSuc is the minimum exit code for successful replay
const MinExitCodeSuc = ExitCodeLeakUnbuf
