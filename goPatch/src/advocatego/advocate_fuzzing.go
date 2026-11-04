// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_fuzzing.go
// Brief: Fuzzing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package advocatego

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	_ "unsafe"
)

var isFuzzing = false
var finishFuzzingStarted = false

var initFuzzing bool

// Initialize fuzzing
//
// Parameter:
//   - tracePath string: For fuzzing approaches that use trace, add the path to the
//     trace, otherwise set to ""
//   - timeout int: Timeout in seconds
//   - init bool: true if called from init, false otherwise
//
//go:linkname InitFuzzing runtime.AdvocateInitFuzzing
func InitFuzzing(tracePath string, timeout int, init bool) {
	if initFuzzing { // called by main but alredy run by init
		return
	}
	initFuzzing = true

	FinishFunc = FinishFuzzing

	prefSel := make(map[string][]int)
	prefFlow := make(map[string][]int)

	InitTracing(timeout, init)

	if tracePath == "" { // GoFuzz
		fuzzingSelectPath := "fuzzingData.log"
		var err error
		prefSel, prefFlow, err = readFuzzingSelectFile(fuzzingSelectPath)
		if err != nil {
			println("Error in reading ", fuzzingSelectPath, ": ", err.Error())
			panic(err)
		}
		runtime.InitFuzzingSelect(prefSel, prefFlow, FinishFuzzing)
	} else { // Constraints
		runtime.InitFuzzingReplay(FinishFuzzing)
		tracePathRewritten = tracePath
		startReplay(timeout)
	}
}

// Run when fuzzing is finished (normally as defer)
// This records the traces and some additional info
//
//go:linkname FinishFuzzing runtime.AdvocateFinishFuzzing
func FinishFuzzing() {
	if finishFuzzingStarted {
		return
	}

	// give program time to finish running
	runtime.ReleaseAllBlocked()
	// runtime.SleepBusy(3)

	finishFuzzingStarted = true
	runtime.WaitForReplayFinish()
	FinishTracing()
}

// Read the file containing the preferred select cases
//
// Parameter:
//   - pathSelect string: path to the file containing the select
//     preferred cases
//
// Returns:
//   - map[string][]int: key: file:line of select, values: list of preferred cases in select
//   - map[string]int: key: file:line of select, values: counter of operation to delay
//   - error
func readFuzzingSelectFile(pathSelect string) (map[string][]int, map[string][]int, error) {
	resSelect := make(map[string][]int)
	resFlow := make(map[string][]int)

	file, err := os.Open(pathSelect)
	if err != nil {
		return resSelect, resFlow, err
	}

	mode := 1

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if line == "@" {
			mode = 2
			continue
		}

		elems := strings.Split(line, ";")
		if len(elems) != 2 {
			return resSelect, resFlow, fmt.Errorf("Incorrect line in fuzzing select file: %s", line)
		}

		path := elems[0]

		if mode == 1 {
			ids := strings.Split(elems[1], ",")

			if len(ids) == 0 {
				continue
			}

			resSelect[path] = make([]int, len(ids))
			for i, id := range ids {
				idInt, err := strconv.Atoi(id)
				if err != nil {
					return resSelect, resFlow, err
				}
				resSelect[path][i] = idInt
			}
		} else {
			counts := strings.Split(elems[1], ",")
			for _, count := range counts {
				countInt, err := strconv.Atoi(count)
				if err != nil {
					return resSelect, resFlow, err
				}
				if _, ok := resFlow[path]; !ok {
					resFlow[path] = make([]int, 0)
				}

				resFlow[path] = append(resFlow[path], countInt)
			}
		}
	}

	return resSelect, resFlow, nil
}
