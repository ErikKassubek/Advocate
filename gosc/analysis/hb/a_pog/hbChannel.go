// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the vc for channels
//
//
// License: BSD-3-Clause

package a_pog

import (
	"gosc/analysis/a_base"
	"gosc/analysis/hb/a_clock"
	"gosc/trace"
	"gosc/utils/log"
)

// UpdateHBChannel updates the vector clocks to a channel element
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *trace.TraceElementChannel: the channel element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func UpdateHBChannel(graph *PoGraph, ch *trace.ElementChannel, recorded bool) {
	if recorded && !ch.Committed() {
		return
	}

	opC := ch.Type(true)
	cl := ch.GetClosed()

	if ch.IsBuffered() {
		switch opC {
		case trace.ChannelSend:
			Send(graph, ch)
		case trace.ChannelRecv:
			if cl { // recv on closed channel
				RecvC(graph, ch, true)
			} else {
				Recv(graph, ch)
			}
		case trace.ChannelClose:
			Close(graph, ch)
		default:
			err := "Unknown operation: " + ch.String()
			log.Error(err)
		}
	} else { // unbuffered channel
		switch opC {
		case trace.ChannelSend:
			partner := ch.GetPartner()
			if partner != nil {
				Unbuffered(graph, ch, partner)
				// increase index for recv is done in analysis/elements/channel.go
			}

		case trace.ChannelRecv: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				Unbuffered(graph, partner, ch)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if cl { // recv on closed channel
					RecvC(graph, ch, false)
				}
			}
		case trace.ChannelClose:
			Close(graph, ch)
		default:
			err := "Unknown operation: " + ch.String()
			log.Error(err)
		}
	}
}

// UpdateHBSelect stores and updates the vector clock of the select element.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - se *trace.TraceElementSelect: the select element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func UpdateHBSelect(graph *PoGraph, se *trace.ElementSelect, recorded bool) {
	noChannel := se.GetChosenDefault() || !se.Committed()

	if !noChannel {
		chosenCase := se.GetChosenCase()
		chosenCase.SetVc(a_clock.Strong, se.VC(a_clock.Strong))

		UpdateHBChannel(graph, chosenCase, recorded)
	}
}

// Unbuffered updates and calculates the vector clocks given a send/receive pair on a unbuffered
// channel.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - sender trace.Element: sender node
//   - recv trace.Element: receiver node
func Unbuffered(graph *PoGraph, sender trace.Element, recv trace.Element) {
	if sender.Committed() && recv.Committed() {
		if graph != nil {
			graph.AddEdge(sender, recv)
		} else {
			AddEdge(sender, recv, false)
		}
	}
}

// Send updates and calculates the pog given a send on a buffered channel.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *TraceElementChannel: The trace element
func Send(graph *PoGraph, ch *trace.ElementChannel) {
	if !ch.Committed() {
		return
	}

	gr := graph
	if graph == nil {
		gr = &po
	}

	id := ch.ResourceID()
	qSize := ch.GetQSize()
	qCount := ch.GetQCount()

	// direct communication without using the buffer
	if qCount == 0 {
		return
	}

	newBuffer(gr, id, qSize)

	count := qCount - 1

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(gr.chanBuffer[id]) >= count && len(gr.chanBuffer[id]) < gr.chanBufferSize[id] {
		gr.chanBuffer[id] = append(gr.chanBuffer[id], a_base.BufferedVC{
			Occupied: false,
			Send:     nil})
	}

	// if count > qSize || chanBuffer[id][count].Occupied {
	// 	log.Error("Write to occupied buffer position or to big count")
	// }

	s := gr.chanBuffer[id][count].Send
	if s != nil {
		if graph != nil {
			graph.AddEdge(s, ch)
		} else {
			AddEdge(s, ch, false)
		}
	}

	gr.chanBuffer[id][count] = a_base.BufferedVC{
		Occupied: true,
		Send:     ch,
	}
}

// Recv updates and calculates the vector clocks given a receive on a buffered channel.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *TraceElementChannel: The trace element
func Recv(graph *PoGraph, ch *trace.ElementChannel) {
	if !ch.Committed() {
		return
	}

	gr := graph
	if graph == nil {
		gr = &po
	}

	id := ch.ResourceID()
	qSize := ch.GetQSize()

	newBuffer(gr, id, qSize)

	s := gr.chanBuffer[id][0].Send

	if s != nil {
		if graph != nil {
			graph.AddEdge(s, ch)
		} else {
			AddEdge(s, ch, false)
		}
	}

	gr.chanBuffer[id] = append(gr.chanBuffer[id][1:], a_base.BufferedVC{
		Occupied: false,
		Send:     nil,
	})
}

// RecvC updates and calculates the vector clocks given a receive on a closed channel.
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *TraceElementChannel: The trace element
//   - buffered bool: true if the channel is buffered
func RecvC(graph *PoGraph, ch *trace.ElementChannel, buffered bool) {
	if !ch.Committed() {
		return
	}

	id := ch.ResourceID()

	if graph != nil {
		if _, ok := graph.closeData[id]; ok {
			c := graph.closeData[id]
			graph.AddEdge(c, ch)
		} else {
			if _, ok := po.closeData[id]; ok {
				c := po.closeData[id]
				AddEdge(c, ch, false)
			}
		}
	}
}

// Close records the close for the pog
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - ch *TraceElementChannel: The trace element
func Close(graph *PoGraph, ch *trace.ElementChannel) {
	if graph != nil {
		graph.closeData[ch.ResourceID()] = ch
	} else {
		po.closeData[ch.ResourceID()] = ch
	}
}

// Create a new map of buffered vector clocks for a channel if not already in
// bufferedVCs.
//
// Parameter:
//   - graph *PoGraph: the graph to add it in
//   - id int: the id of the channel
//   - qSize int: the buffer qSize of the channel
func newBuffer(graph *PoGraph, id int, qSize int) {
	if _, ok := graph.chanBuffer[id]; !ok {
		graph.chanBuffer[id] = make([]a_base.BufferedVC, 1)
		graph.chanBufferSize[id] = qSize
		graph.chanBuffer[id][0] = a_base.BufferedVC{
			Occupied: false,
			Send:     nil,
		}
	}
}
