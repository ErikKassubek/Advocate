// Copyright (c) 2024 Erik Kassubek
//
// File: analysisSelectPartner.go
// Brief: Trace analysis for detection of select cases without any possible partners
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_scenarios

import (
	"advocate/analysis/a_base"
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
	"advocate/utils/timer"
)

// CheckForSelectCaseWithPartner checks for select cases with a valid
// partner. Call when all elements have been processed.
func CheckForSelectCaseWithPartner() {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	// check if not selected cases could be partners
	for i, c1 := range a_base.SelectCases {
		for j := i + 1; j < len(a_base.SelectCases); j++ {
			c2 := a_base.SelectCases[j]

			// if c1.partnerFound && c2.partnerFound {
			// 	continue
			// }

			if c1.ChanID != c2.ChanID || c1.Elem.Elem.ID() == c2.Elem.Elem.ID() || c1.Send == c2.Send {
				continue
			}

			if c2.Send { // c1 should be send, c2 should be recv
				c1, c2 = c2, c1
			}

			hbInfo := a_clock.GetHappensBefore(c1.Elem.Vc, c2.Elem.Vc)
			found := false
			if c1.Buffered && (hbInfo == a_hb.Concurrent || hbInfo == a_hb.After) {
				found = true
			} else if !c1.Buffered && hbInfo == a_hb.Concurrent {
				found = true
			}

			if found {
				a_base.SelectCases[i].PartnerFound = true
				a_base.SelectCases[j].PartnerFound = true
				a_base.SelectCases[i].Partner = append(a_base.SelectCases[i].Partner, a_base.ElemWithVcVal{
					Elem: a_base.SelectCases[j].Sel,
					Vc:   a_base.SelectCases[j].Sel.GetVC(a_clock.Strong),
					Val:  0,
				})
				a_base.SelectCases[j].Partner = append(a_base.SelectCases[j].Partner, a_base.ElemWithVcVal{
					Elem: a_base.SelectCases[i].Sel,
					Vc:   a_base.SelectCases[i].Sel.GetVC(a_clock.Strong),
					Val:  0,
				})
			}
		}
	}

	if len(a_base.SelectCases) == 0 {
		return
	}

	// collect all cases with no partner and all not triggered cases with partner

	for _, c := range a_base.SelectCases {
		opjType := "C"
		if c.Send {
			opjType += "S"
		} else {
			opjType += "R"
		}

		if c.PartnerFound {
			c.Sel.AddCasesWithPosPartner(c.Casi)
			a_base.NumberSelectCasesWithPartner++
		}
	}
}

// CheckForSelectCaseWithPartnerSelect checks for select cases with a valid
// partner. Call whenever a select is processed.
//
// Parameter:
//   - se *TraceElementSelect: The trace elem
//   - vc *VectorClock: The vector clock
func CheckForSelectCaseWithPartnerSelect(se *trace.ElementSelect, vc *a_clock.VectorClock) {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for casi, c := range se.GetCases() {

		id := c.ResourceID()

		buffered := (c.GetQSize() > 0)
		send := (c.Type(true) == trace.ChannelSend)

		found := false
		executed := false
		var partner = make([]a_base.ElemWithVcVal, 0)

		if casi == se.GetChosenIndex() && se.Committed() {
			// no need to check if the channel is the chosen case
			executed = true
			p := se.GetPartner()
			if p != nil {
				found = true
				vcTID := a_base.ElemWithVcVal{
					Elem: p,
					Vc:   p.GetVC(a_clock.Strong).Copy(),
					Val:  0,
				}
				partner = append(partner, vcTID)
			}
		} else {
			// not select cases
			if send {
				for _, mrr := range a_base.MostRecentReceive {
					if possiblePartner, ok := mrr[id]; ok {
						hbInfo := a_clock.GetHappensBefore(vc, possiblePartner.Vc)
						if buffered && (hbInfo == a_hb.Concurrent || hbInfo == a_hb.Before) {
							found = true
							partner = append(partner, possiblePartner)
						} else if !buffered && hbInfo == a_hb.Concurrent {
							found = true
							partner = append(partner, possiblePartner)
						}
					}
				}
			} else { // recv
				for _, mrs := range a_base.MostRecentSend {
					if possiblePartner, ok := mrs[id]; ok {
						hbInfo := a_clock.GetHappensBefore(vc, possiblePartner.Vc)
						if buffered && (hbInfo == a_hb.Concurrent || hbInfo == a_hb.After) {
							found = true
							partner = append(partner, possiblePartner)
						} else if !buffered && hbInfo == a_hb.Concurrent {
							found = true
							partner = append(partner, possiblePartner)
						}
					}
				}
			}
		}

		a_base.SelectCases = append(a_base.SelectCases,
			a_base.AllSelectCase{Sel: se,
				ChanID: id,
				Elem: a_base.ElemWithVc{
					Vc:   vc,
					Elem: se,
				},
				Send:         send,
				Buffered:     buffered,
				PartnerFound: found,
				Partner:      partner,
				Exec:         executed,
				Casi:         casi})

	}
}

// CheckForSelectCaseWithPartnerChannel checks for select cases with a valid
// partner. Call whenever a channel operation is processed.
//
// Parameter:
//   - id int: The id of the channel
//   - vc VectorClock: The vector clock
//   - send bool: True if the operation is a send
//   - buffered bool: True if the channel is buffered
func CheckForSelectCaseWithPartnerChannel(ch trace.Element, vc *a_clock.VectorClock,
	send bool, buffered bool) {

	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for i, c := range a_base.SelectCases {
		if c.PartnerFound || c.ChanID != ch.ResourceID() || c.Send == send || c.Elem.Elem.ID() == ch.ID() {
			continue
		}

		hbInfo := a_clock.GetHappensBefore(vc, c.Elem.Vc)
		found := false
		if send {
			if buffered && (hbInfo == a_hb.Concurrent || hbInfo == a_hb.Before) {
				found = true
			} else if !buffered && hbInfo == a_hb.Concurrent {
				found = true
			}
		} else {
			if buffered && (hbInfo == a_hb.Concurrent || hbInfo == a_hb.After) {
				found = true
			} else if !buffered && hbInfo == a_hb.Concurrent {
				found = true
			}
		}

		if found {
			a_base.SelectCases[i].PartnerFound = true
			a_base.SelectCases[i].Partner = append(a_base.SelectCases[i].Partner, a_base.ElemWithVcVal{
				Elem: ch,
				Vc:   vc,
				Val:  0,
			})
		}
	}
}

// CheckForSelectCaseWithPartnerClose checks for select cases without a valid
// partner. Call whenever a close operation is processed.
//
// Parameter:
//   - id int: The id of the channel
//   - vc VectorClock: The vector clock
func CheckForSelectCaseWithPartnerClose(cl *trace.ElementChannel, vc *a_clock.VectorClock) {
	timer.Start(timer.AnaSelWithoutPartner)
	defer timer.Stop(timer.AnaSelWithoutPartner)

	for i, c := range a_base.SelectCases {
		if c.PartnerFound || c.ChanID != cl.ResourceID() || c.Send {
			continue
		}

		hbInfo := a_clock.GetHappensBefore(vc, c.Elem.Vc)
		found := false
		if c.Buffered && (hbInfo == a_hb.Concurrent || hbInfo == a_hb.After) {
			found = true
		} else if !c.Buffered && hbInfo == a_hb.Concurrent {
			found = true
		}

		if found {
			a_base.SelectCases[i].PartnerFound = true
			a_base.SelectCases[i].Partner = append(a_base.SelectCases[i].Partner, a_base.ElemWithVcVal{
				Elem: cl,
				Vc:   vc,
				Val:  0,
			})
		}
	}
}

// RerunCheckForSelectCaseWithPartnerChannel reruns the
// CheckForSelectCaseWithPartnerChannel for all channel. This
// is needed to find potential communication partners for not executed
// select cases, if the select was executed after the channel
func RerunCheckForSelectCaseWithPartnerChannel() {
	for _, routine := range a_base.MainTrace.Routines() {
		for _, elem := range routine.Elems() {
			if e, ok := elem.(*trace.ElementChannel); ok {
				CheckForSelectCaseWithPartnerChannel(e, e.GetVC(a_clock.Strong),
					e.Type(true) == trace.ChannelSend, e.IsBuffered())
			}
		}
	}
}
