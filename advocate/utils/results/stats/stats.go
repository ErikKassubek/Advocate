// Copyright (c) 2024 Erik Kassubek
//
// File: stats.go
// Brief: Create statistics about programs and traces
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package stats

import (
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/results/buginfo"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// testData stores information about a test
//
// Parameter:
//   - name string: name of the test
//   - numberRuns int: for fuzzing, how often the test was run
//   - results map[string]map[string]int: information about the found bugs in this test
type testData struct {
	name       string
	numberRuns int
	fuzzData   map[string]int
	results    map[statsType]map[buginfo.ResultType]int
}

// toString returns the string representation of the statistics of a test
//
// Returns:
//   - string: the string representation
func (this *testData) toString() string {
	res := fmt.Sprintf("%s,%d,%d,%d,%d", this.name, this.numberRuns, this.fuzzData["nrMut"], this.fuzzData["nrMutInvalid"], this.fuzzData["nrMutDouble"])

	for _, mode := range []statsType{detected, replayWritten, replaySuccessful, unexpectedPanic} {
		for _, code := range buginfo.ResultTypes {
			res += fmt.Sprintf(",%d", this.results[mode][code])
		}
	}

	return res
}

// CreateStats adds the information of an analyzed test to the stats info
//
// Parameter:
//   - testName string: name of the analyzed test
//   - traceID int: id of the trace
//   - fuzzing int: number of fuzzing run
//
// Returns:
//   - error
func CreateStats(testName string, traceID, fuzzing int) error {
	// statsProg, err := statsProgram(pathToProgram)
	// if err != nil {
	// 	return err
	// }

	log.Info("Create statistics")

	statsTrace, err := statsTraces(traceID)
	// if err != _ {
	// log.Error("Failed to create trace statistics: ", err.Error())
	// }

	statsFuzz, err := statsFuzz()
	if err != nil {
		log.Error("Failed to create fuzzing statistics: ", err.Error())
	}

	statsAnalyzerTotal, statsAnalyzerUnique, err := statsAnalyzer(fuzzing)
	if err != nil {
		log.Error("Failed to create analysis statistics: ", err.Error())
	}

	err = os.MkdirAll(paths.ResultStats, os.ModePerm)
	if err != nil {
		log.Error("Could not create stats folder")
	}

	err = writeStatsToFile(testName, statsTrace, statsFuzz, statsAnalyzerTotal, statsAnalyzerUnique)
	if err != nil {
		return err
	}

	return nil

}

// Write the collected statistics to files
//
// Parameter:
//   - testName string: name of the test
//   - statsProg map[statsType]int: statistics about the program
//   - statsTraces map[statsType]int: statistics about the trace
//   - statsFuzz map[statsType]int: miscellaneous statistics
//   - statsAnalyzerTotal map[statsType]map[string]int: statistics about the total analysis and replay
//   - statsAnalyzerUnique map[statsType]map[string]int: statistics about the unique analysis and replay
//
// Returns:
//   - error
func writeStatsToFile(testName string, statsTraces map[statsType]int, statsFuzz map[statsType]int,
	statsAnalyzerTotal, statsAnalyzerUnique map[statsType]map[buginfo.ResultType]int) error {

	fileFuzzPath := filepath.Join(paths.ResultStats, "statsFuzz_"+flags.ProgName+".csv")
	fileTracingPath := filepath.Join(paths.ResultStats, "statsTrace_"+flags.ProgName+".csv")
	fileAnalysisPath := filepath.Join(paths.ResultStats, "statsAnalysis_"+flags.ProgName+".csv")
	fileAllPath := filepath.Join(paths.ResultStats, "statsAll_"+flags.ProgName+".csv")

	headerTracing := "TestName,NrEvents,NrGoroutines,NrAtomicEvents," +
		"NrChannelEvents,NrSelectEvents,NrMutexEvents,NrWaitgroupEvents," +
		"NrCondVariablesEvents,NrOnceOperations"
	dataTracing := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d", testName,
		statsTraces[numberElements], statsTraces[numberRoutines],
		statsTraces[numberAtomicOperations], statsTraces[numberChannelOperations],
		statsTraces[numberSelects], statsTraces[numberMutexOperations],
		statsTraces[numberWaitGroupOperations], statsTraces[numberCondVarOperations],
		statsTraces[numberOnceOperations])

	writeStatsFile(fileTracingPath, headerTracing, dataTracing)

	numberOfActualBugsTotal := 0
	numberOfActualBugsUnique := 0
	for _, code := range buginfo.ResultTypesActual {
		numberOfActualBugsTotal += statsAnalyzerTotal[detected][code]
		numberOfActualBugsUnique += statsAnalyzerUnique[detected][code]
	}

	numberOfLeaksTotal := 0
	numberOfLeaksUnique := 0
	numberOfLeaksTotalFalsePos := 0
	numberOfLeaksUniqueFalsePos := 0

	for _, code := range buginfo.ResultTypesLeak {
		numberOfLeaksTotal += statsAnalyzerTotal[detected][code]
		numberOfLeaksUnique += statsAnalyzerUnique[detected][code]
	}
	numberOfLeaksTotalTruePos := numberOfLeaksTotal - numberOfLeaksTotalFalsePos
	numberOfLeaksUniqueTruePos := numberOfLeaksUnique - numberOfLeaksUniqueFalsePos

	for _, code := range buginfo.ResultTypesLeak {
		numberOfLeaksTotalFalsePos += statsAnalyzerTotal[falsePositive][code]
		numberOfLeaksUniqueFalsePos += statsAnalyzerUnique[falsePositive][code]
	}

	numberOfLeaksWithRewriteTotal := 0
	numberOfLeaksWithRewriteUnique := 0
	for _, code := range buginfo.ResultTypesLeak {
		numberOfLeaksWithRewriteTotal += statsAnalyzerTotal[replayWritten][code]
		numberOfLeaksWithRewriteUnique += statsAnalyzerUnique[replayWritten][code]
	}

	numberOfLeaksResolvedViaReplayTotal := 0
	numberOfLeaksResolvedViaReplayUnique := 0
	for _, code := range buginfo.ResultTypesLeak {
		numberOfLeaksResolvedViaReplayTotal += statsAnalyzerTotal[replaySuccessful][code]
		numberOfLeaksResolvedViaReplayUnique += statsAnalyzerUnique[replaySuccessful][code]
	}

	numberOfPanicsTotal := 0
	numberOfPanicsUnique := 0
	for _, code := range buginfo.ResultTypesPotential {
		numberOfPanicsTotal += statsAnalyzerTotal[detected][code]
		numberOfPanicsUnique += statsAnalyzerUnique[detected][code]
	}

	numberOfPanicsVerifiedViaReplayTotal := 0
	numberOfPanicsVerifiedViaReplayUnique := 0
	for _, code := range buginfo.ResultTypesPotential {
		numberOfPanicsVerifiedViaReplayTotal += statsAnalyzerTotal[replaySuccessful][code]
		numberOfPanicsVerifiedViaReplayUnique += statsAnalyzerUnique[replaySuccessful][code]
	}

	numberUnexpectedPanicsInReplayTotal := 0
	numberUnexpectedPanicsInReplayUnique := 0
	for _, code := range buginfo.ResultTypesPotential {
		numberUnexpectedPanicsInReplayTotal += statsAnalyzerTotal[unexpectedPanic][code]
		numberUnexpectedPanicsInReplayUnique += statsAnalyzerUnique[unexpectedPanic][code]
	}

	numberProbInRecord := 0
	for _, code := range buginfo.ResultTypesRecording {
		numberProbInRecord += statsAnalyzerTotal[detected][code]
	}

	headerAnalysis := "TestName,NumberActualBugTotal,NrLeaksTotal,NrLeaksTotalTP,NrLeaksTotalFP,NrLeaksWithRewriteTotal,NrLeaksResolvedViaReplayTotal,NrPanicsTotal,NrPanicsVerifiedViaReplayTotal,NrUnexpectedPanicsInReplayTotal,NrProbInRecordingTotal,NumberActualBugUnique,NrLeaksUnique,NrLeaksUniqueTP,NrLeaksUniqueFP,NrLeaksWithRewriteUnique,NrLeaksResolvedViaReplayUnique,NrPanicsUnique,NrPanicsVerifiedViaReplayUnique,NrUnexpectedPanicsInReplayUnique"
	dataAnalysis := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d", testName, numberOfActualBugsTotal, numberOfLeaksTotal, numberOfLeaksTotalTruePos, numberOfLeaksTotalFalsePos,
		numberOfLeaksWithRewriteTotal, numberOfLeaksResolvedViaReplayTotal, numberOfPanicsTotal, numberOfPanicsVerifiedViaReplayTotal, numberUnexpectedPanicsInReplayTotal, numberProbInRecord, numberOfActualBugsUnique, numberOfLeaksUnique, numberOfLeaksUniqueTruePos, numberOfLeaksUniqueFalsePos,
		numberOfLeaksWithRewriteUnique, numberOfLeaksResolvedViaReplayUnique, numberOfPanicsUnique, numberOfPanicsVerifiedViaReplayUnique, numberUnexpectedPanicsInReplayUnique)

	writeStatsFile(fileAnalysisPath, headerAnalysis, dataAnalysis)

	headerDetails := "TestName," +
		"NrEvents,NrGoroutines,NrNotEmptyGoroutines,NrSpawnEvents,NrRoutineEndEvents," +
		"NrAtomics,NrAtomicEvents,NrChannels,NrBufferedChannels,NrUnbufferedChannels," +
		"NrChannelEvents,NrBufferedChannelEvents,NrUnbufferedChannelEvents,NrSelectEvents," +
		"NrSelectCases,NrSelectNonDefaultEvents,NrSelectDefaultEvents,NrMutex,NrMutexEvents," +
		"NrWaitgroup,NrWaitgroupEvent,NrCondVariables,NrCondVariablesEvents,NrOnce,NrOnceOperations,"
	dataDetails := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,",
		testName, statsTraces[numberElements],
		statsTraces[numberRoutines], statsTraces[numberNonEmptyRoutines],
		statsTraces[numberOfSpawns], statsTraces[numberRoutineEnds],
		statsTraces[numberAtomics], statsTraces[numberAtomicOperations],
		statsTraces[numberChannels], statsTraces[numberBufferedChannels],
		statsTraces[numberUnbufferedChannels], statsTraces[numberChannelOperations],
		statsTraces[numberBufferedOps], statsTraces[numberUnbufferedOps],
		statsTraces[numberSelects], statsTraces[numberSelectCases],
		statsTraces[numberSelectChanOps], statsTraces[numberSelectDefaultOps],
		statsTraces[numberMutexes], statsTraces[numberMutexOperations],
		statsTraces[numberWaitGroups], statsTraces[numberWaitGroupOperations],
		statsTraces[numberCondVars], statsTraces[numberCondVarOperations],
		statsTraces[numberOnce], statsTraces[numberOnceOperations])

	headers := make([]string, 0)
	data := make([]string, 0)
	for _, mode := range []statsType{detected, replayWritten, replaySuccessful, unexpectedPanic} {
		for _, count := range []string{"Total", "Unique"} {
			for _, code := range buginfo.ResultTypes {
				headers = append(headers, "Nr"+count+strings.ToUpper(string(string(mode)[0]))+string(mode)[1:]+string(code))
				if count == "Total" {
					data = append(data, strconv.Itoa(statsAnalyzerTotal[mode][code]))
				} else {
					data = append(data, strconv.Itoa(statsAnalyzerUnique[mode][code]))
				}
			}
		}
	}
	headerDetails += strings.Join(headers, ",")
	dataDetails += strings.Join(data, ",")

	writeStatsFile(fileAllPath, headerDetails, dataDetails)

	miscData := make([]string, len(fuzzStats))
	for i, header := range fuzzStats {
		if header == "TestName" {
			miscData[i] = testName
			continue
		}
		if val, exists := statsFuzz[header]; exists {
			miscData[i] = strconv.Itoa(val)
		} else {
			miscData[i] = "0"
		}
	}

	writeStatsFile(fileFuzzPath, strings.Join(fuzzStatsStr, ","), strings.Join(miscData, ","))

	return nil
}

// writeStatsFile writes the collected stats to a csv file
//
// Parameter:
//   - path string: path to where the stat file should be created
//   - header string: first line of the stat file containing column names
//   - data string: the stats data to write into the files
func writeStatsFile(path, header, data string) {
	newFile := false
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		newFile = true
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Errorf("Error opening or creating file: %s", err.Error())
		return
	}
	defer file.Close()

	if newFile {
		file.WriteString(header)
		file.WriteString("\n")
	}
	file.WriteString(data)
	file.WriteString("\n")
}
