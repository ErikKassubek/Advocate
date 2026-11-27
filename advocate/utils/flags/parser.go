// Copyright (c) 2025 Erik Kassubek
//
// File: parser.go
// Brief: Parse flags
//
// Author: Erik Kassubek
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

// analysis cases to run
var AnalysisCasesMap = make(map[AnalysisCases]bool)

// ParseAnalysisCases parses the given analysis cases
//
// Returns:
//   - map[AnalysisCases]bool: A map of the analysis cases and if they are set
//   - error: An error if the cases could not be parsed
func ParseAnalysisCases() error {
	AnalysisCasesMap = map[AnalysisCases]bool{
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
		return nil
	}

	if Scenarios == "" {
		for c := range AnalysisCasesMap {
			AnalysisCasesMap[c] = true
		}

		return nil
	}

	for _, c := range Scenarios {
		switch c {
		case 's':
			AnalysisCasesMap[SendOnClosed] = true
		case 'w':
			AnalysisCasesMap[DoneBeforeAdd] = true
		case 'c':
			AnalysisCasesMap[CloseOnClosed] = true
		case 'l':
			AnalysisCasesMap[Leak] = true
		case 'r':
			AnalysisCasesMap[ConcurrentRecv] = true
		case 'u':
			AnalysisCasesMap[UnlockBeforeLock] = true
		case 'd':
			AnalysisCasesMap[ResourceDeadlock] = true
		case 'm':
			AnalysisCasesMap[MixedDeadlock] = true
		default:
			return fmt.Errorf("Invalid analysis case: %c", c)
		}
	}

	all := true
	for key, val := range AnalysisCasesMap {
		if key == All {
			continue
		}
		if !val {
			all = false
			break
		}
	}

	if all {
		AnalysisCasesMap[All] = true
	}

	return nil
}
