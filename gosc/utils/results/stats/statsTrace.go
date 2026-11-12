// Copyright (c) 2024 
//
// File: statsTrace.go
// Brief: Collect statistics about the trace
//
//
// License: BSD-3-Clause

package stats

import (
	"gosc/utils/log"
	"gosc/utils/paths"
	"gosc/utils/types"
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Collect stats about the traces
//
// Parameter:
//   - traceID int: name of trace folder is datapath_traceId
//
// Returns:
//   - map[string]int: map with the stats
//   - error
func statsTraces(traceID int) (map[statsType]int, error) {
	res := map[statsType]int{
		numberElements: 0,

		numberRoutines:         0,
		numberNonEmptyRoutines: 0,

		numberOfSpawns:    0,
		numberRoutineEnds: 0,

		numberAtomics:          0,
		numberAtomicOperations: 0,

		numberChannels:           0,
		numberBufferedChannels:   0,
		numberUnbufferedChannels: 0,
		numberChannelOperations:  0,
		numberBufferedOps:        0,
		numberUnbufferedOps:      0,

		numberSelects:          0,
		numberSelectCases:      0,
		numberSelectChanOps:    0, // number of executed channel operations in select
		numberSelectDefaultOps: 0, // number of executed default operations in select

		numberMutexes:         0,
		numberMutexOperations: 0,

		numberWaitGroups:          0,
		numberWaitGroupOperations: 0,

		numberCondVars:          0,
		numberCondVarOperations: 0,

		numberOnce:           0,
		numberOnceOperations: 0,
	}

	tracePath := filepath.Join(paths.ResultTraces, fmt.Sprintf("goscTrace_%d", traceID))

	// do not count the same twice
	known := map[statsType][]string{
		atomicElem:    {},
		channelElem:   {},
		mutexElem:     {},
		waitGroupElem: {},
		condVarElem:   {},
		onceElem:      {},
	}

	err := filepath.Walk(tracePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() != paths.NameTraceInfo {
			err = parseTraceFile(path, &res, &known)
			if err != nil {
				log.Error(err)
			}
		}

		return nil
	})

	return res, err
}

// parseTraceFile parses a trace file to get all relevant stats information
//
// Parameter:
//   - tracePath string: Path the the trace file
//   - stats *map[statsType]int: Map to store the information in
//   - known *map[statsType][]string: Information about primitives that have already been
//   - seem in other trace files
//
// Returns:
//   - error
func parseTraceFile(tracePath string, stats *map[statsType]int, known *map[statsType][]string) error {
	// open the file
	file, err := os.Open(tracePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// routine, err := getRoutineFromFileName(filepath.Base(tracePath))
	// if err != nil {
	// 	return fmt.Errorf("%s not an trace file", tracePath)
	// }
	(*stats)[numberRoutines]++

	scanner := bufio.NewScanner(file)

	// read the file
	foundNonEmpty := false
	for scanner.Scan() {
		elem := scanner.Text()

		if elem != "" && !foundNonEmpty {
			(*stats)[numberNonEmptyRoutines]++
			foundNonEmpty = true
		}
		(*stats)[numberElements]++
		fields := strings.Split(elem, ",")
		switch fields[0] {
		case "G":
			(*stats)[numberOfSpawns]++
		case "A":
			(*stats)[numberAtomicOperations]++
			if !types.Contains((*known)[atomicElem], fields[2]) {
				(*stats)[numberAtomics]++
				(*known)[atomicElem] = append((*known)[atomicElem], fields[2])
			}
		case "C":
			(*stats)[numberChannelOperations]++
			if fields[7] == "0" {
				(*stats)[numberUnbufferedOps]++
			} else {
				(*stats)[numberBufferedOps]++
			}
			if !types.Contains((*known)[channelElem], fields[3]) {
				(*stats)[numberChannels]++
				if fields[7] == "0" {
					(*stats)[numberUnbufferedChannels]++
				} else {
					(*stats)[numberBufferedChannels]++
				}
				(*known)[channelElem] = append((*known)[channelElem], fields[3])
			}
		case "S":
			(*stats)[numberSelects]++
			cases := strings.Split(fields[4], "~")
			(*stats)[numberSelectCases] += len(cases)
			if cases[len(cases)-1] == "D" {
				(*stats)[numberSelectDefaultOps]++
			} else {
				(*stats)[numberSelectChanOps] += len(cases)
			}
		case "M":
			(*stats)[numberMutexOperations]++
			if !types.Contains((*known)[mutexElem], fields[3]) {
				(*stats)[numberMutexes]++
				(*known)[mutexElem] = append((*known)[mutexElem], fields[3])
			}
		case "W":
			(*stats)[numberWaitGroupOperations]++
			if !types.Contains((*known)[waitGroupElem], fields[3]) {
				(*stats)[numberWaitGroups]++
				(*known)[waitGroupElem] = append((*known)[waitGroupElem], fields[3])
			}
		case "O":
			(*stats)[numberOnceOperations]++
			if !types.Contains((*known)[onceElem], fields[3]) {
				(*stats)[numberOnce]++
				(*known)[onceElem] = append((*known)[onceElem], fields[3])
			}
		case "D":
			(*stats)[numberCondVarOperations]++
			if !types.Contains((*known)[condVarElem], fields[3]) {
				(*stats)[numberCondVars]++
				(*known)[condVarElem] = append((*known)[condVarElem], fields[3])
			}
		case "E":
			(*stats)[numberRoutineEnds]++
		case "N", "I", "R", "F", "OAT":
			// do notring
		default:
			err = errors.New("Unknown trace element: " + fields[0])
		}
	}
	return err
}
