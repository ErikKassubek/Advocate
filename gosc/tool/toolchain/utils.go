// Copyright (c) 2024 
//
// File: headerUnitTests.go
// Brief: Functions to add and remove the GOSC header into file containing
//    unit tests
//
//
// License: BSD-3-Clause

package toolchain

import (
	"gosc/utils/paths"
	"gosc/utils/results/buginfo"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const buildFlagsDefault = "-gcflags=all=-N -l"

// extractTraceNumber extracts the numeric part from a trace directory name
//
// Parameter:
//   - trace string: path to the rewritten trace folder
//
// Returns:
//   - string: trace number
//   - string: bug string of the rewritten trace
func extractTraceNumber(trace string) (string, string) {
	traceNumber := ""
	bugString := ""

	// read trace number
	parts := strings.Split(trace, paths.NameRewrittenConsts+"_")
	if len(parts) > 1 {
		traceNumber = parts[1]
	}

	// read bug string
	rewrittenInfoPath := filepath.Join(trace, paths.NameRewrittenInfo)
	file, err := os.Open(rewrittenInfoPath)
	if err != nil {
		return traceNumber, bugString
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		elems := strings.Split(line, "#")
		if len(elems) > 1 {
			bugString = elems[1]
		}
	}

	return traceNumber, bugString
}

// For a given run, check if it was terminated by a timeout
//
// Parameter:
//   - output (string): path to the output.log file
//
// Returns:
//   - true if an timeout occurred, false otherwise
func checkForTimeout(output string) bool {
	outFile, err := os.Open(output)
	if err != nil {
		return false
	}
	defer outFile.Close()

	content, err := io.ReadAll(outFile)
	if err != nil {
		return false
	}

	if strings.Contains(string(content), "panic: test timed out after") {
		return true
	}

	return false
}

// readReplayResult reads the results and returns whether the replay was successful
//
// Parameter:
//   - output string: path to the output file
//
// Returns:
//   - bool: true if the bug was confirmed, false if the replay failed
func wasReplaySuc(output string) bool {
	outFile, err := os.Open(output)
	if err != nil {
		return false
	}
	defer outFile.Close()

	content, err := io.ReadAll(outFile)
	if err != nil {
		return false
	}

	pref := "Exit Replay with code  "
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, pref) {
			line = strings.TrimPrefix(line, pref)
			exitCode, err := strconv.Atoi(strings.Split(line, " ")[0])
			if err != nil {
				return false
			}

			return exitCode >= buginfo.MinExitCodeSuc
		}
	}

	return false
}
