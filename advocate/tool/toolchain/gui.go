// Copyright (c) 2026 Erik Kassubek
//
// File: gui.go
// Brief: communication with gui
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package toolchain

import (
	"advocate/analysis/a_analysis"
	"advocate/utils/math"
	"advocate/utils/types"
)

var GuiCom chan types.Pair[string, float64] = nil

var canceledGui bool

func calcPerc(done, total int) float64 {
	if total == 0 {
		return 0
	}

	prog := float64(done) / float64(total)

	return math.Clamp(prog, 0, 1)

}

func guiSetProgressTotalPerc(perc float64) {
	if GuiCom == nil {
		return
	}

	GuiCom <- types.NewPair("progressTotal", perc)
}

func guiSetProgressTotal(done, total int) {
	if GuiCom == nil {
		return
	}

	GuiCom <- types.NewPair("progressTotal", calcPerc(done, total))
}

func guiSetProgressOne(done, total int) {
	if GuiCom == nil {
		return
	}

	GuiCom <- types.NewPair("progressOne", calcPerc(done, total))
}

func CanceledGui() bool {
	return canceledGui
}

func Cancel() {
	canceledGui = true
	a_analysis.GuiCanceled = true
}

func ResetCancel() {
	canceledGui = false
	a_analysis.GuiCanceled = false
}
