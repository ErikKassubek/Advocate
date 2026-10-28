// Copyright (c) 2024 Erik Kassubek
//
// File: writer.go
// Brief: Write the internal trace into files
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package io

import (
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/bugs"
	"advocate/utils/timer"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// WriteTrace writes a trace to a file
//
// Parameter:
//   - traceToWrite *analysis.Trace: Pointer to the trace to write
//   - path string: The path to the file to write to
//   - replay bool: If true, write only the elements relevant for replay
//   - control bool: If true, write the active file to start in control mode
func WriteTrace(traceToWrite *trace.Trace, path string, replay, control bool) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	err := CreateTraceFolder(path)
	if err != nil {
		return err
	}

	// write the files

	numberRoutines := traceToWrite.GetNoRoutines()

	wg := sync.WaitGroup{}
	for i := 1; i <= numberRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			fileName := filepath.Join(path, "trace_"+strconv.Itoa(i)+".log")
			file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Error("Error in writing trace to file. Could not open file: ", err.Error())
			}
			defer file.Close()

			rout := traceToWrite.GetRoutine(i)
			// if rout == nil {
			// 	return
			// }

			// sort trace by tPre
			rout.Sort()

			for index, element := range rout.Elems() {
				if !replay || !isReplay(element) {
					continue
				}
				if !element.Committed() {
					element.SetT(trace.Sorting, math.MaxInt)
				}
				elementString := element.String()
				if _, err := file.WriteString(elementString); err != nil {
					log.Error("Error in writing trace to file. Could not write string: ", err.Error())
				}
				if index < rout.Len()-1 {
					if _, err := file.WriteString("\n"); err != nil {
						log.Error("Error in writing trace to file. Could not wrote string: ", err.Error())
					}
				}
			}
			if _, err := file.WriteString("\n"); err != nil {
				log.Error("Error in writing trace to file. Could not wrote string: ", err.Error())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	if control {
		writeControlFile(path)
	}

	return nil
}

func CreateTraceFolder(path string) error {
	// delete folder if exists
	if _, err := os.Stat(path); err == nil {
		log.Info(path + " already exists. Delete folder " + path)
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	// create new folder
	if err := os.Mkdir(path, 0755); err != nil {
		return err
	}

	return nil
}

// Check if the element is relevant for replay
//
// Parameter:
//   - element trace.TraceElement: element to check
//
// Returns:
//   - true if relevant for replay, false if ignored in replay
func isReplay(element trace.Element) bool {
	return !(element.Type(false) == trace.New || element.Type(true) == trace.RoutEnd)
}

// WriteRewriteInfoFile create a file with the result message and the exit code for the rewrite
//
// Parameter:
//   - path string: The path to the file folder to write to
//   - bug bugs.Bug: The rewritten bug
//   - exitCode int: The exit code
//   - resultIndex int: The index of the result
//
// Returns:
//   - error: The error that occurred
func WriteRewriteInfoFile(path string, bug bugs.Bug, exitCode int, resultIndex int) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	fileName := filepath.Join(path, paths.NameRewrittenInfo)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(strconv.Itoa(resultIndex+1) + "#" + bug.GetBugString() + "#" + strconv.Itoa(exitCode)); err != nil {
		return err
	}

	return nil
}

func writeControlFile(path string) {
	activePath := filepath.Join(path, paths.NameReplayConstraints)

	f, err := os.Create(activePath)
	if err != nil {
		return
	}

	defer f.Close()

	f.WriteString("0\n")
}
