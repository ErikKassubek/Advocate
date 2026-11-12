// Copyright (c) 2025 Erik Kassubek
//
// File: values.go
// Brief: Possible Values for modes
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package flags

// Modes
const (
	ModeAnalysis    = "analysis"
	ModeFuzzing     = "fuzzing"
	ModeRecord      = "record"
	ModeReplay      = "replay"
	ModeExperiments = "experiments"
	ModeShow        = "show"
)

// Closures
const (
	ClosurePF     = "pf"
	ClosureHB     = "hb"
	ClosureMHB    = "mhb"
	ClosureSCPC   = "scpc"
	ClosureReplay = "replay"
)
