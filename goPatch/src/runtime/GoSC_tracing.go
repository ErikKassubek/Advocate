// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: gosc_trace.go
// Brief: Functionality for tracing
//
//
// License: BSD-3-Clause

package runtime

var finishTracingFunc func()
var writeTraceToFileFunc func(routine int, fromRuntime bool) bool

var tracingStartNano int64

// InitTracing enables the collection of the trace
//
// Parameter:
//   - finishTracing func(): function injection for the gosc.FinishFuzzing function
//   - writeToTraceFile func(r int, f bool) bool
//   - init bool: true if called from init, false otherwise
func InitTracing(finishTracing func(), writeToTraceFile func(r int, f bool) bool, init bool) {
	GoscTracingDisabled = false
	finishTracingFunc = finishTracing
	writeTraceToFileFunc = writeToTraceFile
	setCurrentRoutineToActive()

	if tracingStartNano == 0 {
		tracingStartNano = nanotime()
	}
}

// DisableTracing disables the trace recording
func DisableTracing() {
	GoscTracingDisabled = true
}

// IsTracingEnabled returns whether tracing is enabled
//
// Returns:
//   - true if enabled, false otherwise
func IsTracingEnabled() bool {
	return !GoscTracingDisabled
}
