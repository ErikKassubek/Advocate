// Copyright (c) 2024 Erik Kassubek
//
// File: select.go
// Brief: File for the selects for fuzzing
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package f_gfuzz

import (
	"advocate/fuzzing/f_base"
	"advocate/trace"
	"sort"
)

// AddFuzzingSelect adds a select to selectInfoTrace
//
// Parameter:
//   - e *trace.TraceElementSelect: the select trace element to add
func AddFuzzingSelect(e *trace.ElementSelect) {
	fs := f_base.FuzzingSelect{
		ID:              e.ReplayID(),
		T:               e.T(trace.Commit),
		ChosenCase:      e.GetChosenIndex(),
		NumberCases:     len(e.GetCases()),
		ContainsDefault: e.GetContainsDefault(),
		CasiWithPos:     e.GetCasiWithPosPartner(),
	}

	SelectInfoTrace[fs.ID] = append(SelectInfoTrace[fs.ID], fs)
	NumberSelects++
}

// SortSelects sorts the list of occurrences of each select by the time value
func SortSelects() {
	for key := range SelectInfoTrace {
		sort.Slice(SelectInfoTrace[key], func(i, j int) bool {
			return SelectInfoTrace[key][i].T < SelectInfoTrace[key][j].T
		})
	}
}
