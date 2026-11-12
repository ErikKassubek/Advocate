// Copyright (c) 2024 
//
// File: complete.go
// Brief: Functions to check if all program elements have been executed at least once
//
//
// License: BSD-3-Clause

package complete

import (
	"gosc/utils/consts"
	"gosc/utils/log"
	"gosc/utils/results/explanation"
	"gosc/utils/types"
	"fmt"
	"os"
	"strings"
)

// Check if all program elements are in trace
//
// Parameter:
//   - resultFolderPath: path to the folder containing the trace files
//   - progPath: path to the program file
//
// Returns:
//   - error: error if any
func Check(resultFolderPath string, progPath string) error {
	progElems, err := getProgramElements(progPath)
	if err != nil {
		log.Error("Error in getProgramElements")
		return err
	}

	traceElems, err := getTraceElements(resultFolderPath)
	if err != nil {
		log.Error("Error in getTraceElements")
		return err
	}

	notInTrace := areAllProgElemInTrace(progElems, traceElems)
	notSelectedSelectCase := getNotSelectedSelectCases()

	err = printNotExecutedToFiles(notInTrace, notSelectedSelectCase, resultFolderPath)

	return err
}

// areAllProgElemInTrace takes all relevant element positions in the program
// and all relevant element positions in the traces to determine if all
// elements in the program have been executed at least once
//
// Parameter:
//   - progElems (map[string][]int) positions (file -> []line) of all relevant elems in the program
//   - traceElems (map[string][]int) positions (file -> []line) of all relevant elems in the traces
//
// Returns:
//   - map[string][]int: all elems (file -> []line) from the program that have
//     never been executed
func areAllProgElemInTrace(progElems map[string][]int, traceElems map[string][]int) map[string][]int {
	res := map[string][]int{}

	for file, lines := range progElems {
		// file not recorded in trace
		if _, ok := traceElems[file]; !ok {
			if _, ok := res[file]; !ok {
				res[file] = make([]int, 0)
			}

			res[file] = append(res[file], -1) // -1 signaling, that no element in file was in trace
			res[file] = append(res[file], lines...)
		}

		for _, line := range lines {
			if !types.Contains(traceElems[file], line) {
				if _, ok := res[file]; !ok {
					res[file] = make([]int, 0)
				}

				res[file] = append(res[file], line)
			}
		}
	}

	return res
}

// GetNotSelectedSelectCases prints the elements and select cases that were not executed
// into a file.
//
// Parameter:
//   - elements: the elements that were not executed
//   - selects: the select cases that were not selected
func printNotExecutedToFiles(elements map[string][]int, selects map[string]map[int][]int,
	path string) error {

	path = fmt.Sprintf("%s%sGoscNotExecuted", path, consts.Sep)

	// create a folder to store results
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	pathOperations := fmt.Sprintf("%s%sGoscNotExecutedOperations.txt", path, consts.Sep)
	pathSelects := fmt.Sprintf("%s%sGoscNotExecutedSelectCases.txt", path, consts.Sep)

	notExecutedOperationsFile, err := os.Create(pathOperations)
	if err != nil {
		return err
	}
	defer notExecutedOperationsFile.Close()

	notExecutedSelectFile, err := os.Create(pathSelects)
	if err != nil {
		return err
	}
	defer notExecutedSelectFile.Close()

	pathOperationsFolder := fmt.Sprintf("%s%sOperations", path, consts.Sep)
	err = os.MkdirAll(pathOperationsFolder, os.ModePerm)
	if err != nil {
		return err
	}

	// write elements that were not executed
	if len(elements) > 0 {
		for file, lines := range elements {
			fileName := strings.ReplaceAll(file, consts.Sep, "_")
			fileName = strings.TrimPrefix(fileName, "_")
			pathFile := fmt.Sprintf("%s%s%s.md", pathOperationsFolder, consts.Sep, fileName)
			fileFile, err := os.Create(pathFile)
			if err != nil {
				return err
			}
			defer fileFile.Close()
			fileFile.WriteString(fmt.Sprintf("# %s\n", file))
			fileFile.WriteString("## Not executed operations\n")

			notExecutedOperationsFile.WriteString(fmt.Sprintf("%s:[", file))
			for i, line := range lines {
				if line == -1 {
					notExecutedOperationsFile.WriteString("No element in file was executed")
					fileFile.Write([]byte("No element in file was executed"))
					break
				} else {
					notExecutedOperationsFile.WriteString(fmt.Sprintf("%d", line))
					if i != len(lines)-1 {
						notExecutedOperationsFile.WriteString(",")
					}
					fileFile.WriteString(fmt.Sprintf("### Line: %d\n", line))
					code, err := explanation.GetProgramCode(file, line, true)
					if err != nil {
						fileFile.WriteString("Error reading file: " + err.Error() + "\n")
						continue
					}
					fileFile.WriteString(code)
				}
			}
			notExecutedOperationsFile.WriteString("]\n")
		}
	} else {
		notExecutedOperationsFile.WriteString("All program elements were executed\n")
	}

	// write select cases that were not selected
	if len(selects) > 0 {
		for file, lines := range selects {
			fileName := strings.ReplaceAll(file, consts.Sep, "_")
			fileName = strings.TrimPrefix(fileName, "_")
			pathFile := fmt.Sprintf("%s/%s.md", pathOperationsFolder, fileName)
			// if file does not exist, create it otherwise append to it
			fileFile, err := os.OpenFile(pathFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer fileFile.Close()
			fileFile.WriteString("## Not selected select cases\n")

			for line, cases := range lines {
				notExecutedSelectFile.WriteString(fmt.Sprintf("%s:%d:[", file, line))
				fileFile.WriteString(fmt.Sprintf("### Line: %d\n", line))
				for i, c := range cases {
					if c == -1 {
						notExecutedSelectFile.WriteString("D")
						fileFile.WriteString("Default case was never selected\n")
					} else {
						notExecutedSelectFile.WriteString(fmt.Sprintf("%d", c))
						fileFile.WriteString(fmt.Sprintf("Case: %d was never selected\n", c))
					}

					if i != len(cases)-1 {
						notExecutedSelectFile.WriteString(",")
					}

					code, err := explanation.GetProgramCode(file, line, true)
					if err != nil {
						fileFile.WriteString("Error reading file: " + err.Error() + "\n")
						continue
					}
					fileFile.WriteString(code)
				}
				notExecutedSelectFile.WriteString("]\n")
			}
		}
	} else {
		notExecutedSelectFile.WriteString("All select cases were executed\n")
	}

	return nil
}
