// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_replay.go
// Brief: Advocate tracing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package advocatego

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	_ "unsafe"
)

var traceFileCounter = 0
var tracePathRecorded = "advocateTrace"

var hasFinished = false

var timerStarted = false
var startTime time.Time
var duration time.Duration

var startWriting = false

var initTracing bool

// InitTracing initializes the tracing.
// The function creates the trace folder and starts the background memory test.
//
//go:linkname InitTracing runtime.AdvocateInitTracing
func InitTracing(timeout int, init bool) {
	if initTracing { // called by main but alredy run by init
		return
	}
	initTracing = true

	FinishFunc = FinishTracing

	startTime = time.Now()
	timerStarted = true

	// remove the trace folder if it exists
	err := os.RemoveAll(tracePathRecorded)
	if err != nil && !os.IsNotExist(err) {
		println("Cannot remove: ", err.Error())
		return
	}

	// create the trace folder
	err = os.Mkdir(tracePathRecorded, 0755)
	if err != nil {
		if !os.IsExist(err) {
			println("Cannot write: ", err.Error())
			return
		}
	}

	if timeout > 0 {
		// start time timeout
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			panic("Timeout")
		}()
	}

	// DetectBlockingGC(1000)

	// go writeTraceIfFull()
	// go removeAtomicsIfFull()
	runtime.InitTracing(FinishTracing, WriteToTraceFile, init)
}

// Write the trace of the program to a file.
// The trace is written in the file named file_name.
// The trace is written in the format of advocate.
//
//go:linkname FinishTracing runtime.AdvocateFinishTracing
func FinishTracing() {
	if hasFinished {
		// needed to prevent program stop while still writing
		// otherwise, trace may be empty
		time.Sleep(time.Second * 20)
		return
	}
	hasFinished = true

	if !finishFuzzingStarted {
		time.Sleep(time.Second)
	}

	runtime.AdvocatRoutineExit()

	runtime.DisableTracing()

	runtime.BuildOAT()

	// DetectBlockingGC()

	if timerStarted {
		duration = time.Since(startTime)
	}

	writeToTraceFiles()
}

// Write the trace to a set of files. The traces are written into a folder
// with name trace. For each routine, a file is created. The file is named
// trace_routineId.log. The trace of the routine is written into the file.
func writeToTraceFiles() {
	startWriting = true
	numRout := runtime.GetNumberOfRoutines()
	writeToTraceFileInfo(numRout)

	currentlyWriting := make([]int, 0)

	for i := 1; i <= numRout+1; i++ {
		active, writing := runtime.IsActive(i)
		if !active {
			continue
		}
		if writing {
			currentlyWriting = append(currentlyWriting, i)
			continue
		}
		// write the trace to the file
		WriteToTraceFile(i, false)
	}

	// wait for the currently writing routine to be written
	for len(currentlyWriting) > 0 {
		for i := 0; i < len(currentlyWriting); {
			active, _ := runtime.IsActive(currentlyWriting[i])
			if !active {
				currentlyWriting = append(currentlyWriting[:i], currentlyWriting[i+1:]...)
			} else {
				i++
			}
		}
	}
}

// Write the trace of a routine to a file.
// The trace is written in the file named trace_routineId.log.
// The trace is written in the format of advocate.
//
// Parameter:
//   - routine: The id of the routine
//   - fromRuntime bool: true if the function was called from runtime
//
// Return:
//   - true, if the trace can be written, false if the writing of the remaining routines has already started
func WriteToTraceFile(routine int, fromRuntime bool) bool {
	if fromRuntime && startWriting {
		return false
	}

	// create the file if it does not exist and open it
	fileName := filepath.Join(tracePathRecorded, "trace_"+strconv.Itoa(routine)+".log")

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	c := runtime.TraceToChanByID(uint64(routine))

	for res := range c {
		if _, err := file.WriteString(res); err != nil {
			panic(err)
		}
	}

	return true
}

/*
 * Write a trace info file
 */
func writeToTraceFileInfo(numberRoutines int) {
	fileName := filepath.Join(tracePathRecorded, "trace_info.log")

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reachedActive := 0

	// TODO: make this work again

	numberReleased := runtime.GetReplayIndex()

	if numberReleased >= runtime.NumberElemClosure {
		reachedActive = 1
	}
	allActiveReleased := 0
	if numberReleased >= runtime.NumberElemConstraint+runtime.NumberElemClosure {
		allActiveReleased = 1
	}

	exitCode, exitPos := runtime.GetExitCode()
	replayOldest, replayDisabled, replayAck := runtime.GetReplayInfo()

	// exitCode, exitPos := 0, ""
	// replayOldest, replayDisabled, replayAck := 0, 0, 0

	file.WriteString(fmt.Sprintf("ExitCode!%d\n", exitCode))
	file.WriteString(fmt.Sprintf("ExitPosition!%s\n", exitPos))
	file.WriteString(fmt.Sprintf("ReplayTimeout!%d\n", replayOldest))
	file.WriteString(fmt.Sprintf("ReplayDisabled!%d\n", replayDisabled))
	file.WriteString(fmt.Sprintf("ReplayAck!%d\n", replayAck))
	file.WriteString(fmt.Sprintf("NumberRoutines!%d\n", numberRoutines))
	file.WriteString(fmt.Sprintf("ActiveReached!%d\n", reachedActive))
	file.WriteString(fmt.Sprintf("AllActiveReleased!%d\n", allActiveReleased))
	if timerStarted {
		file.WriteString(fmt.Sprintf("Runtime!%d", int(duration.Seconds())))
	} else {
		file.WriteString("Runtime:0")
	}

}

// Delete empty files in the trace folder.
// The function deletes all files in the trace folder that are empty.
// func deleteEmptyFiles() {
// 	files, err := os.ReadDir(tracePathRecorded)
// 	if err != nil {
// 		panic(err)
// 	}

// 	for _, file := range files {
// 		if file.IsDir() {
// 			continue
// 		}

// 		stat, err := os.Stat(tracePathRecorded + "/" + file.Name())
// 		if err != nil {
// 			continue
// 		}
// 		if stat.Size() == 0 {
// 			err := os.Remove(tracePathRecorded + "/" + file.Name())
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 	}

// }
