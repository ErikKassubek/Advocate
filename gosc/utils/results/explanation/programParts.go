// Copyright (c) 2024 
//
// File: programParts.go
// Brief: Read the program code at the positions of the bug
//
//
// License: BSD-3-Clause

package explanation

import (
	"gosc/utils/consts"
	"gosc/utils/log"
	"gosc/utils/paths"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Get the positions of the bug elements in the program
//
// Parameter:
//   - traceElems map[int]string: The trace elements of the bug
//
// Returns:
//   - map[int][]string: Dict for the code snippets
func getBugPositions(traceElems map[int][]string) (map[int][]string, error) {
	res := make(map[int][]string)

	for i, elem := range traceElems {
		for _, e := range elem {
			pos := strings.Split(e, consts.PosSep)
			file := pos[0]
			line, err := strconv.Atoi(pos[1])
			if err != nil {
				log.Error("Invalid line: ", pos[1])
			}

			code, err := GetProgramCode(file, line, true)
			if err != nil {
				res[i] = append(res[i], "")
				log.Error(err.Error())
			} else {
				res[i] = append(res[i], code)
			}
		}
	}

	return res, nil
}

// GetProgramCode returns the code snippet of a program file at a specific line
//
// Parameter:
//   - file string: The path to the file
//   - line int: The line number
//   - numbers bool: If line numbers should be included
//
// Returns:
//   - string: The code snippet
//   - error: An error if the file could not be read
func GetProgramCode(file string, line int, numbers bool) (string, error) {
	file = paths.ToLocal(file)
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if line < 0 || line >= len(lines) {
		return "", errors.New("line number out of range")
	}

	line = line - 1 // 1 based index of lines to 0 base index of slice

	buffer := 10

	start := max(0, line-buffer)
	end := min(line+buffer, len(lines)-1)

	code := lines[start : end+1]

	res := "```go"

	if start != 1 {
		res += "...\n"
	}

	width := len(fmt.Sprintf("%d", 20+start))

	for i, l := range code {
		ln := i + start + 1
		if numbers {
			res += fmt.Sprintf("%*d ", width, ln)
		}
		res += l

		if ln-1 == line {
			res += "                    // <================= "
		}

		res += "\n"
	}

	if end != len(lines)-1 {
		res += "...\n"
	}

	res += "```"

	return res, nil
}
