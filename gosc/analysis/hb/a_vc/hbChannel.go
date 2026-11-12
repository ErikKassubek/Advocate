// Copyright (c) 2025 
//
// File: hbAtomic.go
// Brief: Update the vc for channels
//
//
// License: BSD-3-Clause

package a_vc

import (
	"gosc/analysis/a_base"
	"gosc/analysis/hb/a_clock"
	"gosc/trace"
	"gosc/utils/log"
)

// UpdateHBChannel updates the vector clocks to a channel element
//
// Parameter:
//   - ch *trace.TraceElementChannel: the channel element
func UpdateHBChannel(ch *trace.ElementChannel) {
	routine := ch.RoutineID()

	ch.SetVc(a_clock.Strong, CurrentVC[routine])
	ch.SetVc(a_clock.Weak, CurrentWVC[routine])

	if !ch.Committed() {
		return
	}

	opC := ch.Type(true)
	cl := ch.GetClosed()

	if ch.IsBuffered() {
		switch opC {
		case trace.ChannelSend:
			Send(ch)
		case trace.ChannelRecv:
			if cl { // recv on closed channel
				RecvC(ch, true)
			} else {
				Recv(ch, CurrentVC, CurrentWVC)
			}
		case trace.ChannelClose:
			Close(ch)
		default:
			err := "Unknown operation: " + ch.String()
			log.Error(err)
		}
	} else { // unbuffered channel
		switch opC {
		case trace.ChannelSend:
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.RoutineID()
				partner.SetVc(a_clock.Strong, CurrentVC[partnerRout])
				sel := partner.GetSelect()
				if sel != nil {
					sel.SetVc(a_clock.Strong, CurrentVC[partnerRout])
				}
				Unbuffered(ch, partner)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if !cl {
					StuckChan(routine)
				}
			}

		case trace.ChannelRecv: // should not occur, but better save than sorry
			partner := ch.GetPartner()
			if partner != nil {
				partnerRout := partner.RoutineID()
				partner.SetVc(a_clock.Strong, CurrentVC[partnerRout])
				Unbuffered(partner, ch)
				// increase index for recv is done in analysis/elements/channel.go
			} else {
				if cl { // recv on closed channel
					RecvC(ch, false)
				} else {
					StuckChan(routine)
				}
			}
		case trace.ChannelClose:
			Close(ch)
		default:
			err := "Unknown operation: " + ch.String()
			log.Error(err)
		}
	}
}

// UpdateHBSelect stores and updates the vector clock of the select element.
//
// Parameter:
//   - se *trace.TraceElementSelect: the select element
func UpdateHBSelect(se *trace.ElementSelect) {
	noChannel := se.GetChosenDefault() || !se.Committed()

	routine := se.RoutineID()

	se.SetVc(a_clock.Strong, CurrentVC[routine])
	se.SetVc(a_clock.Weak, CurrentVC[routine])

	if noChannel {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
	} else {
		chosenCase := se.GetChosenCase()
		chosenCase.SetVc(a_clock.Strong, se.VC(a_clock.Strong))

		UpdateHBChannel(chosenCase)
	}

	cases := se.GetCases()

	for _, c := range cases {
		c.SetVc(a_clock.Strong, se.VC(a_clock.Strong))
	}
}

// Unbuffered updates and calculates the vector clocks given a send/receive pair on a unbuffered
// channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - routSend int: the route of the sender
//   - routRecv int: the route of the receiver
//   - tID_send string: the position of the send in the program
//   - tID_recv string: the position of the receive in the program
func Unbuffered(sender trace.Element, recv trace.Element) {
	if sender.Committed() && recv.Committed() {
		CurrentVC[recv.RoutineID()].Sync(CurrentVC[sender.RoutineID()])
		CurrentVC[sender.RoutineID()] = CurrentVC[recv.RoutineID()].Copy()
		CurrentVC[sender.RoutineID()].Inc(sender.RoutineID())
		CurrentVC[recv.RoutineID()].Inc(recv.RoutineID())
		CurrentWVC[sender.RoutineID()].Inc(sender.RoutineID())
		CurrentWVC[recv.RoutineID()].Inc(recv.RoutineID())
	} else {
		CurrentVC[sender.RoutineID()].Inc(sender.RoutineID())
		CurrentWVC[sender.RoutineID()].Inc(sender.RoutineID())
	}
}

// Send updates and calculates the vector clocks given a send on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func Send(ch *trace.ElementChannel) {
	routine := ch.RoutineID()

	if !ch.Committed() {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
		return
	}

	id := ch.ResourceID()
	qSize := ch.GetQSize()
	qCount := ch.GetQCount()

	r := a_base.MostRecentSend[routine][id]
	if r.Elem != nil {
		CurrentVC[routine].Sync(r.Vc)
	}

	// direct communication without using the buffer
	if qCount == 0 {
		return
	}

	newBuffer(id, qSize)

	count := qCount - 1

	// if the buffer size of the channel is very big, it would be a wast of RAM to create a map that could hold all of then, especially if
	// only a few are really used. For this reason, only the max number of buffer positions used is allocated.
	// If the map is full, but the channel has more buffer positions, the map is extended
	if len(chanBuffer[id]) >= count {
		chanBuffer[id] = append(chanBuffer[id], a_base.BufferedVC{
			Occupied: false,
			Send:     nil})
	}

	// if count > qSize || chanBuffer[id][count].Occupied {
	// 	log.Error("Write to occupied buffer position or to big count")
	// }

	s := chanBuffer[id][count].Send
	if s != nil {
		v := s.VC(a_clock.Strong)
		CurrentVC[routine].Sync(v)
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)

	chanBuffer[id][count] = a_base.BufferedVC{
		Occupied: true,
		Send:     ch,
	}
}

// Recv updates and calculates the vector clocks given a receive on a buffered channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - vc map[int]*VectorClock: the current vector clocks
//   - wVc map[int]*VectorClock: the current weak vector clocks
func Recv(ch *trace.ElementChannel, vc, wVc map[int]*a_clock.VectorClock) {
	id := ch.ResourceID()
	routine := ch.RoutineID()
	qSize := ch.GetQSize()

	if !ch.Committed() {
		vc[routine].Inc(routine)
		wVc[routine].Inc(routine)
		return
	}

	newBuffer(id, qSize)
	s := chanBuffer[id][0].Send

	if s != nil {
		vc[routine] = vc[routine].Sync(s.VC(a_clock.Strong))
	}

	r := a_base.MostRecentReceive[routine][id]
	if r.Elem != nil {
		vc[routine] = vc[routine].Sync(r.Vc)
	}

	chanBuffer[id] = append(chanBuffer[id][1:], a_base.BufferedVC{
		Occupied: false,
		Send:     nil,
	})

	vc[routine].Inc(routine)
	wVc[routine].Inc(routine)
}

// StuckChan updates and calculates the vector clocks for a stuck channel element
//
// Parameter:
//   - routine int: the route of the operation
func StuckChan(routine int) {
	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// Close updates and calculates the vector clocks given a close on a channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
func Close(ch *trace.ElementChannel) {
	if !ch.Committed() {
		return
	}

	routine := ch.RoutineID()

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// RecvC updates and calculates the vector clocks given a receive on a closed channel.
//
// Parameter:
//   - ch *TraceElementChannel: The trace element
//   - buffered bool: true if the channel is buffered
func RecvC(ch *trace.ElementChannel, buffered bool) {
	if !ch.Committed() {
		return
	}

	id := ch.ResourceID()
	routine := ch.RoutineID()

	if _, ok := a_base.CloseData[id]; ok {
		c := a_base.CloseData[id]
		CurrentVC[routine].Sync(c.VC(a_clock.Strong))
	}

	CurrentVC[routine].Inc(routine)
	CurrentVC[routine].Inc(routine)
}

// Create a new map of buffered vector clocks for a channel if not already in
// baseA.BufferedVCs.
//
// Parameter:
//   - id int: the id of the channel
//   - qSize int: the buffer qSize of the channel
func newBuffer(id int, qSize int) {
	if _, ok := chanBuffer[id]; !ok {
		chanBuffer[id] = make([]a_base.BufferedVC, 1)
		chanBufferSize[id] = qSize
		chanBuffer[id][0] = a_base.BufferedVC{
			Occupied: false,
			Send:     nil,
		}
	}
}
