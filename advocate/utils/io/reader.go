// Copyright (c) 2024 Erik Kassubek
//
// File: reader.go
// Brief: Read trace files and create the internal trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package io

import (
	"advocate/analysis/a_base"
	"advocate/trace"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/timer"
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CreateTraceFromFiles creates the trace from all files in a folder.
//
// Parameter:
//   - filePath string: The path to the folder
//   - asMain bool: if true, set as main trace
//
// Returns:
//   - *trace.Trace: the trace
//   - int: The number of routines
//   - int: The number of elements
//   - error: An error if the trace could not be created
func CreateTraceFromFiles(folderPath string, asMain bool) (*trace.Trace, int, int, error) {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	numberRoutines := 0
	// traverse all files in the folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, 0, 0, err
	}

	tr := trace.NewTrace()

	elemCounter := 0
	for _, file := range files {
		name := file.Name()

		if file.IsDir() {
			continue
		}

		if name == paths.NameTimes {
			continue
		}

		filePath := filepath.Join(folderPath, name)

		if name == paths.NameTraceInfo {
			getTraceInfoFromFile(filePath)
		}

		routine, err := getRoutineFromFileName(name)
		if err != nil {
			if name != paths.NameTraceInfo && name != paths.NameReplayConstraints && name != paths.NameReplayForkTree {
				log.Error("Could not get routine from file name ", name)
			}
			continue
		}

		numberElems, err := createTraceFromFile(&tr, filePath, routine)

		if err != nil {
			return nil, 0, elemCounter, err
		}
		elemCounter += numberElems
		numberRoutines++

		if elemCounter > flags.MaxNumberElements {
			return &tr, numberRoutines, elemCounter, fmt.Errorf("Too many elements")
		}

		if control.WasCanceled() {
			return &tr, numberRoutines, elemCounter, fmt.Errorf("Canceled by memory")
		}
	}

	tr.Sort()

	if asMain {
		a_base.SetMainTrace(&tr)
	}

	return &tr, numberRoutines, elemCounter, nil
}

// getTraceInfoFromFile reads in the information from a the trace_info.log file
//
// Parameter:
//   - filePath string: the path to the trace_info.log file
//
// Returns:
//   - error
func getTraceInfoFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Error opening file: " + filePath)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	exitCode := 0
	exitPos := ""

	timeoutOldest := 0
	timeoutDisabled := 0
	timeoutAck := 0
	activeReleased := 0
	allActiveReleased := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineSplit := strings.Split(line, "!")
		if len(lineSplit) != 2 {
			continue
		}

		var err error

		switch lineSplit[0] {
		case "Runtime":
			rt, err := strconv.Atoi(lineSplit[1])
			if err == nil {
				a_base.SetRuntimeDurationSec(rt)
			}
		case "ExitCode":
			ec, err := strconv.Atoi(lineSplit[1])
			if err == nil {
				exitCode = ec
			}
		case "ExitPosition":
			exitPos = lineSplit[1]
		case "ReplayTimeout":
			timeoutOldest, err = strconv.Atoi(lineSplit[1])
		case "ReplayDisabled":
			timeoutDisabled, err = strconv.Atoi(lineSplit[1])
		case "ReplayAck":
			timeoutAck, err = strconv.Atoi(lineSplit[1])
		case "ActiveReached":
			var activeReplayReached int
			activeReplayReached, err = strconv.Atoi(lineSplit[1])
			if activeReplayReached > 0 {
				activeReleased = 1
			}
		case "AllActiveReleased":
			allActiveReleased, _ = strconv.Atoi(lineSplit[1])
		}

		if err != nil {
			log.Error(err.Error())
		}
	}

	a_base.SetExitInfo(exitCode, exitPos)
	a_base.SetReplayInfo(timeoutOldest, timeoutDisabled, timeoutAck, activeReleased, allActiveReleased)

	return nil
}

// Read and build the trace from a file
//
// Parameter:
//   - tr *trace.Trace: the trace to add the elements to
//   - filePath string: The path to the log file
//   - routine int: The routine id
//
// Returns:
//   - int: number of elements
//   - error: An error if the trace could not be created
func createTraceFromFile(tr *trace.Trace, filePath string, routine int) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Error opening file: " + filePath)
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	tr.AddRoutine(routine)

	counter := 0
	for scanner.Scan() {
		line := scanner.Text()
		err := processElement(tr, line, routine)
		if err != nil {
			log.Errorf("Error in processing trace element %s: ", line, err)
		}
		counter++

		if control.WasCanceled() {
			return counter, fmt.Errorf("Canceled by memory")
		}
	}

	return counter, scanner.Err()
}

// Process one element from the log file.
//
// Parameter:
//   - tr *trace.Trace: the trace to add the elements to
//   - element string: The element to process
//   - routine int: The routine id, equal to the line number
//
// Returns:
//   - error: An error if the element could not be processed
func processElement(tr *trace.Trace, element string, routine int) error {
	if element == "" {
		return errors.New("Element is empty")
	}
	fields := strings.Split(element, ",")
	var err error
	switch fields[0] {
	case "A":
		if flags.IgnoreAtomics {
			return nil
		}
		err = tr.AddTraceElementAtomic(routine, fields[1], fields[2], fields[3], fields[4])
	case "C":
		if len(fields) != 11 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 11", element, len(fields))
		}
		err = tr.AddTraceElementChannel(routine, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7], fields[8], fields[9], fields[10])
	case "M":
		if len(fields) != 9 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 9", element, len(fields))
		}
		err = tr.AddTraceElementMutex(routine, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7], fields[8])
	case "G":
		if len(fields) != 5 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 5", element, len(fields))
		}
		err = tr.AddTraceElementFork(routine, fields[1], fields[2], fields[3])
	case "S":
		if len(fields) != 8 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 6", element, len(fields))
		}
		err = tr.AddTraceElementSelect(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6], fields[7])
	case "W":
		if len(fields) != 9 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 7", element, len(fields))
		}
		err = tr.AddTraceElementWait(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6], fields[7], fields[8])
	case "O":
		if len(fields) != 7 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 7", element, len(fields))
		}
		err = tr.AddTraceElementOnce(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "D":
		if len(fields) != 7 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 7", element, len(fields))
		}
		err = tr.AddTraceElementCond(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "N":
		if len(fields) != 7 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 7", element, len(fields))
		}
		err = tr.AddTraceElementAlloc(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "E":
		if len(fields) != 3 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 3", element, len(fields))
		}
		err = tr.AddTraceElementRoutineEnd(routine, fields[1])
	case "F":
		if len(fields) != 6 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 6", element, len(fields))
		}
		err = tr.AddTaceElementFunc(routine, fields[1], fields[2], fields[3], fields[4])
	case "R":
		if len(fields) != 3 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 3", element, len(fields))
		}
		err = tr.AddTaceElementReturn(routine, fields[1])
	case "I":
		if len(fields) != 7 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 7", element, len(fields))
		}
		err = tr.AddTraceElementControllFlow(routine, fields[1], fields[2], fields[3], fields[4], fields[5])
	case "OAT":
		err = tr.AddTraceObjectAware(routine, fields[1])
	default:
		return errors.New("Unknown element type in: " + element)
	}

	if err != nil {
		return err
	}

	return nil
}

// getRoutineFromFileName extracts the file ID from a trace file. Trace files
// always have the name trace_[ID]
//
// Parameter:
//   - fileName string: name of the trace file
//
// Returns:
//   - int: if fileName is valid the trace id, otherwise 0
//   - error
func getRoutineFromFileName(fileName string) (int, error) {
	// the file name is "trace_routineID.log"
	// remove the .log at the end
	fileName1 := strings.TrimSuffix(fileName, ".log")
	if fileName1 == fileName {
		return 0, errors.New("File name does not end with .log")
	}

	fileName2 := strings.TrimPrefix(fileName1, "trace_")
	if fileName2 == fileName1 {
		return 0, errors.New("File name does not start with trace_")
	}

	routine, err := strconv.Atoi(fileName2)
	if err != nil {
		return 0, err
	}

	return routine, nil
}
