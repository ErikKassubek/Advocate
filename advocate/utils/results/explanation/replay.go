// Copyright (c) 2024 Erik Kassubek
//
// File: replay.go
// Brief: Read the info about the rewrite and replay of the bug
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package explanation

import (
	"advocate/utils/consts"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/paths"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Get the text for the rewrite and replay info for the explanation
//
// Parameter:
//   - bugType string: bug code
//   - code map[string]string: replay output codes
//   - index int: index of the bug that is explained
//
// Returns:
//   - map[bugKeys]string: Information for the rewrite/replay part of the explanation file
func getRewriteInfo(bugType helper.ResultType, codes map[string]string, index string) map[bugKeys]string {
	res := make(map[bugKeys]string)

	rewType := getRewriteType(bugType)

	res[desc] = ""
	res[exitCode] = ""
	res[exitCodeDesc] = ""
	res[replaySuc] = "was not possible"

	var err error

	switch rewType {
	case Actual:
		res[desc] += "The bug is an actual bug. Therefore no rewrite is possible."
		codes[fmt.Sprint(index)] = "fail"
	case Potential:
		res[desc] += "The bug is a potential bug.\n"
		res[desc] += consts.TheAnalyzerHasTriedToRewriteTheTraceInSuchAWay + " "
		res[desc] += "that the bug will be triggered when replaying the trace."
	}
	res[exitCode], res[exitCodeDesc], res[replaySuc], err = getReplayInfo(codes, index)

	if err != nil {
		log.Error("Error getting replay info: ", err)
	}

	return res

}

// From the bug code get wether the bug was actual, possible or a (possible) leak
//
// Parameter:
//   - bugCode string: the bug code
//
// Returns:
//   - string: Actual, Possible, Leak or Pos Leak
func getRewriteType(bugCode helper.ResultType) BugClass {
	switch string(bugCode)[:1] {
	case "A":
		return Actual
	case "P":
		return Potential
	case "L":
		return Actual
	}
	return UnknownBugClass
}

// Get the output codes from the output.log file
//
// Returns:
//
//			map[string]stringL exit codes
//	     []string: waiting elements
func getOutputInfo() (map[string]string, bool) {
	if _, err := os.Stat(paths.ResultOutput); os.IsNotExist(err) {
		res := "No replay info available. Output.log does not exist."
		return map[string]string{"AdvocateFailExplanationInfo": res, "AdvocateFailResplaySucInfo": "information not available"}, false
	}

	// read the output file
	content, err := os.ReadFile(paths.ResultOutput)
	if err != nil {
		res := "No replay info available. Could not read output.log file"
		return map[string]string{"AdvocateFailExplanationInfo": res, "AdvocateFailResplaySucInfo": "information not available"}, false
	}

	lines := strings.Split(string(content), "\n")

	replayPos := make(map[string]bool)
	replayCode := make(map[string]string)
	bugrepPrefix := "Bugreport info: "
	replayReadPrefix := "Reading trace from " + consts.RewrittenConsts + "_"
	exitCodePrefix := "Exit Replay with code"
	waitingPrefix := "WAITING"

	lastReplayIndex := ""
	lastReplayIndexInfoFound := true
	blockedByReplay := false

	for _, line := range lines {
		if strings.HasPrefix(line, bugrepPrefix) {
			line = strings.TrimPrefix(line, bugrepPrefix)
			lineSplit := strings.Split(line, ",")
			if len(lineSplit) == 2 {
				index := lineSplit[0]
				if lineSplit[1] == "suc" {
					replayPos[index] = true
				} else {
					replayPos[index] = false
				}
				if lineSplit[1] == "double" {
					replayCode[index] = "double"
				}
				if lineSplit[1] == "fail" {
					replayCode[index] = "fail"
				}
			}
		} else if strings.HasPrefix(line, replayReadPrefix) {
			if !lastReplayIndexInfoFound {
				replayCode[lastReplayIndex] = "panic"
			}
			lastReplayIndex = strings.TrimPrefix(line, replayReadPrefix)
			// if !strings.Contains(lastReplayIndex, "_") {
			// 	lastReplayIndex = "0_" + lastReplayIndex
			// }
			lastReplayIndexInfoFound = false
		} else if strings.HasPrefix(line, exitCodePrefix) {
			line = strings.TrimPrefix(line, exitCodePrefix)
			line = strings.TrimSpace(line)
			replayCode[lastReplayIndex] = strings.Split(line, " ")[0]
			lastReplayIndexInfoFound = true
		} else if strings.HasPrefix(line, waitingPrefix) {
			line = strings.TrimPrefix(line, waitingPrefix)
			blockedByReplay = true
		}
	}

	return replayCode, blockedByReplay
}

// Get the text for the replay info for the explanation
//
// Parameter:
//   - code map[string]string: replay output codes
//   - index int: index of the bug that is explained
//
// Returns:
//   - string: exit code
//   - string: replay result explanation
//   - string: replay success info
//   - error
func getReplayInfo(codes map[string]string, index string) (string, string, string, error) {
	if _, ok := codes["AdvocateFailExplanationInfo"]; ok {
		return "", codes["AdvocateFailExplanationInfo"], codes["AdvocateFailResplaySucInfo"], fmt.Errorf("Could not read output file")
	}

	exitCode := codes[index]
	replaySuc := "failed"
	if exitCode == "double" {
		replaySuc = "was already performed for this bug in another test"
		return "double", "", replaySuc, nil
	}
	if exitCode == "fail" {
		return "fail", exitCodeExplanation["fail"], "was not run", nil
	}
	if exitCode == "panic" {
		return "panic", exitCodeExplanation["panic"], "was terminated unexpectedly", nil
	}

	exitCodeInt, err := strconv.Atoi(exitCode)
	if err != nil {
		return "fail", exitCodeExplanation["fail"], "was not run", nil
	}
	if exitCodeInt == 0 {
		replaySuc = "ended without confirming the bug"
	} else if exitCodeInt >= 20 {
		replaySuc = consts.ConfirmedTheBug
	}

	return exitCode, exitCodeExplanation[exitCode], replaySuc, nil
}
