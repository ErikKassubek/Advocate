// Copyright (c) 2024 Erik Kassubek
//
// File: results.go
// Brief: Function for debug results and for results found bugs
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package results

import (
	"advocate/trace"
	"advocate/utils/consts"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/benign"
	"advocate/utils/types"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// resultLevel is an enum for the severity of a result
type resultLevel int

// values for the resultLevel enum
const (
	NONE resultLevel = iota
	CRITICAL
	WARNING
	INFORMATION
)

var resultTypeMap = map[helper.ResultType]string{
	helper.ASendOnClosed:           "Actual Send on Closed Channel",
	helper.ARecvOnClosed:           "Actual Receive on Closed Channel",
	helper.ACloseOnClosed:          "Actual Close on Closed Channel",
	helper.AConcurrentRecv:         "Concurrent Receive",
	helper.ACloseOnNilChannel:      "Actual close on nil channel",
	helper.ABlocking:               "Actual Non-Cyclic Blocking Bug",
	helper.ADeadlock:               "Actual Cyclic Deadlock",
	helper.ANegWG:                  "Actual negative Wait Group",
	helper.AUnlockOfNotLockedMutex: "Actual unlock of not locked mutex",

	helper.PSendOnClosed:     "Possible send on closed channel",
	helper.PNegWG:            "Possible negative waitgroup counter",
	helper.PUnlockBeforeLock: "Possible unlock of a not locked mutex",
	helper.PCyclicDeadlock:   "Possible cyclic deadlock",
	helper.PMixedDeadlock:    "Possible Mixed Deadlock",

	helper.LUnknown:   "Leak on routine or unknown element",
	helper.LChan:      "Leak on channel",
	helper.LNilChan:   "Leak on nil channel",
	helper.LSelect:    "Leak on select w",
	helper.LMutex:     "Leak on mutex",
	helper.LWaitGroup: "Leak on wait group",
	helper.LCond:      "Leak on conditional variable",

	helper.RUnknownPanic: "Unknown Panic",
	helper.RTimeout:      "Timeout",

	// SNotExecutedWithPartner: "Not executed select with potential partner",
}

var (
	outputReadableFile       string
	outputMachineFile        string
	foundBug                 = false
	resultsWarningReadable   []string
	resultsCriticalReadable  []string
	resultsWarningMachine    []string
	resultCriticalMachine    []string
	resultInformationMachine []string
	resultWithoutTime        []string
)

var lockedGC = make(map[string]map[int]struct{})

// store all context channel that have been canceled
// store all dones with its file, line and id
var contextCancel = make(map[int]struct{})
var contextDone = make(map[string]map[int]int)

// ResultElem declares an interface for a result elem
type ResultElem interface {
	isInvalid() bool
	stringMachine() string
	stringReadable() string
	stringMachineShort() string
	getFile() string
	getLine() int
}

// TraceElementResult is a type to represent an element that is
// part of a found bug
type TraceElementResult struct {
	RoutineID  int
	ResourceID int
	TRequest   int
	ObjType    trace.OperationType
	File       string
	Line       int
}

var blockedGC = make(map[string]map[int]struct{}) // file -> line

// getFile returns the file path of the element
//
// Returns:
//   - string: the file path
func (this TraceElementResult) getFile() string {
	return this.File
}

// getLine returns the  line of the element
//
// Returns:
//   - int: line
func (this TraceElementResult) getLine() int {
	return this.Line
}

// stringMachineShort returns a short machine readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (this TraceElementResult) stringMachineShort() string {
	return fmt.Sprintf("T%s%d%s%s%s%s%s%d", consts.PosSep, this.ResourceID, consts.PosSep, this.ObjType, consts.PosSep, this.File, consts.PosSep, this.Line)
}

// stringMachine returns a machine readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (this TraceElementResult) stringMachine() string {
	return fmt.Sprintf("T%s%d%s%d%s%d%s%s%s%s%s%d", consts.PosSep, this.RoutineID, consts.PosSep, this.ResourceID,
		consts.PosSep, this.TRequest, consts.PosSep, this.ObjType, consts.PosSep, this.File, consts.PosSep, this.Line)
}

// stringReadable returns a human readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (this TraceElementResult) stringReadable() string {
	return fmt.Sprintf("%s:%d@%d", this.File, this.Line, this.TRequest)
}

// isInvalid checks if the result element is not corrupted/empty
//
// Returns:
//   - bool: true if valid, false otherwise
func (this TraceElementResult) isInvalid() bool {
	return this.ObjType == "" || this.Line == -1
}

// Result logs a found bug
//
// Parameter:
//   - level resultLevel: level of the message (critical, warning, ...)
//   - resType ResultType: type of bug that was found
//   - arg1 []ResultElem]: elements directly involved in the bug (e.g. in send on closed the send)
//   - arg2 []ResultElem]: elements indirectly involved in the bug (e.g. in send on closed the close), set to nil if no elem
func Result(level resultLevel, resType helper.ResultType, arg1 []ResultElem, arg2 []ResultElem) {
	if filterInvalidResults(resType, arg1) {
		return
	}

	argType1, argType2 := resType.Args()

	// if flags.FuzzingMode == "Guided" && !resType.IsActual() {
	// 	return
	// }

	foundBug = true

	if resType == helper.ABlocking {
		for _, a := range arg1 {
			file := a.getFile()
			if _, ok := blockedGC[file]; !ok {
				blockedGC[file] = make(map[int]struct{})
			}
			blockedGC[file][a.getLine()] = struct{}{}
		}
	}

	falsePos := "tp"

	if flags.CheckBenign && (resType.IsLeak() || resType.IsBlocking()) {
		falsePositive, err := benign.IsBenign(resType, arg1[0].getFile(), arg1[0].getLine(), blockedGC, contextCancel, contextDone)
		if err != nil {
			log.Errorf("Could not determine if bug is benign: %s", err.Error())
		}
		if falsePositive {
			falsePos = "fp"
		}
	}

	resultReadable := resultTypeMap[resType] + ":" + falsePos + ":\n\t" + argType1 + ": "
	resultMachine := string(resType) + "," + falsePos + ","
	resultMachineShort := string(resType)

	for i, arg := range arg1 {
		if arg.isInvalid() {
			continue
		}
		if i != 0 {
			resultReadable += ";"
			resultMachine += ";"
		}
		resultReadable += arg.stringReadable()
		resultMachine += arg.stringMachine()
		resultMachineShort += arg.stringMachineShort()
	}

	resultReadable += "\n"
	if arg2 != nil || len(arg2) == 0 {
		resultReadable += "\t" + argType2 + ": "
		resultMachine += ","
		for i, arg := range arg2 {
			if arg.isInvalid() {
				continue
			}
			if i != 0 {
				resultReadable += ";"
				resultMachine += ";"
			}
			resultReadable += arg.stringReadable()
			resultMachine += arg.stringMachine()
			resultMachineShort += arg.stringMachineShort()
		}
	}

	resultReadable += "\n"
	resultMachine += "\n"

	switch level {
	case WARNING:
		if !types.Contains(resultWithoutTime, resultMachineShort) {
			resultsWarningReadable = append(resultsWarningReadable, resultReadable)
			resultsWarningMachine = append(resultsWarningMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	case CRITICAL:
		if !types.Contains(resultWithoutTime, resultMachineShort) {
			resultsCriticalReadable = append(resultsCriticalReadable, resultReadable)
			resultCriticalMachine = append(resultCriticalMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	case INFORMATION:
		if !types.Contains(resultWithoutTime, resultMachineShort) {
			resultInformationMachine = append(resultInformationMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	}
}

// AddContext stores all context channel, that have been canceled and stores the
// corresponding data for the done
//
// Parameter:
//   - file string: file of the channel
//   - line int: line of the channel
//   - id int: channel id
func AddContext(file string, line, id int) {
	// cancel
	if strings.HasSuffix(file, "goPatch/src/context/context.go") && line == 565 {
		contextCancel[id] = struct{}{}
		return
	}

	// done
	if _, ok := contextCancel[id]; ok {
		if _, ok := contextDone[file]; !ok {
			contextDone[file] = make(map[int]int)
		}
		contextDone[file][line] = id
	}
}

// Some results are invalid or intentionally not shown. This function returns,
// if the given parameters constitute such a result
//
// Parameter:
//   - resType ResultType: type of bug that was found
//   - arg1 []ResultElem]: elements directly involved in the bug (e.g. in send on closed the send)
//
// Returns:
//   - bool: true if the result is invalid and should be ignored, false otherwise
func filterInvalidResults(resType helper.ResultType, arg1 []ResultElem) bool {
	if resType != helper.RUnknownPanic && resType != helper.RTimeout && len(arg1) == 0 {
		return true
	}

	if control.WasCanceled() {
		return true
	}

	if resType == helper.ABlocking && len(arg1) == 1 && strings.HasSuffix(arg1[0].getFile(), paths.Join(true, false, "src", "testing", "testing.go")) {
		return true
	}

	return false
}

// InitResults sets the output file paths and clears al previous results
//
// Parameter:
//   - outReadable: path to the output file, no output file if empty
//   - outMachine: path to the output file for the reordered trace, no output file if empty
func InitResults(outReadable string, outMachine string) {
	Reset()
	outputReadableFile = outReadable
	outputMachineFile = outMachine
}

// CreateResultFiles writes out the results to the machine and human
// readable result files nad print them to the terminal
//
// Parameter:
//   - noPrint bool: if true, do not print the errors to the terminal
//
// Returns:
//   - int: number of bugs found
//   - error
func CreateResultFiles(noPrint bool) (int, error) {
	counter := 1
	resMachine := ""
	resReadable := "```\n==================== Summary ====================\n\n"

	if !noPrint {
		fmt.Print("==================== Summary ====================\n\n")
	}

	found := false

	if len(resultsCriticalReadable) > 0 {
		found = true
		resReadable += "-------------------- Critical -------------------\n\n"

		if !noPrint {
			fmt.Print("-------------------- Critical -------------------\n\n")
		}

		for _, result := range resultsCriticalReadable {
			resReadable += strconv.Itoa(counter) + " " + result + "\n"

			if !noPrint {
				fmt.Println(strconv.Itoa(counter) + " " + result)
			}

			counter++
		}

		for _, result := range resultCriticalMachine {
			resMachine += result
		}
	}

	if !flags.NoWarning {
		if len(resultsWarningReadable) > 0 {
			found = true
			resReadable += "\n-------------------- Warning --------------------\n\n"
			if !noPrint {
				fmt.Print("\n-------------------- Warning --------------------\n\n")
			}

			for _, result := range resultsWarningReadable {
				resReadable += strconv.Itoa(counter) + " " + result + "\n"

				if !noPrint {
					fmt.Println(strconv.Itoa(counter) + " " + result)
				}

				counter++
			}

			for _, result := range resultsWarningMachine {
				resMachine += result
			}
		}

		for _, result := range resultInformationMachine {
			resMachine += result
		}
	}

	if !found {
		resReadable += "No bugs found" + "\n"

		if !noPrint {
			fmt.Println("No bugs found")
		}
	}

	resReadable += "```"

	// write output readable
	if _, err := os.Stat(outputReadableFile); err == nil {
		if err := os.Remove(outputReadableFile); err != nil {
			return getNumberRes(), err
		}
	}

	fileR, err := os.OpenFile(outputReadableFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return getNumberRes(), err
	}
	defer fileR.Close()

	if _, err := fileR.WriteString(resReadable); err != nil {
		return getNumberRes(), err
	}

	// write output machine
	if _, err := os.Stat(outputMachineFile); err == nil {
		if err := os.Remove(outputMachineFile); err != nil {
			return getNumberRes(), err
		}
	}

	var fileM *os.File
	fileM, err = os.OpenFile(outputMachineFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return getNumberRes(), err
	}
	defer fileM.Close()

	if _, err := fileM.WriteString(resMachine); err != nil {
		return getNumberRes(), err
	}

	return getNumberRes(), nil
}

// getNumberRes returns the number of found bugs
func getNumberRes() int {
	if flags.NoWarning {
		return len(resultCriticalMachine)
	}
	return len(resultCriticalMachine) + len(resultsWarningMachine) + len(resultInformationMachine)
}

// Reset the global values storing the found results
func Reset() {
	resultsWarningReadable = make([]string, 0)
	resultsCriticalReadable = make([]string, 0)
	resultsWarningMachine = make([]string, 0)
	resultCriticalMachine = make([]string, 0)
	resultInformationMachine = make([]string, 0)

	resultWithoutTime = make([]string, 0)

	outputMachineFile = ""
	outputReadableFile = ""

	foundBug = false
}

// GetBugWasFound returns if since the last reset, a bug was found
//
// Returns:
//   - foundBug
func GetBugWasFound() bool {
	return foundBug
}
