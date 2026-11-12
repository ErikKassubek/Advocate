// Copyright (c) 2025 
//
// File: control.go
// Brief: Cancel analysis when not enough memory
//
//
// License: BSD-3-Clause

package control

import (
	"gosc/utils/flags"
	"gosc/utils/log"
	"context"
	"math"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
)

// variables for memory management
var (
	isCanceled    atomic.Bool
	IsCanceledRAM atomic.Bool

	numberCommands  int
	commandsLock    = sync.Mutex{}
	runningCommands = make(map[int]context.CancelFunc)
)

// SetMaxNumberElem sets the max number elements
func SetMaxNumberElem() {
	if flags.MaxNumberElements < 0 {
		flags.MaxNumberElements = math.MaxInt
	}
}

// Supervisor periodically checks the used and free memory
// If the trace is to big and the available RAM to small, this can lead
// to problems. In this case we abort the analysis
func Supervisor(baseAClearTrace, baseAClearData, baseFClear func()) {
	// Get the memory stats
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("Error getting memory info: %v", err)
	}

	// Get the swap stats
	// s, err := mem.SwapMemory()
	// if err != nil {
	// 	log.Errorf("Error getting swap info: %v", err)
	// }

	thresholdRAM := uint64(float64(v.Total) * 0.15)
	// thresholdSwap := uint64(1025 * 1024 * 1024) // 1GB

	// startSwap := s.Used

	for {
		// Get the memory stats
		v, err = mem.VirtualMemory()
		if err != nil {
			log.Errorf("Error getting memory info: %v", err)
		}

		// Get the swap stats
		// s, err = mem.SwapMemory()
		// if err != nil {
		// 	log.Errorf("Error getting swap info: %v", err)
		// }

		// cancel if available RAM is below the threshold or the used swap is above the threshold
		if v.Available < thresholdRAM {
			cancelRAM(baseAClearTrace, baseAClearData, baseFClear)
			time.Sleep(5 * time.Second)
			continue
		} else {
			Reset()
		}

		// if s.Used > thresholdSwap+startSwap {
		// 	cancelRAM()
		// 	time.Sleep(5 * time.Second)
		// 	continue
		// } else {
		// 	Reset()
		// }

		// Sleep for a while before checking again
		time.Sleep(500 * time.Millisecond)
	}
}

// Cancel sets the analysis to canceled
func Cancel() {
	isCanceled.Store(true)
}

// Cancel the analysis if not enough ram is available
func cancelRAM(baseAClearTrace, baseAClearData, baseFClear func()) {
	isCanceled.Store(true)
	IsCanceledRAM.Store(true)
	log.Error("Not enough RAM")
	printAllGoroutines()
	cancelAllRunningCom()

	baseAClearTrace()
	baseAClearData()
	baseFClear()

	// give all function time to cancel and then make sure to clear the memory
	time.Sleep(2 * time.Second)
	runtime.GC()
	debug.FreeOSMemory()
}

// WasCanceled returns if the analysis was canceled
//
// Returns:
//   - bool: true if the analysis was canceled
func WasCanceled() bool {
	return isCanceled.Load()
}

// WasCanceledRAM returns if the analysis was canceled because of insufficient ram
//
// Returns:
//   - bool: true if the analysis was canceled because of insufficient ram*
func WasCanceledRAM() bool {
	return IsCanceledRAM.Load()
}

// Reset the cancel values to false
func Reset() {
	if WasCanceled() {
		log.Important("RAM Reset")
		isCanceled.Store(false)
		IsCanceledRAM.Store(false)
	}
}

// AddRunningCom stores the cancel function for a context of a running command
// It returns an ID, with which the command can be removed
//
// Parameter:
//   - cancel context.CancelFunc: the context cancel function
//
// Runtime:
//   - int: an id with which the element can be removed
func AddRunningCom(cancel context.CancelFunc) int {
	commandsLock.Lock()
	defer commandsLock.Unlock()

	numberCommands++

	id := numberCommands
	runningCommands[id] = cancel

	return id
}

// RemoveRunningCom removes a cancel function from the memory monitor
//
// Paramter:
//   - id int: the id of the parameter to remove
func RemoveRunningCom(id int) {
	commandsLock.Lock()
	defer commandsLock.Unlock()

	delete(runningCommands, id)
}

// cancelAllRunningCom cancels all currently running commands
func cancelAllRunningCom() {
	commandsLock.Lock()
	defer commandsLock.Unlock()

	for _, cancel := range runningCommands {
		cancel()
	}

	runningCommands = make(map[int]context.CancelFunc)
}
