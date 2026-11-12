// Copyright (c) 2024 , Mario Occhinegro
//
// File: header.go
// Brief: Functions to add and remove the GOSC header into/from files containing
//    a main function
//
//
// License: BSD-3-Clause

package toolchain

import (
	"bufio"
	"errors"
	"fmt"
	"gosc/analysis/a_base"
	"gosc/fuzzing/f_base"
	"gosc/fuzzing/f_constraints"
	"gosc/utils/flags"
	"gosc/utils/io"
	"gosc/utils/paths"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getBuildArg(replay bool, tracePath string,
	replayTimeout int, record bool, fuzzing int, isMain bool) (buildArg string) {

	buildArg = buildFlagsDefault + " "

	if replay { // replay
		if record {
			buildArg += fmt.Sprintf("-goscfuzzing -goscpath=%s -gosctimeout=%d", tracePath, flags.Timeout)
		} else {
			buildArg += fmt.Sprintf("-goscreplay -goscpath=%s -gosctimeout=%d", tracePath, replayTimeout)
		}
	} else if fuzzing > 0 {
		buildArg += fmt.Sprintf("-goscfuzzing -goscpath=%s -gosctimeout=%d", tracePath, replayTimeout)
	} else { // recording
		buildArg += fmt.Sprintf("-gosctrace  -gosctimeout=%d", replayTimeout)
	}

	if isMain {
		buildArg += " -goscmain"
	}

	if flags.RecordPlus {
		buildArg += " -goscrecordplus"
	}

	// buildArg += "'"

	return
}

// ============================================
// MARK: Main
// ============================================

// Insert the header into a main function and return the build parameters
//
// Parameter:
//   - fileName string: path to the main file
//   - replay bool: true for replay, false for only recording
//   - replayNumber string: id of the trace to replay
//   - replayTimeout int: replay for timeout
//   - record bool: if both replay and record are set, the replay is rerecorded
//   - fuzzing int: fuzzing run, if no fuzzing: -1, for initial run: 0
//
// Returns:
//   - string: build parameters
//   - error
func buildArgsMain(fileName string, replay bool, replayNumber string,
	record bool, fuzzing int) (string, error) {
	if fileName == "" {
		return "", errors.New("Please provide a file  name")
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", fmt.Errorf("File %s does not exist", fileName)
	}

	exists, err := mainMethodExists(fileName)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("Main Method not found in file")
	}

	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("Could not open main file to add header")
	}
	defer file.Close()

	fmt.Println("FileName: ", fileName)
	fmt.Println("TestName: Main")

	replayPath := ""
	if replay || fuzzing > 0 {
		if replayNumber != "" {
			if fuzzing <= 0 {
				replayPath = paths.NameRewrittenConsts + "_" + replayNumber
			} else {
				replayPath = filepath.Join(paths.FuzzingTraces, fmt.Sprintf("%s_%d", paths.NameFuzzingTraces, fuzzing))
			}
		} else if flags.TracePath != "" {
			replayPath = flags.TracePath
		} else {
			replayPath = paths.NameRewrittenConsts
		}

		if flags.Mode == flags.ModeReplay {
			trConst := f_constraints.TraceToConstraint(&a_base.MainTrace)
			f_base.WriteConstraint(replayPath, &a_base.MainTrace, trConst, nil)
		}
	}

	return getBuildArg(replay, replayPath, flags.Timeout, record, fuzzing, true), nil
}

// Check if there is a main function in the given file
//
// Parameter:
//   - fileName string: name of the file
//
// Returns
//
//   - bool: true if the file contains a main function, false otherwise
//   - error
func mainMethodExists(fileName string) (bool, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return false, err
	}
	defer file.Close()

	regexStr := "func main\\(\\) {"
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if regex.MatchString(line) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// ============================================
// MARK: Test
// ============================================

// Add the header into a unit test and return the build parameter
//
// Parameter:
//   - fileName string: path to the file containing the the test
//   - testName string: name of the test
//   - replay bool: true for replay, false for only recording
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - replayInfo string: path of the fuzzing trace or if the replay trace
//   - record bool: true to rerecord the leaks
//
// Returns:
//   - string: build args
//   - error
func buildArgsUnit(fileName, testName string, replay bool, fuzzing int, replayInfo string, record bool) (string, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", fmt.Errorf("file %s does not exist", fileName)
	}

	testExists, err := testExists(fileName, testName)
	if err != nil {
		return "", err
	}

	if !testExists {
		return "", errors.New("Test Method not found in file")
	}

	fmt.Println("FileName: ", fileName)
	fmt.Println("TestName: ", testName)

	replayPath := ""
	if replay || fuzzing > 0 {
		if replayInfo != "" {
			if fuzzing <= 0 {
				replayPath = paths.NameRewrittenConsts + "_" + replayInfo
			} else {
				replayPath = filepath.Join(paths.FuzzingTraces, fmt.Sprintf("%s_%d", paths.NameFuzzingTraces, fuzzing))
			}
		} else if flags.TracePath != "" {
			replayPath = flags.TracePath
		} else {
			replayPath = paths.NameRewrittenConsts
		}

		if flags.Mode == flags.ModeReplay {
			tr, _, _, _ := io.CreateTraceFromFiles(replayPath, false)
			trConst := f_constraints.TraceToConstraint(tr)
			f_base.WriteConstraint(replayPath, tr, nil, trConst)
		}
	}

	return getBuildArg(replay, replayPath, flags.Timeout, record, fuzzing, false), nil
}

// Check if a test exists
//
// Parameter:
//   - fileName string: path to the file
//   - testName string: name of the test
//
// Returns:
//   - bool: true if the test exists, false otherwise
//   - error
func testExists(fileName string, testName string) (bool, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "func "+testName) && strings.Contains(line, "testing.T") {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}
