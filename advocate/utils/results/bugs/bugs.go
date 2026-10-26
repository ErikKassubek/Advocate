// Copyright (c) 2024 Erik Kassubek
//
// File: bugs.go
// Brief: Operations for handling found bugs
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package bugs

import (
	"advocate/analysis/a_base"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"errors"
	"sort"
	"strconv"
	"strings"
)

// BugElementSelectCase is a type to store a specific id in a select
//
// Parameter:
//   - ID int: id of the involved channel
//   - ObjType string: object type
//   - Index int: internal index of the select int the case
type BugElementSelectCase struct {
	ID      int
	ObjType string
	Index   int
}

// GetBugElementSelectCase builds a BugElementSelectCase from a string
//
// Parameter:
//   - arg string: the string representing the case
//
// Returns:
//   - BugElementSelectCase: the bug select as a BugElementSelectCase
//   - error
func GetBugElementSelectCase(arg string) (BugElementSelectCase, error) {
	elems := strings.Split(arg, ":")
	id, err := strconv.Atoi(elems[1])
	if err != nil {
		return BugElementSelectCase{}, err
	}
	objType := elems[2]
	index, err := strconv.Atoi(elems[3])
	if err != nil {
		return BugElementSelectCase{}, err
	}
	return BugElementSelectCase{id, objType, index}, nil
}

// Bug is a type to describe and store a found bug
//
// Parameter:
//   - Type ResultType: The type of the bug
//   - FalsePos bool: True if the bug is most likely a false positive
//   - TraceElement1 []trace.TraceElement: first list of trace element involved in the bug
//     normally the elements that actually cause the bug, e.g. for send on close the send
//   - TraceElement2 []trace.TraceElement: second list of trace element involved in the bug
//     normally the elements indirectly involved or elements to solve the bug (possible partner),
//     e.g. for send on close the close
type Bug struct {
	Type          helper.ResultType
	FalsePos      bool
	TraceElement1 []trace.Element
	// TraceElement1Sel []BugElementSelectCase
	TraceElement2 []trace.Element
}

// GetBugString Convert the bug to a unique string. Mostly used internally
//
// Returns:
//   - string: The bug as a string
func (this Bug) GetBugString() string {
	paths := make([]string, 0)

	for _, t := range this.TraceElement1 {
		paths = append(paths, t.Pos().String())
	}
	for _, t := range this.TraceElement2 {
		paths = append(paths, t.Pos().String())
	}

	sort.Strings(paths)

	res := string(this.Type)
	for _, path := range paths {
		res += path
	}
	return res
}

// ToString convert the bug to a string. Mostly used for output
//
// Returns:
//   - string: The bug as a string
func (this Bug) ToString() string {
	typeStr := ""
	arg1Str, arg2Str := this.Type.Args()

	switch this.Type {
	case helper.RUnknownPanic:
		typeStr = "Unknown Panic:"
	case helper.RTimeout:
		typeStr = "Timeout"
	case helper.RReplayFail:
		typeStr = "ReplayFail"
	case helper.ASendOnClosed:
		typeStr = "Actual Send on Closed Channel:"
	case helper.ACloseOnClosed:
		typeStr = "Actual Close on Closed Channel:"
	case helper.ACloseOnNilChannel:
		typeStr = "Actual close on nil channel:"
	case helper.AConcurrentRecv:
		typeStr = "Concurrent Receive:"
	case helper.ABlocking:
		typeStr = "Blocking routine:"
	case helper.ADeadlock:
		typeStr = "Deadlock:"
	case helper.ANegWG:
		typeStr = "Actual negative Wait Group:"
	case helper.AUnlockOfNotLockedMutex:
		typeStr = "Actual unlock of not locked mutex:"
	case helper.PSendOnClosed:
		typeStr = "Possible send on closed channel:"
	case helper.PNegWG:
		typeStr = "Possible negative waitgroup counter:"
	case helper.PUnlockBeforeLock:
		typeStr = "Possible unlock of a not locked mutex:"
	case helper.PCyclicDeadlock:
		typeStr = "Possible cyclic deadlock:"
	case helper.PMixedDeadlock:
		typeStr = "Possible Mixed Deadlock:"
	case helper.LUnknown:
		typeStr = "Leak on routine"
	case helper.LChan:
		typeStr = "Leak on a channel:"
	case helper.LNilChan:
		typeStr = "Leak on nil channel:"
	case helper.LSelect:
		typeStr = "Leak on select:"
	case helper.LMutex:
		typeStr = "Leak on mutex:"
	case helper.LWaitGroup:
		typeStr = "Leak on wait group:"
	case helper.LCond:
		typeStr = "Leak on conditional variable:"
	default:
		log.Error("Unknown bug type in toString: " + string(this.Type))
		return ""
	}

	res := typeStr + "\n\t" + arg1Str
	for i, elem := range this.TraceElement1 {
		if i != 0 {
			res += ";"
		}
		res += elem.String()
	}

	if arg2Str != "" {
		res += "\n\t" + arg2Str

		if len(this.TraceElement2) == 0 {
			res += "-"
		}

		for i, elem := range this.TraceElement2 {
			if i != 0 {
				res += ";"
			}
			res += elem.String()
		}
	}

	return res
}

// Println prints the bug
func (this Bug) Println() {
	println(this.ToString())
}

// ProcessBug processes the bug that was selected from the analysis results
//
// Parameter:
//   - bugStr: The bug that was selected
//
// Returns:
//   - bool: true, if the bug was not a possible, but a actually occurring bug
//     Bug: The bug that was selected
//     error: An error if the bug could not be processed
func ProcessBug(bugStr string) (bool, Bug, error) {
	bug := Bug{}

	bugSplit := strings.Split(bugStr, ",")
	if len(bugSplit) != 4 && len(bugSplit) != 3 {
		return false, bug, errors.New("Could not split bug: " + bugStr)
	}

	bugType := bugSplit[0]

	containsArg1 := true
	containsArg2 := true
	actual := false

	switch bugType {
	case "R01":
		bug.Type = helper.RUnknownPanic
		actual = true
	case "R02":
		bug.Type = helper.RTimeout
		actual = true
	case "R03":
		bug.Type = helper.RReplayFail
		actual = true
	case "A01":
		bug.Type = helper.ASendOnClosed
		actual = true
	case "A02":
		bug.Type = helper.ACloseOnClosed
		actual = true
	case "A03":
		bug.Type = helper.ACloseOnNilChannel
		actual = true
	case "A04":
		bug.Type = helper.ANegWG
		actual = true
	case "A05":
		bug.Type = helper.AUnlockOfNotLockedMutex
		actual = true
	case "A06":
		bug.Type = helper.ABlocking
		actual = true
	case "A07":
		bug.Type = helper.ADeadlock
		actual = true
	case "A08":
		bug.Type = helper.AConcurrentRecv
		actual = true
	case "A09":
		bug.Type = helper.AMixedDeadlock
	case "P01":
		bug.Type = helper.PSendOnClosed
	case "P02":
		bug.Type = helper.PNegWG
	case "P03":
		bug.Type = helper.PUnlockBeforeLock
	case "P04":
		bug.Type = helper.PCyclicDeadlock
		containsArg2 = true
	case "P05":
		bug.Type = helper.PMixedDeadlock
	case "L00":
		containsArg1 = false
		bug.Type = helper.LUnknown
	case "L01":
		bug.Type = helper.LChan
	case "L02":
		bug.Type = helper.LNilChan
		containsArg2 = false
	case "L03":
		bug.Type = helper.LSelect
	case "L04":
		bug.Type = helper.LMutex
		containsArg2 = false
	case "L05":
		bug.Type = helper.LWaitGroup
		containsArg2 = false
	case "L06":
		bug.Type = helper.LChan
	// case "S00":
	// 	bug.Type = SNotExecutedWithPartner
	// 	containsArg2 = true
	default:
		return actual, bug, errors.New("Unknown bug type in process bug: " + bugStr)
	}

	if !containsArg1 {
		return actual, bug, nil
	}

	bug.FalsePos = (bugSplit[1] == "fp")

	bugArg1 := bugSplit[2]
	bugArg2 := ""
	if containsArg2 && len(bugSplit) == 4 {
		bugArg2 = bugSplit[3]
	}

	bug.TraceElement1 = make([]trace.Element, 0)
	// bug.TraceElement1Sel = make([]BugElementSelectCase, 0)

	for _, bugArg := range strings.Split(bugArg1, ";") {
		if strings.TrimSpace(bugArg) == "" {
			continue
		}

		if strings.HasPrefix(bugArg, "T") {
			elem, err := a_base.GetTraceElementFromBugArg(bugArg)
			if err != nil {
				return actual, bug, err
			}
			bug.TraceElement1 = append(bug.TraceElement1, elem)
		}
		// else if strings.HasPrefix(bugArg, "S") {
		// 	elem, err := GetBugElementSelectCase(bugArg)
		// 	if err != nil {
		// 		println("Could not read: " + bugArg + " from results")
		// 		return actual, bug, err
		// 	}
		// 	// bug.TraceElement1Sel = append(bug.TraceElement1Sel, elem)
		// }
	}

	bug.TraceElement2 = make([]trace.Element, 0)

	if !containsArg2 {
		return actual, bug, nil
	}

	for _, bugArg := range strings.Split(bugArg2, ";") {
		if strings.TrimSpace(bugArg) == "" {
			continue
		}

		if bugArg[0] == 'T' {
			elem, err := a_base.GetTraceElementFromBugArg(bugArg)
			if err != nil {
				return actual, bug, err
			}

			bug.TraceElement2 = append(bug.TraceElement2, elem)
		}
	}

	return actual, bug, nil
}
