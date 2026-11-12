// Copyright (c) 2025 
//
// File: flags.go
// Brief: Command line flags
//
//
// License: BSD-3-Clause

package flags

// Paths and names
var (
	ProgPath  string
	RootPath  string
	TracePath string

	ProgName string
	ExecName string
)

// Modes
var (
	Mode        string
	ModeMain    bool
	FuzzingMode string
)

// timeouts and limits
var (
	Timeout        int
	TimeoutFuzzing int
	MaxFuzzingRun  int

	MaxNumberElements int
)

// recording
var (
	RecordPlus bool
)

// logging
var (
	Output     bool
	NoInfo     bool
	NoProgress bool
	NoWarning  bool
)

// statistics
var (
	MeasureTime      bool
	CreateStatistics bool
	NotExecuted      bool
	CheckBenign      bool
)

// memory and panic
var (
	NoMemorySupervisor bool
	AlwaysPanic        bool
)

// settings
var (
	IgnoreAtomics         bool
	IgnoreCriticalSection bool

	OnlyAPanicAndLeak bool

	CancelTestIfBugFound bool

	Settings string

	Scenarios string

	ShowFile bool

	Experiment string

	ClosureType      string
	DisableStableIDs bool
)

// execution control
var (
	Continue      bool
	SkipExisting  bool
	NoRewrite     bool
	NoSkipRewrite bool
	DeleteTraces  bool
)

// contains gui
var (
	ContainsGui bool
)
