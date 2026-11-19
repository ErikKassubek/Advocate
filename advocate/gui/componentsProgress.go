// Copyright (c) 2026 Erik Kassubek
//
// File: componentsMainTestSelector.go
// Brief: Create main/test selector
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"advocate/utils/math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentProgress struct {
	*fyne.Container

	progressWidget *widget.ProgressBar
}

func createProgressBar() *componentProgress {
	cms := &componentProgress{}

	cms.progressWidget = widget.NewProgressBar()

	cms.Container = container.NewVBox(
		cms.progressWidget,
	)

	cms.progressWidget.SetValue(0)

	return cms
}

func (this *componentProgress) setValue(v float64) {
	fyne.Do(func() {
		this.progressWidget.SetValue(v)
	})
}

func (this *componentProgress) setProgress(done int, total int) {
	if total == 0 {
		return
	}

	prog := float64(done) / float64(total)

	prog = math.Clamp(prog, 0, 1)

	this.progressWidget.SetValue(prog)
}
