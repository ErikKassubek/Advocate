// Copyright (c) 2026
//
// File: gosc.go
// Brief: Func for gosc
//
//
// License: BSD-3-Clause

package gosc

import "time"

var FinishFunc func()
var timeoutStarted = false

func startTimeout(timeout int) {
	if timeoutStarted {
		return
	}

	timeoutStarted = true
	if timeout > 0 {
		// start time timeout
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			panic("Timeout")
		}()
	}
}
