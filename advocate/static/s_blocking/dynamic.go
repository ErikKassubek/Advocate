// Copyright (c) 2026 Erik Kassubek
//
// File: dynamic.go
// Brief: The dynamic parts of the blocking detection
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/analysis/a_base"
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
)

// getBlockedResources returns the resources of which at least one routine is blocking
//
// Returns:
//   - map[int]trace.Resource: blocked object id to blocked resource.
//   - []trace.Resource: list of blocked resources
func getBlockedResources() (map[trace.Element][]trace.Resource, map[int]trace.Resource) {
	res := make(map[trace.Element][]trace.Resource)

	blockedRes := make(map[int]trace.Resource)

	for _, e := range a_base.MainTrace.GetBlocked() {
		// TODO: handle select
		resources := a_base.MainTrace.GetResources(e)

		for _, r := range resources {
			res[e] = append(res[e], r)
			blockedRes[r.Id()] = r
		}
	}

	// TODO: added to trace all variables, remove
	if false {
		return res, a_base.MainTrace.Resources()
	}

	return res, blockedRes
}

func buildFuncCallToSSAFunc() {
	blocking.funcCallToSSAFunc = make(map[*trace.ElementFunc]*s_ssa.Function)

	for f := range a_base.MainTrace.CallTree().GetTree() {
		fn := s_ssa.GetSSAFuncFromName(data.Ssa(), f.GetSSAName())
		if fn == nil {
			log.Errorf("Could not find ssa function for %s", f.GetSSAName())
			continue
		}
		blocking.funcCallToSSAFunc[f] = fn
	}
}
