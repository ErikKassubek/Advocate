// GOSC-FILE_START

// Copyright (c) 2026 
//
// File: gosc_time.go
// Brief: Set of functions using time
//
//
// License: BSD-3-Clause

package runtime

// Sleep the routine for a given time
//
// Parameter:
//   - seconds float64: seconds to sleep for
func sleep(seconds float64) {
	timeSleep(sToNs(seconds))
}

// Convert seconds to nanoseconds
//
// Parameter:
//   - seconds float64
//
// Returns
//   - int64 seconds in nanoseconds
func sToNs(seconds float64) int64 {
	return int64(seconds * 1e9)
}

// check if a certain time has passed from a given start time
//
// Parameter:
//   - startTime int64: the start time in nano seconds
//   - durationS int64: the duration in seconds
//
// Returns:
//   - true if the difference between startNS and the current time is at least durationS
func hasTimePast(startNs int64, durationS int64) bool {
	durationNano := durationS * 1e9
	return nanotime()-startNs >= durationNano
}

// Get the current time in nanoseconds
//
// Returns:
//   - int64: the time in nanoseconds
func currentTime() int64 {
	return nanotime()
}

// The following is mostly a copy and adaption of functions in time/sleep.go

// type for a timer
//
// Fields:
//   - C <-chan struct{}: channel to fire on
//   - r *timeTimer: the timer
type afterTimer struct {
	C <-chan struct{}
	r *timeTimer
}

// Create a new timer
//
// Parameter:
//
//	d int64: time in ns until fire
//
// Returns:
//   - *afterTimer: the timer
func newTimer2(d int64) *afterTimer {
	c := make(chan struct{}, 1)
	t := &afterTimer{
		C: c,
		r: newTimer(when(d), 0, sendTime, c, nil),
	}
	return t
}

// function that is executed when the timer fires
// Needed for *timeTimer in afterTimer
//
// Parameter:
//
//	see runtime/time.go -> newTimer
func sendTime(c any, seq uintptr, delay int64) {
	select {
	case c.(chan struct{}) <- struct{}{}:
	default:
	}
}

// Get the nanotime value that is nano ns in the future
//
// Parameter:
//   - nano int64: difference between now and result of when in ns
//
// Returns:
//   - int64 nanotime that is nano ns in the future, if nano = 0 return current nanotime,
//     if this value would be 0 because of overflow, it is set to the max int64 value
func when(nano int64) int64 {
	if nano <= 0 {
		return currentTime()
	}
	t := currentTime() + int64(nano)
	if t < 0 {
		// N.B. runtimeNano() and d are always positive, so addition
		// (including overflow) will never result in t == 0.
		t = 1<<63 - 1 // math.MaxInt64
	}
	return t
}

// Get a channel to wait on. The channel will send after a given time
//
// Parameter:
//   - d int64: time in ns until fire
//
// Returns:
//
//	<-chan struct{}: after d ns, something will be send on this channel
func after(d int64) <-chan struct{} {
	return newTimer2(d).C
}

// Sleep busy
func SleepBusy(dInSec int) {
	dInNS := sToNs(float64(dInSec))
	startTime := currentTime()
	for {
		duration := currentTime() - startTime
		if duration >= dInNS {
			return
		}
	}
}
