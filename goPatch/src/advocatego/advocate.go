// Copyright (c) 2026 Erik Kassubek
//
// File: advocate.go
// Brief: Func for advocate
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package advocatego

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
			println("TIMEOUT")
			panic("Timeout")
		}()
	}
}

