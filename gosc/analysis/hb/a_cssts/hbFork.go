// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the cssts for fork
//
//
// License: BSD-3-Clause

package a_cssts

import (
	"gosc/trace"
)

// UpdateHBFork update and calculate happens before information for fork operations
//
// Parameter:
//   - fo *TraceElementFork: the fork element
func UpdateHBFork(fo *trace.ElementFork) {
	AddEdgeFork(fo)
}
