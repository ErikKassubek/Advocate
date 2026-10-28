// Copyright (c) 2025 Erik Kassubek
//
// File: helper.go
// Brief: Helper functions for concurrency used by multiple methods
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_helper

import "advocate/trace"

// Valid filters out elements which do not correspond to valid operations, e.g.
// end of a routine
//
// Parameter:
//   - elem trace.Element: the element to test
//
// Returns:
//   - bool: true if the element is valid, false otherwise
func Valid(elem trace.Element) bool {
	t := elem.Type(false)
	return !(t == trace.Replay || t == trace.New || elem.Type(true) == trace.RoutEnd)
}
