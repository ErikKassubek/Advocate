// Copyright (c) 2025 
//
// File: parser.go
// Brief: Parse flags
//
//
// License: BSD-3-Clause

package flags

import (
	"fmt"
)

// AnalysisCases represent a possible type ob bug/leak/info the HB info should look for
type AnalysisCases string

// Possible type ob bug/leak/info the HB info should look for
const (
	All          AnalysisCases = "all"
	SendOnClosed AnalysisCases = "sendOnClosed"
	// ReceiveOnClosed  AnalysisCases = "receiveOnClosed"
	DoneBeforeAdd    AnalysisCases = "doneBeforeAdd"
	CloseOnClosed    AnalysisCases = "closeOnClosed"
	ConcurrentRecv   AnalysisCases = "concurrentRecv"
	Leak             AnalysisCases = "leak"
	UnlockBeforeLock AnalysisCases = "unlockBeforeLock"
	MixedDeadlock    AnalysisCases = "mixedDeadlock"
	ResourceDeadlock AnalysisCases = "resourceDeadlock"
)

// ParseAnalysisCases parses the given analysis cases
//
// Returns:
//   - map[AnalysisCases]bool: A map of the analysis cases and if they are set
//   - error: An error if the cases could not be parsed
func ParseAnalysisCases() (map[AnalysisCases]bool, error) {
	analysisCases := map[AnalysisCases]bool{
		All:          false, // all cases enabled
		SendOnClosed: false,
		// ReceiveOnClosed:  false,
		DoneBeforeAdd:    false,
		CloseOnClosed:    false,
		ConcurrentRecv:   false,
		Leak:             false,
		UnlockBeforeLock: false,
		MixedDeadlock:    false,
		ResourceDeadlock: false,
	}

	if Scenarios == "-" {
		return analysisCases, nil
	}

	if Scenarios == "" {
		for c := range analysisCases {
			analysisCases[c] = true
		}

		// not in thesis
		analysisCases[ConcurrentRecv] = false
		// analysisCases[ReceiveOnClosed] = false

		return analysisCases, nil
	}

	for _, c := range Scenarios {
		switch c {
		case 's':
			analysisCases[SendOnClosed] = true
		case 'r':
			// analysisCases[ReceiveOnClosed] = true
		case 'w':
			analysisCases[DoneBeforeAdd] = true
		case 'n':
			analysisCases[CloseOnClosed] = true
		case 'b':
			analysisCases[ConcurrentRecv] = true
		case 'l':
			analysisCases[Leak] = true
		case 'u':
			analysisCases[UnlockBeforeLock] = true
		case 'c':
			analysisCases[ResourceDeadlock] = true
		// MIXED DEADLOCK [REMOVE]
		case 'm':
			analysisCases[MixedDeadlock] = true
		default:
			return nil, fmt.Errorf("Invalid analysis case: %c", c)
		}
	}

	all := true
	for key, val := range analysisCases {
		if key == All {
			continue
		}
		if !val {
			all = false
			break
		}
	}

	if all {
		analysisCases[All] = true
	}

	return analysisCases, nil
}
