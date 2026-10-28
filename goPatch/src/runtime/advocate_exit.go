// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_exit.go
// Brief: Functionality for the exit codes
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

var advocateExitCode = 0
var advocateExitCodePos = ""
var hasReturnedExitCode = false

const (
	ExitCodeDefault          = 0
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

var ExitCodeNames = map[int]string{
	0:  "The replay terminated normally",
	3:  "The program panicked unexpectedly",
	10: "Timeout",
	20: "Leak: Leaking unbuffered channel or select was unstuck",
	21: "Leak: Leaking buffered channel was unstuck",
	22: "Leak: Leaking Mutex was unstuck",
	23: "Leak: Leaking Cond was unstuck",
	24: "Leak: Leaking WaitGroup was unstuck",
	30: "Send on close",
	31: "Receive on close",
	32: "Close on close",
	33: "Close on nil",
	34: "Negative WaitGroup counter",
	35: "Unlock of unlocked mutex",
	41: "Cyclic deadlock",
	42: "Mixed Deadlock",
}

// GetExitCode returns the exit code and exit position
//
// Returns:
//   - int: exit code
//   - string: exit position
func GetExitCode() (int, string) {
	return advocateExitCode, advocateExitCodePos
}

// SetExitCodeFromPanicMsg sets the panic info from the panic message
//
// Parameter:
//   - msg any: the panic message
func SetExitCodeFromPanicMsg(msg any) {
	skip := 3

	switch m := msg.(type) {
	case plainError:
		if m.Error() == "send on closed channel" {
			advocateExitCode = ExitCodeSendClose
			skip = 5
		} else if m.Error() == "close of closed channel" {
			advocateExitCode = ExitCodeCloseClose
			skip = 4
		} else if m.Error() == "close of nil channel" {
			advocateExitCode = ExitCodeCloseNil
			skip = 4
		}
	case string:
		if m == "sync: negative WaitGroup counter" {
			advocateExitCode = ExitCodeNegativeWG
			skip = 5
		} else if hasPrefix(m, "test timed out") || hasPrefix(m, "Timeout") {
			advocateExitCode = ExitCodeTimeout
		} else if m == "sync: unlock of unlocked mutex" {
			advocateExitCode = ExitCodeUnlockBeforeLock
			skip = 6
		} else if m == "sync: Unlock of unlocked RWMutex" {
			advocateExitCode = ExitCodeUnlockBeforeLock
			skip = 4
		} else if m == "sync: RUnlock of unlocked RWMutex" {
			advocateExitCode = ExitCodeUnlockBeforeLock
			skip = 5
		} else if m == "Timeout" {
			advocateExitCode = ExitCodeTimeout
			skip = 0
		}
	default:
		var p _panic
		p.arg = msg
		preprintpanics(&p)
		printpanics(&p)
		print("\n")
		printAllGoroutines()
	}

	_, file, line, _ := Caller(skip)
	advocateExitCodePos = file + posSep + intToString(line)

	if advocateExitCode == 0 {
		advocateExitCode = ExitCodePanic
	}
}

// Exit the program with the given code if the program panics.
//
// Parameter:
//   - msg: the panic message
func AdvocatePanic(msg any) {
	// if hasPanicked {
	// 	println("Second Panic")
	// 	return
	// }
	// hasPanicked = true

	constraintMain.releaseAllBlocked()

	SetExitCodeFromPanicMsg(msg)

	if IsAdvocateFuzzingEnabled() {
		finishFuzzingFunc()
	} else if IsTracingEnabled() {
		finishTracingFunc()
	}

	// if !IsReplayEnabled() {
	// 	return
	// }

	ExitReplayWithCode(advocateExitCode, msg)
}

// Exit the program with the given code.
//
// Parameter:
//   - code: the exit code
//   - msg any: message
func ExitReplayWithCode(code int, msg any) {
	if !hasReturnedExitCode {
		// if !isExitCodeConfOnEndElem(code) && !stuckReplayExecutedSuc {
		// 	return
		// }
		println("\nExit Replay with code ", code, ExitCodeNames[code])
		hasReturnedExitCode = true
	} else {
		println("Exit code already returned")
	}

	if msg != "" {
		print("Exit Message: ")

		var p _panic
		p.arg = msg
		preprintpanics(&p)
		printpanics(&p)
		print("\n")
	}

	exit(int32(code))
}

func ExitReplayTimeout() {
	advocateExitCode = ExitCodeTimeout
	AdvocatePanic("Timeout")
}
