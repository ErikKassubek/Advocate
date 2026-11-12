// Copyright (c) 2024 
//
// File: /gosc/trace/utils.go
// Brief: Collection of utility functiond for trace analysis
//
//
// License: BSD-3-Clause

package trace

import (
	"gosc/utils/consts"
	"fmt"
	"strconv"
	"strings"
)

// SameRoutine determines if for aal trace elements in the list, if they are
// operations on the same primitive, they have the same routine
//
// Parameter:
//   - elems ...[]TraceElement: lists of trace elements
//
// Returns:
//   - true if for each primitive, the element in elems are always in the same routine
func SameRoutine(elems ...[]Element) bool {
	ids := make(map[int]int)
	for _, elem := range elems {
		for i, e := range elem {
			if _, ok := ids[i]; !ok {
				ids[i] = e.RoutineID()
			} else if ids[i] != e.RoutineID() {
				return false
			}
		}
	}

	return true
}

// PosFromPosString returns the file and line from a pos string
//
// Parameter:
//   - pos string: [file]:[line]
//
// Returns:
//   - string: file
//   - int: line
//   - error
func PosFromPosString(pos string) (string, int, error) {
	posSplit := strings.Split(pos, consts.PosSep)
	if len(posSplit) != 2 {
		return "", 0, fmt.Errorf("Invalid pos %s", pos)
	}

	line, err := strconv.Atoi(posSplit[1])
	if err != nil {
		return "", 0, fmt.Errorf("Invalid pos %s: %s", pos, err.Error())
	}

	return posSplit[0], line, nil
}

// PosStringFromPos returns the pos string from a file and line
//
// Parameter:
//   - string: file
//   - int: line
//
// Returns:
//   - pos string: [file]:[line]
func PosStringFromPos(file string, line int) string {
	return fmt.Sprintf("%s%s%d", file, consts.PosSep, line)
}
