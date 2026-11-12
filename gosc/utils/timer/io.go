// Copyright (c) 2025 
//
// File: io.go
// Brief: Function to write the time file
//
//
// License: BSD-3-Clause

package timer

import (
	"gosc/utils/flags"
	"gosc/utils/log"
	"gosc/utils/paths"
	"fmt"
	"os"
	"path/filepath"
)

// UpdateTimeFileDetail writes the time information to a file
//
// Parameter:
//   - testName string: name of the test
//   - numberReplay int: number of replay
func UpdateTimeFileDetail(testName string, numberReplay int) {
	if !flags.MeasureTime {
		return
	}

	err := os.MkdirAll(paths.ResultTime, os.ModePerm)
	if err != nil {
		log.Errorf("Error creating directory:", err)
		return
	}

	timeFilePath := filepath.Join(paths.ResultTime, "times_detail_"+flags.ProgName+".csv")

	newFile := false
	_, err = os.Stat(timeFilePath)
	if os.IsNotExist(err) {
		newFile = true
	}

	file, err := os.OpenFile(timeFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Error opening or creating file:", err)
		return
	}
	defer file.Close()

	if newFile {
		csvTitels := "TestName,Run,Recording,Io,Analysis,AnaHb,AnaExitCode,AnaLeak,AnaClose,AnaConcurrent,AnaResource,AnaSelWithoutPartner,AnaUnlock,AnaWait,AnaMixed,FuzzingAna,FuzzingMut,Rewrite,Replay,NumberReplay\n"
		if _, err := file.WriteString(csvTitels); err != nil {
			log.Error("Could not write time: ", err)
		}
	}

	timeInfo := fmt.Sprintf(
		"%s,%s,%d\n", testName, ToString(), numberReplay)

	// Write to the file
	if _, err := file.WriteString(timeInfo); err != nil {
		log.Error("Could not write time: ", err)
	}
}

// UpdateTimeFileOverview write the current timer values to a file
// if time measurement is enabled,
//
// Parameter:
//   - testName string: name of the test
func UpdateTimeFileOverview(testName string) {
	if !flags.MeasureTime {
		return
	}

	err := os.MkdirAll(paths.ResultTime, os.ModePerm)
	if err != nil {
		log.Errorf("Error creating directory:", err)
		return
	}

	timeFilePath := filepath.Join(paths.ResultTime, "times_total_"+flags.ProgName+".csv")

	newFile := false
	_, err = os.Stat(timeFilePath)
	if os.IsNotExist(err) {
		newFile = true
	}

	file, err := os.OpenFile(timeFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Error opening or creating file:", err)
		return
	}
	defer file.Close()

	if newFile {
		csvTitels := "TestName,Time\n"
		if _, err := file.WriteString(csvTitels); err != nil {
			log.Error("Could not write time: ", err)
		}
	}

	timeInfo := ""
	if testName == "*Total*" {
		timeInfo = fmt.Sprintf(
			"%s,%.5f\n", "Total", GetTime(Total).Seconds())
	} else {
		timeInfo = fmt.Sprintf(
			"%s,%.5f\n", testName, GetTime(TotalTest).Seconds())
	}

	// Write to the file
	if _, err := file.WriteString(timeInfo); err != nil {
		log.Error("Could not write time: ", err)
	}
}
