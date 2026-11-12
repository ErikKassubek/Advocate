// Copyright (c) 2025 
//
// File: channel.go
// Brief: Constraints for channels
//
//
// License: BSD-3-Clause

package a_constraints

import "gosc/trace"

// AddChannel creates constraints for unbuffered channels
// If creates a positive constraint between the send and corresponding recv
//
// Parameter:
//   - elem *trace.ElementChannel: the element
func AddChannel(elem *trace.ElementChannel) {
	// For now, only create constraints for unbuffered channels
	if elem.IsBuffered() {
		return
	}

	if elem.Type(true) == trace.ChannelRecv {
		p := elem.GetPartner()
		AddConstraint(true, p, elem)
	}
}
