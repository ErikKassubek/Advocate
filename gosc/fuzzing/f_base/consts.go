// Copyright (c) 2025 
//
// File: consts.go
// Brief: Consts for fuzzing
//
//
// License: BSD-3-Clause

package f_base

// Possible values for fuzzing mode
const (
	Default     = ""            // not set
	GFuzz       = "GFuzz"       // only GFuzz
	GFuzzHB     = "GFuzzHB"     // GFuzz with use of hb info
	GFuzzHBFlow = "GFuzzHBFlow" // GFuzz with use of hb info and flow mutation
	GoPie       = "GoPie"       // only goPie
	GoCR        = "GoCR"        // improved goPie without predictive analysis
	Full        = "Full"        // improved goPie with predictive analysis
	Guided      = "Guided"      // hb guided fuzzing

)

var Modes = []string{GFuzz, GFuzzHB, GFuzzHBFlow, GoPie, GoCR, Full, Guided}

// Possible mut types
const (
	MutSelType  = 0
	MutPiType   = 1
	MutFlowType = 2
)
