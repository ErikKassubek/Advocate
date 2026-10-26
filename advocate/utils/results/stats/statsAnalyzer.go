// Copyright (c) 2024 Erik Kassubek
//
// File: statsAnalyzer.go
// Brief: Collect stats about the analysis and the replay
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package stats

import (
	consts "advocate/utils/consts"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/paths"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// getNewDataMap provides a new map to store the analyzer stats.
// It has the form bugTypeID -> counter
//
// Returns:
//   - map[helper.ResultType]int: The new map
func getNewDataMap() map[helper.ResultType]int {
	keys := helper.ResultTypes

	m := make(map[helper.ResultType]int)
	for _, key := range keys {
		m[key] = 0
	}

	return m
}

// getNewDataMapMap provides a map used for collecting statistics of the analysis.
// The fields are detected, replayWritten, replaySuccessful, unexpectedPanic.
// Each field contains a data map as created by getNewDataMap()
//
// Returns:
//   - map[string]map[helper.ResultType]int: The map
func getNewDataMapMap() map[statsType]map[helper.ResultType]int {
	return map[statsType]map[helper.ResultType]int{
		detected:         getNewDataMap(),
		replayWritten:    getNewDataMap(),
		replaySuccessful: getNewDataMap(),
		unexpectedPanic:  getNewDataMap(),
	}
}

// Parse the analyzer and replay output to collect the corresponding information
//
// Parameter:
//   - fuzzing int: number of fuzzing run, -1 for not fuzzing
//
// Returns:
//   - map[statsType]map[helper.ResultType]int: map with total information
//   - map[statsType]map[helper.ResultType]int: map with unique information
//   - error
func statsAnalyzer(fuzzing int) (map[statsType]map[helper.ResultType]int, map[statsType]map[helper.ResultType]int, error) {
	// reset foundBugs
	foundBugs := make(map[string]processedBug)

	resUnique := getNewDataMapMap()

	resTotal := getNewDataMapMap()

	bugs := filepath.Join(paths.CurrentResult, "bugs")
	_, err := os.Stat(bugs)
	if os.IsNotExist(err) {
		return resUnique, nil, nil
	}

	err = filepath.Walk(bugs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if fuzzing == -1 {
			if strings.HasPrefix(info.Name(), "bug_") ||
				strings.HasPrefix(info.Name(), "diagnostics_") ||
				strings.HasPrefix(info.Name(), "replay_") ||
				strings.HasPrefix(info.Name(), "leak_") {
				err := processBugFile(path, foundBugs, resTotal, resUnique)
				if err != nil {
					log.Error(err)
				}
			}
		} else {
			if strings.HasPrefix(info.Name(), "bug_"+strconv.Itoa(fuzzing)+"_") ||
				strings.HasPrefix(info.Name(), "replay_"+strconv.Itoa(fuzzing)+"_") ||
				strings.HasPrefix(info.Name(), "potential_"+strconv.Itoa(fuzzing)+"_") ||
				strings.HasPrefix(info.Name(), "diagnostics_"+strconv.Itoa(fuzzing)+"_") ||
				strings.HasPrefix(info.Name(), "leak_"+strconv.Itoa(fuzzing)+"_") {
				err := processBugFile(path, foundBugs, resTotal, resUnique)
				if err != nil {
					log.Error(err)
				}
			}
		}

		for _, bug := range foundBugs {
			resUnique[detected][bug.bugType]++

			if bug.replayWritten {
				resUnique[replayWritten][bug.bugType]++
			}

			if bug.replaySuc {
				resUnique[replaySuccessful][bug.bugType]++
			}

			if bug.falsePos {
				resUnique[falsePositive][bug.bugType]++
			}

		}

		return nil
	})

	return resTotal, resUnique, err
}

// Store a bug that has been processed in the statistics.
// Used to count number of unique bugs
// Properties:
//
//   - paths []string: list of paths to each element involved in the bug
//   - bugType string: ID of the bug type
//   - replayWritten bool: true if a replay trace was created for the bug
//   - replaySuc bool: true if the replay of the bug was successful
//   - falsePos bool: true if the bug is likely a false positive
type processedBug struct {
	paths         []string
	bugType       helper.ResultType
	replayWritten bool
	replaySuc     bool
	falsePos      bool
}

// Get a string representation of a bug
//
// Returns:
//   - string: string representation of the bug
func (this *processedBug) getKey() string {
	res := string(this.bugType)
	for _, path := range this.paths {
		res += path
	}
	return res
}

// Parse a bug file to get the information
//
// Parameter:
//   - filePath string: path to the bug file
//   - resTotal map[string]map[string]int: total results
//   - resUnique map[string]map[string]int: unique results
//
// Returns:
//   - error
func processBugFile(filePath string, foundBugs map[string]processedBug,
	resTotal map[statsType]map[helper.ResultType]int, resUnique map[statsType]map[helper.ResultType]int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var bugType helper.ResultType

	bug := processedBug{}
	bug.paths = make([]string, 0)

	// read the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// get detected bug
		if strings.HasPrefix(line, "# ") {
			textSplit := strings.Split(line, ": ")
			if len(textSplit) != 2 {
				continue
			}

			textSplit2 := strings.Split(textSplit[1], " - ")

			bugType = helper.ResultType(textSplit2[0])
			line = textSplit2[1]
			bug.bugType = bugType
		} else if strings.HasPrefix(line, "-> ") { // get paths
			bug.paths = append(bug.paths, strings.TrimPrefix(line, "-> "))
		} else if strings.Contains(line, "The analyzer found a way to resolve the leak") {
			bug.replayWritten = true
		} else if strings.Contains(line, consts.TheAnalyzerHasTriedToRewriteTheTraceInSuchAWay) {
			bug.replayWritten = true
		} else if strings.Contains(line, consts.TheBugIsLikelyAFalsePositive) {
			bug.falsePos = true
		} else if strings.HasPrefix(line, consts.ItExitedWithTheFollowingCode) {
			code := strings.TrimPrefix(line, consts.ItExitedWithTheFollowingCode)

			num, err := strconv.Atoi(code)
			if err != nil {
				num = -1
			}

			if num == 3 {
				(resUnique)[unexpectedPanic][bugType]++
				if resTotal != nil {
					(resTotal)[unexpectedPanic][bugType]++
				}
			}

			if num >= 20 {
				bug.replaySuc = true
			}
		}
	}

	if bug.bugType == "" {
		return fmt.Errorf("Invalid bug file")
	}

	if resTotal != nil {
		(resTotal)[detected][bugType]++

		if bug.replayWritten {
			(resTotal)[replayWritten][bugType]++
		}

		if bug.replaySuc {
			(resTotal)[replaySuccessful][bugType]++
		}

		if bug.falsePos {
			(resTotal)[falsePositive][bugType]++
		}
	}

	key := bug.getKey()
	if b, ok := foundBugs[key]; ok {
		if bug.replaySuc {
			b.replaySuc = true
		}
		if bug.replayWritten {
			b.replayWritten = true
		}
		foundBugs[key] = b
	} else {
		foundBugs[key] = bug
	}

	return nil
}
