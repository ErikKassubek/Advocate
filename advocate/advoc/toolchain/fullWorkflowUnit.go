// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: runFullWorkflowMain.go
// Brief: Function to run the whole ADVOCATE workflow, including running,
//    analysis and replay on all unit tests of a program
//
// Author: Erik Kassubek, Mario Occhinegro
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/analysis/a_base"
	"advocate/utils/command"
	"advocate/utils/consts"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/complete"
	"advocate/utils/results/results"
	"advocate/utils/results/stats"
	"advocate/utils/timer"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Run ADVOCATE for all given unit tests
//
// Parameter:
//   - pathToAdvocate string: pathToAdvocate
//   - runRecord bool: run the recording. If set to false, but runAnalysis or runReplay is
//     set the trace at tracePath is used
//   - runAnalysis bool: run the analysis on a path
//   - runReplay bool: run replay, if runAnalysis is true, those replays are used
//   - pathToTest string: path to the test file, should be set if exec is set
//   - also runs the tests once without any recoding/replay to get a base value
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - fuzzingTrace string: path to the fuzzing trace path. If not used path (GFuzz or Flow), opr not fuzzing, set to empty string
//   - firstRun bool: this is the first run, only set to false for fuzzing (except for the first fuzzing)
//   - cont bool: continue an already started run
//   - onlyRecord bool: if true, run only the recording without any analysis
//
// Returns:
//   - int: TraceID
//   - int: number results
//   - error
func runWorkflowUnit(runRecord, runAnalysis, runReplay bool,
	pathToTest string, fuzzing int, fuzzingTrace string,
	firstRun bool, fileNumber,
	testNumber int) (int, int, error) {
	// Validate required inputs
	if paths.Advocate == "" {
		return 0, 0, errors.New("Path to advocate is empty")
	}
	if paths.Prog == "" {
		return 0, 0, errors.New("Directory is empty")
	}

	isFuzzing := (fuzzing != -1)

	// Change to the directory
	if err := os.Chdir(paths.Prog); err != nil {
		return 0, 0, fmt.Errorf("Failed to change directory: %v", paths.Prog)
	}

	if firstRun && !flags.Continue {
		if !flags.SkipExisting {
			os.RemoveAll("advocateResult")
		}

		if info, _ := os.Stat("advocateResult"); info == nil {
			if err := os.MkdirAll("advocateResult", os.ModePerm); err != nil {
				return 0, 0, fmt.Errorf("Failed to create advocateResult directory: %v", err)
			}
		}

		// Remove possibly leftover traces from unexpected aborts that could interfere with replay
		// RemoveTraces(dir)
		removeLogs(paths.Prog)
	}

	// Find all _test.go files in the directory
	testFiles, _, totalFiles, err := FindTestFiles(paths.Prog, flags.Continue && flags.ExecName == "")
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to find test files: %v", err)
	}

	attemptedTests, skippedTests, currentFile := 0, 0, fileNumber

	// resultPath := filepath.Join(dir, "advocateResult")
	var numberResults int

	ranTest := false
	// Process each test file
	for _, file := range testFiles {
		if pathToTest != "" && pathToTest != file {
			continue
		}

		if flags.ExecName == "" {
			log.Progressf("Progress: %d/%d", currentFile, totalFiles)
			log.Progressf("Processing file: %s", file)
		}

		packagePath := filepath.Dir(file)
		testFunctions, err := FindTestFunctions(file)
		if err != nil {
			log.Infof("Could not find test functions in %s: %v", file, err)
			continue
		}

		for _, testFunc := range testFunctions {
			if flags.ExecName != "" && flags.ExecName != testFunc {
				continue
			}

			for control.WasCanceledRAM() {
				log.Error("Wait RAM")
				time.Sleep(6 * time.Second)
			}

			a_base.Clear()
			// f_base.Clear()

			if !isFuzzing {
				timer.ResetTest()
				timer.Start(timer.TotalTest)
			}

			ranTest = true

			attemptedTests++
			packageName := filepath.Base(packagePath)
			fileName := filepath.Base(file)

			if fuzzing == -1 {
				log.Progressf("Running test %s in package %s in file %s", testFunc, packageName, file)
			}

			adjustedPackagePath := strings.TrimPrefix(packagePath, paths.Prog)
			if !strings.HasSuffix(adjustedPackagePath, string(filepath.Separator)) {
				adjustedPackagePath = adjustedPackagePath + string(filepath.Separator)
			}
			fileNameWithoutEnding := strings.TrimSuffix(fileName, ".go")
			directoryName := paths.SetCurrentResult(currentFile, testNumber, fileNameWithoutEnding, testFunc)
			// if flags.Continue && fileNumber != 0 {
			// 	directoryName = paths.SetCurrentResult(fileNumber, testNumber, fileNameWithoutEnding, testFunc)
			// }

			if fuzzing < 1 {
				log.Info("Create ", directoryName)
				if err := os.MkdirAll(directoryName, os.ModePerm); err != nil {
					log.Errorf("Failed to create directory %s: %v", directoryName, err)
					if !isFuzzing {
						timer.Stop(timer.TotalTest)
					}
					// continue
				}
			}

			// Execute full workflow
			nrReplay, anaPassed, err := unitTestFullWorkflow(
				runRecord, runAnalysis, runReplay, testFunc, adjustedPackagePath, file, fuzzing,
				fuzzingTrace)

			timer.UpdateTimeFileDetail(testFunc, nrReplay)

			if !isFuzzing {
				timer.ResetTest()
				timer.UpdateTimeFileOverview(testFunc)
			}

			// Move logs and results to the appropriate directory
			total := fuzzing != -1
			collect(paths.Prog, packagePath, paths.CurrentResult, total, true)

			if err != nil {
				log.Error(err.Error())
				skippedTests++
			}

			if anaPassed {
				numberResults += generateBugReports(movedTraces, fuzzing)
			}

			if flags.CreateStatistics {
				stats.CreateStats(testFunc, movedTraces, fuzzing)
			}

			if flags.DeleteTraces && !flags.CreateStatistics {
				RemoveTraces(paths.Prog)
			}

			if total {
				removeLogs(paths.Prog)
			}

			if !isFuzzing {
				timer.Stop(timer.TotalTest)
			} else {
				break
			}
		}

		if isFuzzing && ranTest {
			break
		}

		currentFile++
	}

	if flags.ExecName != "" && !ranTest {
		return 0, 0, fmt.Errorf("could not find test function %s", flags.ExecName)
	}

	// Check for untriggered selects
	if flags.NotExecuted && flags.ExecName != "" {
		err := complete.Check(filepath.Join(paths.Prog, "advocateResult"), paths.Prog)
		if err != nil {
			log.Error("Could not run check for untriggered select and not executed progs: ", err.Error())
		}
	}

	// Output test summary
	if flags.ExecName == "" {
		log.Info("Finished run for all tests")
		log.Infof("Attempted tests: %d", attemptedTests)
		if attemptedTests == 0 {
			log.Errorf("Could not find any tests")
		} else {
			log.Infof("Skipped tests: %d", skippedTests)
		}
	} else {
		log.Infof("Finished run for %s", flags.ExecName)
	}

	return movedTraces, numberResults, nil
}

// FindTestFiles finds all _test.go files in the specified directory
//
// Parameter:
//   - dir string: folder to search in
//   - cont bool: continue testing
//
// Returns:
//   - []string: found files
//   - int: min file num, only if cont, otherwise 0
//   - int: total number of files
//   - error
func FindTestFiles(dir string, cont bool) ([]string, int, int, error) {
	var testFiles []string

	alreadyProcessed, maxFileNum := make(map[string]struct{}), 0
	var err error

	if flags.Continue {
		alreadyProcessed, maxFileNum, err = getFilesInResult(dir)
		if err != nil {
			log.Error(err)
			return testFiles, 0, 0, err
		}
	}

	totalNumFiles := 0
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if strings.HasSuffix(name, "_test.go") {
			totalNumFiles++
			if _, ok := alreadyProcessed[name]; !cont || !ok {
				testFiles = append(testFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return testFiles, maxFileNum, totalNumFiles, err
}

// getFilesInResult finds all test files that have already been run and therefore
// have an entry in the results. Clear up incomplete result folders
//
// Parameter:
//   - dir string: path to the directory containing the test files
//
// Returns:
//   - map[string]struct{}: map containing already processed test files
//   - int: total number of files
//   - error
func getFilesInResult(dir string) (map[string]struct{}, int, error) {
	res := make(map[string]struct{})

	path := filepath.Join(dir, "advocateResult")

	patternPrefix := `file\([0-9]+\)-test\([0-9]+\)-`
	patternFileNum := `^file\((\d+)\)-test\(\d+\)-.+$`
	rePrefix := regexp.MustCompile(patternPrefix)
	reNum := regexp.MustCompile(patternFileNum)

	files, err := os.ReadDir(path)
	if err != nil {
		return res, 0, err
	}

	maxFileNum := -1
	maxKey := ""
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		name := file.Name()
		nameClean := rePrefix.ReplaceAllString(name, "")
		lastIndex := strings.LastIndex(nameClean, "-")
		if lastIndex != -1 {
			nameClean = nameClean[:lastIndex] // Keep everything before the last separator
		}

		numbers := reNum.FindStringSubmatch(name)

		if len(numbers) > 1 {
			numberInt, err := strconv.Atoi(numbers[1])
			if err != nil {
				return res, 0, err
			}
			if numberInt > maxFileNum {
				maxKey = nameClean + ".go"
				maxFileNum = numberInt
			}
		}

		res[nameClean+".go"] = struct{}{}
	}

	// remove all folders created by the last file and remove the file name from the processed
	if flags.CancelTestIfBugFound && maxFileNum != -1 {
		for _, file := range files {
			if !file.IsDir() || !strings.Contains(file.Name(), fmt.Sprintf("file(%d)", maxFileNum)) {
				continue
			}

			_ = os.RemoveAll(filepath.Join(path, file.Name()))
		}
		log.Error()
		delete(res, maxKey)
		maxFileNum = maxFileNum - 1
	}

	return res, maxFileNum, nil
}

// FindTestFunctions find all test function in the specified file
//
// Parameter:
//   - file string: file to search in
//
// Returns:
//   - []string: functions
//   - error
func FindTestFunctions(file string) ([]string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var testFunctions []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "func Test") && strings.Contains(line, "*testing.T") {
			testFunc := strings.TrimSpace(strings.Split(line, "(")[0])
			testFunc = strings.TrimPrefix(testFunc, "func ")
			testFunctions = append(testFunctions, testFunc)
		}
	}
	return testFunctions, nil
}

// Run the full workflow for a given unit test.
// This will run, record, analyzer and, if necessary, rewrite and replay the test
//
// Parameter:
//   - runRecord bool: run the recording. If set to false, but runAnalysis or runReplay is
//     set the trace at tracePath is used
//   - runAnalysis bool: run the analysis on a path
//   - runReplay bool: run replay, if runAnalysis is true, those replays are used
//   - progName string: name of the program
//   - testName string: name of the test
//   - pkg string: adjusted package path
//   - file string: file with the test
//   - fuzzing int: -1 if not fuzzing, otherwise number of fuzzing run, starting with 0
//   - fuzzingTrace string: the path to the fuzzing trace
//   - onlyRecord bool: if true, run only the recording without any analysis
//
// Returns:
//   - int: number of run replays
//   - bool: true if analysis passed without error
//   - error
func unitTestFullWorkflow(
	runRecord, runAnalysis, runReplay bool,
	testName, pkg, file string,
	fuzzing int, fuzzingTrace string) (int, bool, error) {

	outFile, err := os.OpenFile(paths.NameOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, false, fmt.Errorf("Failed to open log file: %v", err)
	}
	defer outFile.Close()

	// Redirect stdout and stderr to the file
	origStdout := os.Stdout
	origStderr := os.Stderr

	os.Stdout = outFile
	os.Stderr = outFile

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Validate required inputs
	if paths.Prog == "" {
		return 0, false, errors.New("Directory is empty")
	}
	if testName == "" {
		return 0, false, errors.New("Test name is empty")
	}
	// if pkg == "" {
	// 	return 0, errors.New("Package is empty")
	// }
	if file == "" {
		return 0, false, errors.New("Test file is empty")
	}

	// Change to the directory
	if err := os.Chdir(paths.Prog); err != nil {
		return 0, false, fmt.Errorf("Failed to change directory: %v", err)
	}

	pkg = strings.TrimPrefix(pkg, paths.Prog)

	if runRecord {
		if flags.MeasureTime && fuzzing < 1 {
			err := unitTestRun(pkg, file, testName, origStdout, origStderr)
			if err != nil {
				if checkForTimeout(paths.NameOutput) {
					log.Timeout("Running T0 timed out")
				}
			}
		}

		err = unitTestRecord(pkg, file,
			testName, fuzzing, fuzzingTrace, paths.NameOutput, origStdout, origStderr)
		if err != nil {
			log.Error("Recording failed: ", err.Error())
		}
	}

	if runAnalysis {
		pkgPath := filepath.Join(paths.Prog, pkg)
		err = unitTestAnalyzer(pkgPath, paths.NameTrace, fuzzing, file, testName)
		if err != nil {
			return 0, false, err
		}

		if flags.OnlyAPanicAndLeak {
			return 0, true, nil
		}
	}

	numberReplay := 0
	if runReplay {
		numberReplay = unitTestReplay(paths.Prog, pkg, file, testName, paths.NameOutput, runAnalysis, origStdout, origStderr)
	}

	return numberReplay, runAnalysis, nil
}

// unitTestRun runs a test without recording/replay
//
// Parameter:
//   - pkg string: path to the package containing the test
//   - file string: path to the file containing the test function
//   - name of the test function to run
//   - osOut *os.File: file/output to write to not being what os.Stdout points to
//   - osErr *os.File: file/output to write to not being what os.Stdout points to
//
// Returns:
//   - error
func unitTestRun(pkg, file, testName string, origStdout, origStderr *os.File) error {
	timer.Start(timer.Run)
	defer timer.Stop(timer.Run)

	os.Unsetenv("GOROOT")

	log.Info("Run T0")
	packagePath := paths.MakePathLocal(pkg)
	var err error
	if flags.Timeout != -1 {
		timeoutRecString := fmt.Sprintf("%ds", flags.Timeout)
		err = command.RunCommand(origStdout, origStderr, command.NoTimeout, "go", "test", "-v", "-timeout", timeoutRecString, "-count=1", "-run="+testName, packagePath)
	} else {
		err = command.RunCommand(origStdout, origStderr, command.NoTimeout, "go", "test", "-v", "-count=1", "-run="+testName, packagePath)
	}

	return err
}

// unitTestRun runs a test and records the trace and run info
//
// Parameter:
//   - pkg string: path to the package containing the test
//   - file string: path to the file containing the test function
//   - testName string: name of the test function to run
//   - fuzzing int: number of fuzzing run. If not fuzzing, or first fuzzing run without guidance set to 0
//   - fuzzingPath string: path to the fuzzing guidance information file, if not fuzzing set to ""
//   - output string: path to where the output should be created
//   - osOut *os.File: file/output to write to not being what os.Stdout points to
//   - osErr *os.File: file/output to write to not being what os.Stdout points to
//
// Returns:
//   - error
func unitTestRecord(pkg, file, testName string,
	fuzzing int, fuzzingPath, output string, osOut, osErr *os.File) error {
	timer.Start(timer.Recording)
	defer timer.Stop(timer.Recording)

	isFuzzing := (fuzzing > 0)

	buildFlags, err := buildArgsUnit(file, testName, false, fuzzing, fuzzingPath, false)
	if err != nil {
		return fmt.Errorf("Error in adding header: %v", err)
	}

	// Run the test
	log.Info("Execute Test")

	// Set GOROOT
	os.Setenv("GOROOT", paths.GoPatch)

	command.RunCommand(osOut, osErr, command.NoTimeout, paths.Go, "version")

	pkgPath := paths.MakePathLocal(pkg)
	err = command.RunCommand(osOut, osErr, command.NoTimeout, paths.Go, "test", buildFlags, "-v", "-count=1", "-run="+testName, pkgPath)
	if err != nil {
		if isFuzzing {
			if checkForTimeout(output) {
				log.Timeout("Recording timed out")
			}
		} else {
			if checkForTimeout(output) {
				log.Timeout("Fuzzing recording timed out")
			}
		}
	}
	log.Info("Test executed")

	err = os.Unsetenv("GOROOT")

	if err != nil {
		log.Error("Failed to unset GOROOT: ", err.Error())
	}

	return err
}

// unitTestRun runs the analysis on a recorded trace
//
// Parameter:
//   - pkgPath string: path to the analyzed package
//   - traceName string: name of the trace to analyze
//   - fuzzing int: number of fuzzing run. If not fuzzing, or first fuzzing run without guidance set to 0
//   - testFile string: name of the analzed test file
//   - testName string: name of the analzed test, "main" if main
//
// Returns:
//   - error
//
// The trace is expected to be at dir/pkg/traceName
func unitTestAnalyzer(pkgPath, traceName string, fuzzing int, testFile, testName string) error {
	tracePath := filepath.Join(pkgPath, traceName)

	log.Infof("Run the analyzer for %s", tracePath)

	outM := filepath.Join(pkgPath, paths.NameResultMachine)
	outR := filepath.Join(pkgPath, paths.NameResultReadable)
	outT := filepath.Join(pkgPath, consts.RewrittenConsts)
	err := runAnalyzer(tracePath, outR, outM, outT, fuzzing, testFile, testName)

	if err != nil {
		return err
	}

	return nil
}

// unitTestReplay runs a replay for a test
//
// Parameter:
//   - pathToFoRoot string: path to the root of the modified go runtime
//   - pathToPatchedGoRuntime string: path to the patched runtime executable
//   - dir: path to the root of the analyzed project
//   - pkg string: path to the package containing the test, global path should be dir/pkg
//   - file string: path to the file containing the test function
//   - testName string: name of the test function to run
//   - output string: path to the output file
//   - runAnalysis bool: whether the rewritten traces from the analysis or the
//     given trace path should be used
//   - osOut *os.File: file/output to write to not being what os.Stdout points to
//   - osErr *os.File: file/output to write to not being what os.Stdout points to
//
// Returns:
//   - int: number of executed replays
func unitTestReplay(dir, pkg, file,
	testName, output string, fromAnalysis bool, osOut, osErr *os.File) int {
	timer.Start(timer.Replay)
	defer timer.Stop(timer.Replay)

	log.Info("Start guided execution")

	pathPkg := filepath.Join(dir, pkg)

	rewrittenConstraints := make([]string, 0)

	if fromAnalysis {
		rewrittenConstraints, _ = filepath.Glob(filepath.Join(pathPkg, consts.RewrittenConsts+"_*"))
	} else {
		rewrittenConstraints = append(rewrittenConstraints, flags.TracePath)
	}

	log.Infof("Found %d traces to replay", len(rewrittenConstraints))

	for i, trace := range rewrittenConstraints {
		traceNum, bugString := extractTraceNumber(trace)
		// record := getRerecord(trace)
		record := false

		// we do not need to replay a bug that has already been replayed by
		// another replay
		if !flags.NoSkipRewrite && results.WasAlreadyConfirmed(bugString) {
			continue
		}

		buildFlags, _ := buildArgsUnit(file, testName, true, -1, traceNum, record)

		os.Setenv("GOROOT", paths.GoPatch)

		log.Infof("Run Program for Replay %d/%d", i+1, len(rewrittenConstraints))
		pkgPath := paths.MakePathLocal(pkg)
		command.RunCommand(osOut, osErr, command.NoTimeout, paths.Go, "test", buildFlags, "-v", "-count=1", "-run="+testName, pkgPath)
		log.Infof("Finish Replay %d/%d", i+1, len(rewrittenConstraints))

		if wasReplaySuc(output) {
			results.AddBug(bugString, true)
		} else {
			results.AddBug(bugString, false)
		}

		os.Unsetenv("GOROOT")
	}

	return len(rewrittenConstraints)
}
