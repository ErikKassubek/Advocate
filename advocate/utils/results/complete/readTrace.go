// Copyright (c) 2024 Erik Kassubek
//
// File: readTrace.go
// Brief: Read in a trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package complete

import (
	"advocate/utils/consts"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/types"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// getTraceElements passes all trace files in resultFolderPath and
// extracts the position of all elements
//
// Parameter:
//   - resultFolderPath string: path to the dir containing the trace files
//
// Returns:
//   - map[string][]int: map containing all relevant lines in the form
//     file -> []lines
//   - error
func getTraceElements(resultFolderPath string) (map[string][]int, error) {
	res := make(map[string][]int)

	// for each subfolder in resultFolderPath, not recursively
	subfolder, err := getSubfolders(resultFolderPath)
	if err != nil {
		log.Error("Error in getting subfolders")
		return nil, err
	}

	for _, folder := range subfolder {
		importLine := -1
		importFile := ""
		resLocal := make(map[string][]int)

		err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Error("Error in walking trace: ", filepath.Clean(path))
				return err
			}

			fileName := filepath.Base(path)

			if info.IsDir() && fileName == paths.NameRewrittenConsts {
				return filepath.SkipDir
			}

			// read command line
			if fileName == paths.NameOutput {
				importFile, importLine, err = readCommandFile(path)
				if err != nil {
					log.Error("Error in reading command: ", filepath.Clean(path))
					return err
				}
				return nil
			}

			// read trace file
			if !strings.HasPrefix(fileName, "trace_") || !strings.HasSuffix(fileName, ".log") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				log.Error("Error in reading trace: ", filepath.Clean(path))
				return err
			}

			elems := strings.Split(string(content), "\n")

			for _, elem := range elems {
				field := strings.Split(elem, ",")
				if len(field) == 0 {
					continue
				}

				if field[0] == "A" || field[0] == "X" {
					continue
				}

				posStr := field[len(field)-1]
				pos := strings.Split(posStr, consts.PosSep)
				if len(pos) != 2 {
					log.Error("Invalid pos: ", posStr)
					continue
				}

				file := pos[0]
				line, err := strconv.Atoi(pos[1])
				if err != nil {
					continue
				}

				if _, ok := resLocal[file]; !ok {
					resLocal[file] = make([]int, 0)
				}
				resLocal[file] = append(resLocal[file], line)

				if field[0] == "S" {
					foundSelect(file, line, field[4])
				}
			}

			return nil
		})

		if err != nil {
			log.Error("Error in walking trace")
			return nil, err
		}

		// fix lines of trace with import
		for i, line := range resLocal[importFile] {
			if line >= importLine {
				resLocal[importFile][i]--
			}
		}

		// add resLocal into res
		for file, lines := range resLocal {
			if _, ok := res[file]; !ok {
				res[file] = make([]int, 0)
			}

			for _, line := range lines {
				if !types.Contains(res[file], line) {
					res[file] = append(res[file], line)
				}
			}
		}
	}

	return res, nil
}

// getSubfolders returns the path to all subfolders in a given dir
//
// Parameter:
//   - path string: the path to the main folder
//
// Returns:
//   - []string: list of all subfolders in path
//   - error
func getSubfolders(path string) ([]string, error) {
	var subfolders []string

	// Öffnen des Verzeichnisses
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	// Lesen des Verzeichnisinhalts
	files, err := dir.Readdir(-1) // -1 bedeutet, alle Einträge lesen
	if err != nil {
		return nil, err
	}

	// Filtern der Unterordner
	for _, file := range files {
		if file.IsDir() {
			subfolderPath := filepath.Join(path, file.Name())
			subfolders = append(subfolders, subfolderPath)
		}
	}

	return subfolders, nil
}

// readCommandFile reads to output.log file to determine the position of the
// inserted header.
//
// Parameter:
//   - path string: path to the output.log file
//
// Returns:
//   - string: path to the file containing the header
//   - int: line of "import advocate"
//   - error
func readCommandFile(path string) (string, int, error) {
	importLine := -1
	importFile := ""

	// read the command file
	content, err := os.ReadFile(path)
	if err != nil {
		log.Error("Error in reading command: ", filepath.Clean(path))
		return importFile, importLine, err
	}
	// find the line starting with Import added at line:
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Import added at line: ") {
			line := strings.TrimPrefix(line, "Import added at line: ")
			importLine, err = strconv.Atoi(line)
			if err != nil {
				log.Error("Error in converting import line: ", line)
				return importFile, importLine, err
			}
		}
	}

	if importLine == -1 {
		log.Error("Error in reading import or header line")
		return importFile, importLine, errors.New("Error in reading import or header line")
	}

	return importFile, importLine, nil
}
