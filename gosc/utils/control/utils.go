// Copyright (c) 2025 
//
// File: helper.go
// Brief: Utils for memory
//
//
// License: BSD-3-Clause

package control

import (
	"gosc/utils/log"
	"runtime"
)

// printAllGoroutines prints the stack traces of all routines
func printAllGoroutines() {
	buf := make([]byte, 1<<20) // 1 MB buffer
	n := runtime.Stack(buf, true)
	log.Infof("%s", buf[:n])
}
