// Copyright (c) 2024 
//
// File: progInfo.go
// Brief: Read the info required for running the program
//
//
// License: BSD-3-Clause

package explanation

import (
	"gosc/utils/paths"
	"errors"
	"os"
	"strings"
)

// Read the program info from the output.log file
//
// Returns:
//   - map[bugKeys]string: information about the analyzed test, e.g. file/test name and header position info
//   - error
func readProgInfo() (map[bugKeys]string, error) {
	res := make(map[bugKeys]string)

	f, err := os.ReadFile(paths.ResultOutput)
	if err != nil {
		return res, err
	}

	lines := strings.Split(string(f), "\n")

	if len(lines) < 3 {
		return res, errors.New("output file is too short")
	}

	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}

		if strings.Contains(lines[i], "FileName: ") {
			res[file] = strings.TrimPrefix(lines[i], "FileName: ")
		} else if strings.Contains(lines[i], "TestName: ") {
			res[name] = strings.TrimPrefix(lines[i], "TestName: ")
		} else if strings.Contains(lines[i], "Import added at line: ") {
			res[importLine] = strings.TrimPrefix(lines[i], "Import added at line: ")
		}
	}

	res[file] = paths.ToUnix(strings.TrimSpace(res[file]))
	res[name] = strings.TrimSpace(res[name])
	res[importLine] = strings.TrimSpace(res[importLine])

	return res, nil
}
