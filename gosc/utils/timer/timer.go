// Copyright (c) 2024 
//
// File: timeMeasurement.go
// Brief: Timer to measure the times
//
//
// License: BSD-3-Clause

package timer

import (
	"time"
)

// Timer is a timer that can be started and stopped
//
// Parameter:
//   - startTime time.Time: time of the last start
//   - elapsedTime time.Duration: total elapsed time
//   - running bool: true if running, false if stopped
type Timer struct {
	startTime time.Time
	elapsed   time.Duration
	running   bool
}

// Start a timer
func (this *Timer) Start() {
	if this.running {
		return
	}

	this.startTime = time.Now()
	this.running = true
}

// Stop a timer
func (this *Timer) Stop() {
	if !this.running {
		return
	}
	this.elapsed += time.Since(this.startTime)
	this.running = false
}

// GetTime returns the elapsed time of the timer
//
// Returns:
//   - time.Duration: current elapsed time of timer
func (this *Timer) GetTime() time.Duration {
	if this.running {
		return this.elapsed + time.Since(this.startTime)
	}
	return this.elapsed
}

// Reset the timer
func (this *Timer) Reset() {
	this.running = false
	this.elapsed = time.Duration(0)
}

// IsRunning returns if the timer is currently running
func (this *Timer) IsRunning() bool {
	return this.running
}
