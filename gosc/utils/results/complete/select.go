// Copyright (c) 2024 
//
// File: select.go
// Brief: Read all select cases in the program
//
//
// License: BSD-3-Clause

package complete

import (
	"strings"
)

var selects = make(map[string]map[int][]bool)       // file -> line -> []wasSelected
var containsDefault = make(map[string]map[int]bool) // file -> line -> containsDefault

// FoundSelect is called when a select statement is found.
// It records the select statement.
//
// Parameter:
//   - elem: the select statement
func foundSelect(file string, line int, cases string) {
	casesSplit := strings.Split(cases, "~")
	for i, c := range casesSplit {
		if c == "" { // empty case
			continue
		} else if c == "D" { // selected default case
			addSelect(file, line, len(casesSplit), i)
			if _, ok := containsDefault[file]; !ok {
				containsDefault[file] = make(map[int]bool)
			}
			containsDefault[file][line] = true
		} else if c == "d" { // not selected default case
			if _, ok := containsDefault[file]; !ok {
				containsDefault[file] = make(map[int]bool)
			}
			containsDefault[file][line] = true
		} else { // case
			fields := strings.Split(c, ".")
			if fields[2] == "0" { // not selected case
				continue
			} else {
				addSelect(file, line, len(casesSplit), i)
			}
		}
	}
}

// AddSelect is called to add a select case into selects.
// It records the selected case.
//
// Parameter:
//   - file: the file in which the select statement is found
//   - line: the line number of the select statement
//   - numberCases: the number of cases in the select statement, including the default case
//   - selected: the index of the selected case
func addSelect(file string, line int, numberCases int, selected int) {
	// ignore definition of select in src/runtime/select.go
	if strings.HasSuffix(file, "src/runtime/select.go") {
		return
	}

	if _, ok := selects[file]; !ok {
		selects[file] = make(map[int][]bool)
	}
	if _, ok := selects[file][line]; !ok {
		selects[file][line] = make([]bool, numberCases)
	}
	selects[file][line][selected] = true
}

// GetNotSelectedSelectCases returns the select cases that were not selected.
//
// Returns:
//   - map[string]map[int][]int: the not selected select cases, file -> line -> []caseIndex
func getNotSelectedSelectCases() map[string]map[int][]int {
	res := make(map[string]map[int][]int)
	for file, lines := range selects {
		for line, cases := range lines {
			for i, c := range cases {
				if !c {
					if _, ok := res[file]; !ok {
						res[file] = make(map[int][]int)
					}
					if _, ok := res[file][line]; !ok {
						res[file][line] = make([]int, 0)
					}

					if contains, ok := containsDefault[file][line]; i == len(cases)-1 && ok && contains { // default case
						res[file][line] = append(res[file][line], -1)
					} else {
						res[file][line] = append(res[file][line], i)
					}
				}
			}
		}
	}
	return res
}
