// Copyright (c) 2025 Erik Kassubek
//
// File: paths.go
// Brief: Important paths
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package paths

import (
	"advocate/utils/consts"
	"advocate/utils/flags"
	"advocate/utils/log"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// names
const (
	NameTrace             = "advocateTrace"
	NameResult            = "advocateResult"
	NameOutput            = "output.log"
	NameFuzzingData       = "fuzzingData.log"
	NameFuzzingTraces     = "advocateFuzzingTraces"
	NameReplayConstraints = "replay_constratins.log"
	NameReplayForkTree    = "replay_forktree.log"
	NameRewrittenConsts   = "advocateRewritten"
	NameTimes             = "times.log"
	NameTraceInfo         = "trace_info.log"
	NameResultMachine     = "results_machine.log"
	NameResultReadable    = "results_readable.log"
	NameRewrittenInfo     = "rewrite_info.log"
	NameStats             = "stats"
	NameStatsTime         = "times"
	NameBugs              = "bugs"
	NameTraces            = "traces"
	NameOut               = "output"
)

// advocate
var (
	Advocate = ""
	GoPatch  = ""
	Go       = ""
)

// prog
var (
	Prog    = ""
	ProgDir = ""
)

// work
var (
	FuzzingTraces = ""
	Output        = ""
)

// Result
var (
	Result        = ""
	CurrentResult = ""
	ResultOutput  = ""
	ResultBugs    = ""
	ResultStats   = ""
	ResultTime    = ""
	ResultTraces  = ""
	ResultOut     = ""
)

func BuildPaths(main bool) {
	pathsAdvocate()
	pathsProg()
	pathsResult(main)
}

func pathsAdvocate() {
	execPath, err := os.Executable()
	if err != nil {
		log.Error(err.Error())
	}

	Advocate = CleanPathHome(filepath.Dir(filepath.Dir(execPath)))
	GoPatch = filepath.Join(Advocate, "goPatch")
	Go = filepath.Join(Advocate, "goPatch", "bin", "go")

	if runtime.GOOS == "windows" {
		Go += ".exe"
	}
}

func pathsProg() {
	Prog = CleanPathHome(flags.ProgPath)
	ProgDir = GetDirectory(Prog) // only for main
	Output = filepath.Join(ProgDir, NameOutput)
	FuzzingTraces = filepath.Join(ProgDir, NameFuzzingTraces)
}

func pathsResult(main bool) {
	if main {
		Result = filepath.Join(ProgDir, NameResult)
	} else {
		Result = filepath.Join(Prog, NameResult)
	}
	CurrentResult = Result
	ResultOut = filepath.Join(CurrentResult, NameOut)
	ResultOutput = filepath.Join(ResultOut, NameOutput)
	ResultBugs = filepath.Join(CurrentResult, NameBugs)

	ResultStats = filepath.Join(Result, NameStats)
	ResultTime = filepath.Join(Result, NameStatsTime)
	ResultTraces = filepath.Join(Result, NameTraces)
}

func SetCurrentResult(fileNumber, testNumber int, fileName, testName string) string {
	dirName := fmt.Sprintf("file(%d)-test(%d)-%s-%s", fileNumber, testNumber, fileName, testName)
	CurrentResult = filepath.Join(Result, dirName)
	ResultOut = filepath.Join(CurrentResult, NameOut)
	ResultBugs = filepath.Join(CurrentResult, NameBugs)
	ResultOutput = filepath.Join(ResultOut, NameOutput)
	ResultStats = filepath.Join(CurrentResult, NameStats)
	ResultTime = filepath.Join(CurrentResult, NameStatsTime)
	ResultTraces = filepath.Join(CurrentResult, NameTraces)
	return CurrentResult
}

// GetDirectory returns the folder a file is in from the path
//
// Parameter:
//   - path string: the path to the file
//
// Returns:
//   - string: if path points to file, the folder it is in, if it points to a folder, the path
func GetDirectory(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return path
	}

	if info.IsDir() {
		// Already a directory
		return filepath.Clean(path)
	}

	// It's a file, return its directory
	return filepath.Dir(path)
}

// GetMainPath takes a path. If the path points to a file, it will return the path.
// If not it will check if the folder it points to contains a main.go file.
// If it does, it will return the path to the file
//
// Parameter:
//   - path string: path
//
// Returns:
//   - string: path to the main file
//   - error
func GetMainPath(path string) (string, error) {
	path = CleanPathHome(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		mainPath := filepath.Join(path, "main.go")

		if _, err := os.Stat(mainPath); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("main.go not found in directory %s", path)
			}
			return "", err
		}
		return mainPath, nil
	}

	// It's a file, return the path as is
	return filepath.Clean(path), nil
}

// CheckPath checks if the provided path to the program that should
// be run/analyzed exists. If not, it panics.
//
// Parameter:
//   - path string: path to check
//
// Returns:
//   - string: the cleaned path
//   - error: error if path not exists, nil otherwise
func CheckPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("Path cannot be empty")
	}

	progPath := CleanPathHome(path)

	_, err := os.Stat(progPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return progPath, fmt.Errorf("Path %s does not exists", progPath)
		}
		return progPath, err
	}

	return progPath, nil
}

// GetProgName returns a program name from the path
//
// Parameter:
//   - path string: path to the program
//
// Returns:
//   - string: name for the program
func GetProgName(path string) string {
	path = strings.ReplaceAll(path, "~"+consts.Sep, "")
	path = strings.ReplaceAll(path, "."+consts.Sep, "")
	path = strings.ReplaceAll(path, ".", "-")
	return strings.ReplaceAll(path, consts.Sep, "-")
}
