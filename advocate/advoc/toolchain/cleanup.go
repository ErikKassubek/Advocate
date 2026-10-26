// Copyright (c) 2025 Erik Kassubek
//
// File: cleanup.go
// Brief: Cleanup traces and files
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/utils/consts"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var LastMovedTraceDest = ""

// Function to move results files from the package directory to the destination directory
//
// Parameter:
//   - progPath string: path to the program
//   - packagePath string: path to the package directory
//   - destination string: path to the destination directory
//   - total bool: merge all already created logs into total log, for fuzzing
//   - collectConstr bool: collect the constraint files
func collect(progPath, packagePath, destination string, total bool, collectConstr bool) {
	filesToMove := []string{
		paths.NameTrace,
		paths.NameResultMachine,
		paths.NameResultReadable,
		paths.NameOutput,
	}

	pattersToMove := []string{
		"rewrittenTrace*",
	}

	if collectConstr {
		pattersToMove = append(pattersToMove, consts.RewrittenConsts+"*")
	}

	logsToCollect := []string{
		filepath.Join(packagePath, paths.NameResultMachine),
		filepath.Join(packagePath, paths.NameResultReadable),
		filepath.Join(progPath, paths.NameOutput),
	}

	pathTraces := filepath.Join(destination, paths.NameTraces)
	pathOut := filepath.Join(destination, paths.NameOut)

	err := os.MkdirAll(pathTraces, os.ModePerm)
	if err != nil {
		log.Error("Error creating folder:", err)
	}

	err = os.MkdirAll(pathOut, os.ModePerm)
	if err != nil {
		log.Error("Error creating folder:", err)
	}

	if total {
		for _, file := range logsToCollect {
			src := file
			dest := filepath.Join(pathOut, "total_"+filepath.Base(file))

			_, err := os.Stat(dest)
			new := os.IsNotExist(err)

			srcFile, err := os.Open(src)
			if err != nil {
				// log.Errorf("Could not open src file: %s %s", src, err.Error())
				continue
			}
			defer srcFile.Close()

			destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				// log.Errorf("Could not open dest file: %s %s", dest, err.Error())
				continue
			}
			defer destFile.Close()

			if !new {
				_, err = destFile.WriteString("==================================\n")
			}

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				log.Error("Could not merge ", src, " int ", dest, ": ", err.Error())
			}
		}
	}

	for _, file := range filesToMove {
		src := filepath.Join(packagePath, file)
		if file == paths.NameOutput {
			src = filepath.Join(progPath, file)
		}
		dest := filepath.Join(destination, file)

		if file == paths.NameTrace {
			movedTraces++
			dest = filepath.Join(pathTraces, file+"_"+strconv.Itoa(movedTraces))
			LastMovedTraceDest = dest
		} else {
			dest = filepath.Join(pathOut, file)
		}

		_ = os.Rename(src, dest)
		// if err != nil {
		// log.Errorf("Could not rename: %s %s", dest, err.Error())
		// }
	}

	for _, pattern := range pattersToMove {
		files, err := filepath.Glob(filepath.Join(packagePath, pattern))
		if err != nil {
			log.Error(err.Error())
		}

		for _, trace := range files {
			dest := filepath.Join(pathTraces, filepath.Base(trace))
			_ = os.Rename(trace, dest)
		}
	}
}

// RemoveTraces removes all traces, both recorded and rewritten from the path
//
// Parameter:
//   - path string: path to the folder containing the traces
func RemoveTraces(path string) {
	pattersToMove := []string{
		"advocateTrace_*",
		"rewrittenTrace*",
		paths.NameFuzzingData,
		// "fuzzingTrace_*",
	}

	files := make([]string, 0)
	filepath.WalkDir(path, func(p string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if flags.TracePath != "" && strings.HasPrefix(p, strings.TrimSuffix(flags.TracePath, string(filepath.Separator))) {
			return nil
		}

		// Use Glob to check if the file/directory matches the pattern
		for _, pattern := range pattersToMove {
			match, err := filepath.Match(pattern, filepath.Base(p))
			if err != nil {
				return err
			}

			if match {
				files = append(files, p)
			}
		}

		return nil
	})

	for _, trace := range files {
		os.RemoveAll(trace)
	}

	os.Remove(paths.ResultTraces)
}

// removeLogs removes the result and output files
//
// Parameter:
//   - path to the folder containing the result and output files
func removeLogs(path string) {
	logsToRemove := []string{
		paths.NameResultMachine,
		paths.NameResultReadable,
		paths.NameOutput,
	}

	for _, logFile := range logsToRemove {
		os.Remove(filepath.Join(path, logFile))
	}
}

// ClearFuzzingTrace removes the fuzzing trace folder
func ClearFuzzingTrace() {

	if !flags.DeleteTraces {
		_ = os.Rename(paths.FuzzingTraces, filepath.Join(paths.CurrentResult, paths.NameFuzzingTraces))
		// if err != nil {
		// 	log.Errorf("failed to move folder %s to %s: %s", fuzzingPath, fuzzingPath, err.Error())
		// }
	} else {
		err := os.RemoveAll(paths.FuzzingTraces)
		if err != nil {
			log.Errorf("Could not delete fuzzingTraces: %s", err.Error())
		}
	}
}
