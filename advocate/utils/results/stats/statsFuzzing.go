// Copyright (c) 2025 Erik Kassubek
//
// File: statsFuzzing.go
// Brief: Create stats about fuzzing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package stats

import (
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CreateStatsFuzzing creates statistics about fuzzing runs
//
// Parameter:
//   - pathFolder string: path to where the stats files should be created
//
// Returns:
//   - error
func CreateStatsFuzzing(pathFolder string) error {
	// collect the info from the analyzer
	statsAnalyzerPath := filepath.Join(paths.ResultStats, "statsAnalysis_"+flags.ProgName+".csv")
	statsFuzzingPath := filepath.Join(paths.ResultStats, "statsFuzzing_"+flags.ProgName+".csv")
	statsFuzzPath := filepath.Join(paths.ResultStats, "statsFuzz_"+flags.ProgName+".csv")

	log.Info("Create fuzzing statistics")

	headers := "TestName,NrRuns,NrMuts,NrMutsInvalid,NrMutsDouble"

	for _, mode := range []string{"detected", "replayWritten", "replaySuccessful", "unexpectedPanic"} {
		for _, code := range []string{"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "A09", "P01", "P02", "P03", "P04", "P05", "L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10", "L11", "R01", "R02"} {
			headers += fmt.Sprintf(",Nr%s%s", strings.ToUpper(string(mode[0]))+mode[1:], code)
		}
	}
	headers += "\n"

	data := make(map[string]testData)

	lastTestName := ""
	counter := 0

	// get the test names and number of fuzzing runs
	analysisFile, err := os.Open(statsAnalyzerPath)
	if err != nil {
		return err
	}
	defer analysisFile.Close()

	scanner := bufio.NewScanner(analysisFile)

	// skip the first line
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		elems := strings.Split(line, ",")
		if len(elems) == 0 {
			continue
		}

		testName := elems[0]
		if lastTestName == testName {
			counter++
		} else {
			if lastTestName != "" {
				td := testData{name: lastTestName, numberRuns: counter}
				data[lastTestName] = td
			}
			lastTestName = testName
			counter = 1
		}
	}

	if lastTestName != "" {
		td := testData{name: lastTestName, numberRuns: counter}
		data[lastTestName] = td
	}

	// get the number of unique bugs for this test
	testDir, err := os.ReadDir(paths.Result)
	if err != nil {
		return err
	}

	for _, test := range testDir {
		if !test.IsDir() {
			continue
		}

		testNameSplit := strings.Split(test.Name(), "-")
		testName := testNameSplit[len(testNameSplit)-1]

		bugDir, err := os.ReadDir(paths.ResultBugs)
		if err != nil {
			continue
		}

		foundBugs := make(map[string]processedBug)
		res := getNewDataMapMap()

		for _, bug := range bugDir {
			processBugFile(filepath.Join(paths.ResultBugs, bug.Name()), foundBugs, nil, res)
		}

		for _, bug := range foundBugs {
			res[detected][bug.bugType]++

			if bug.replayWritten {
				res[replayWritten][bug.bugType]++
			}

			if bug.replaySuc {
				res[replaySuccessful][bug.bugType]++
			}
		}

		td := data[testName]
		td.results = res
		data[testName] = td
	}

	// read the fuzz data
	fuzzFile, err := os.Open(statsFuzzPath)
	if err != nil {
		return err
	}
	defer analysisFile.Close()

	scannerFuzz := bufio.NewScanner(fuzzFile)

	// skip the first line
	scannerFuzz.Scan()

	for scannerFuzz.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")

		if len(fields) < 4 {
			continue
		}

		testName := fields[0]

		td := data[testName]
		if td.fuzzData == nil {
			td.fuzzData = make(map[string]int)
		}

		nrMut, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Error(err.Error())
		}

		nrMutInvalid, err := strconv.Atoi(fields[2])
		if err != nil {
			log.Error(err.Error())
		}

		nrMutDouble, err := strconv.Atoi(fields[3])
		if err != nil {
			log.Error(err.Error())
		}

		nrActiveReleased, err := strconv.Atoi(fields[4])
		if err != nil {
			log.Error(err.Error())
		}

		td.fuzzData["nrMut"] += nrMut
		td.fuzzData["nrMutInvalid"] += nrMutInvalid
		td.fuzzData["nrMutDouble"] += nrMutDouble
		td.fuzzData["ActiveReleased"] += nrActiveReleased

		data[testName] = td
	}

	// write the data to the file
	_, err = os.Stat(statsFuzzingPath)
	fileExisted := (err == nil || !os.IsNotExist(err))

	fuzzingFile, err := os.OpenFile(statsFuzzingPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fuzzingFile.Close()

	if !fileExisted {
		fuzzingFile.WriteString(headers)
	}

	for _, d := range data {
		fuzzingFile.WriteString(d.toString() + "\n")
	}

	return nil
}
