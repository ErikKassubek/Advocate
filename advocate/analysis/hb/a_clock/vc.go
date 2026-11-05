// Copyright (c) 2024 Erik Kassubek
//
// File: clock.go
// Brief: Struct and functions of vector clocks vc
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_clock

import (
	"advocate/utils/consts"
	"advocate/utils/log"
	"fmt"
	"runtime"
	"strconv"
)

type VcType int

const (
	Weak VcType = iota
	Strong
)

// VectorClock is a vector clock
// Fields:
//
//   - size int: The size of the vector clock
//   - clock []int: The vector clock
type VectorClock struct {
	size  int
	clock map[uint32]uint32
}

// NewVectorClock creates and returns a new, empty vector clock
//
// Parameter:
//   - size int: The size of the vector clock
//
// Returns:
//   - *VectorClock: The new vector clock
func NewVectorClock(size int) *VectorClock {
	if size < 0 {
		size = 0
	}
	c := make(map[uint32]uint32)
	return &VectorClock{
		size:  size,
		clock: c,
	}
}

// NewVectorClockSet creates a new vector clock and set it
//
// Parameter:
//   - size int: The size of the vector clock
//   - cl (map[uint32]uint)32: The vector clock
//
// Returns:
//   - *VectorClock: A Pointer to the new vector clock
func NewVectorClockSet(size int, cl map[uint32]uint32) *VectorClock {
	vc := NewVectorClock(size)

	if cl == nil {
		return vc
	}

	if size < 0 {
		size = 0
	}

	for rout, val := range cl {
		if rout > uint32(size) {
			continue
		}
		vc.clock[rout] = val
	}

	return vc
}

// GetSize returns the size of the vector clock
//
// Returns:
//   - int: The size of the vector clock
func (this VectorClock) GetSize() int {
	return int(this.size)
}

// GetValue returns the value of the vector clock at a given index
//
// Parameter:
//   - index int: the index to get the value for
//
// Returns:
//   - uint32: the value at the given index
func (this *VectorClock) GetValue(index int) uint32 {
	if index > this.size {
		return 0
	}

	if val, ok := this.clock[uint32(index)]; ok {
		return val
	}
	return 0
}

// SetValue sets a value of the vector clock at a given index
//
// Parameter:
//   - index int: the index to set the value for
//   - value uint32: the new value
func (this *VectorClock) SetValue(index int, value uint32) {
	this.clock[uint32(index)] = value
}

// GetClock returns the vector clock
//
// Returns:
//   - map[uint32]uint32: The vector clock
func (this *VectorClock) GetClock() map[uint32]uint32 {
	return this.clock
}

// String returns a string representation of the vector clock
//
// Returns:
//   - string: The string representation of the vector clock
func (this *VectorClock) String() string {
	str := "["
	for i := 1; i <= this.size; i++ {
		str += fmt.Sprint(this.GetValue(i))
		if i <= this.size-1 {
			str += ", "
		}
	}
	str += "]"
	return str
}

// AsSlice returns the vector clock in a slice form
//
// Returns:
//   - []int: The vc as a slice
func (this *VectorClock) AsSlice() []int {
	vc := make([]int, this.size)
	for k, v := range this.clock {
		vc[k-1] = int(v)
	}

	return vc
}

// Inc increments the vector clock at the given position
//
// Parameter:
//   - routine int: The routine to increment
func (this *VectorClock) Inc(routine int) {
	if this == nil {
		return
	}

	if routine > int(this.size) {
		return
	}

	if this.clock == nil {
		this.clock = make(map[uint32]uint32)
	}

	this.clock[uint32(routine)]++
}

// Sync updates the vector clock with the received vector clock
//
// Parameter:
//   - rec *VectorClock: The received vector clock
//
// Returns:
//   - *VectorClock: The synced vc (not a copy)
func (this *VectorClock) Sync(rec *VectorClock) *VectorClock {
	if this == nil {
		this = rec.Copy()
		return this
	}

	if rec == nil {
		return this
	}

	if this.size == 0 && rec.size == 0 {
		_, file, line, _ := runtime.Caller(1)
		log.Error("Sync of empty vector clocks: " + file + consts.PosSep + strconv.Itoa(line))
	}

	if this.size == 0 {
		this = NewVectorClock(rec.size)
	}

	if rec.size == 0 {
		return this
	}

	for i := 1; i <= this.size; i++ {
		if rec.GetValue(i) > this.GetValue(i) {
			this.SetValue(i, rec.GetValue(i))
		}
	}

	return this
}

// Copy creates a copy of the vector clock
//
// Returns:
//   - *VectorClock: The copy of the vector clock
func (this *VectorClock) Copy() *VectorClock {
	if this == nil {
		return nil
	}

	newVc := NewVectorClock(this.size)
	for rout, val := range this.clock {
		newVc.clock[rout] = val
	}
	return newVc
}

// IsEqual checks if the the parameter vc2 is equal to the vc
func (this *VectorClock) IsEqual(vc2 *VectorClock) bool {
	if this.size != vc2.size {
		return false
	}

	for i := 1; i <= this.size; i++ {
		if this.GetValue(i) != vc2.GetValue(i) {
			return false
		}
	}

	return true
}

// IsMapVcEqual determines if two maps of vector clock are equal, meaning for
// each key they have the same vector clock as vale
//
// Parameter:
//   - v1 map[int]*VectorClock: vector clock 1
//   - v2 map[int]*VectorClock: vector clock 2
//
// Returns:
//   - bool: true if they are equal, false otherwise
func IsMapVcEqual(v1 map[int]*VectorClock, v2 map[int]*VectorClock) bool {
	if len(v1) != len(v2) {
		return false
	}

	for k, vc1 := range v1 {
		vc2, ok := v2[k]
		if !ok || !vc1.IsEqual(vc2) {
			return false
		}
	}

	return true
}
