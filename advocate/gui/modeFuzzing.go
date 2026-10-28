// Copyright (c) 2026 Erik Kassubek
//
// File: record.go
// Brief: Gui for record
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func (this *window) setFuzzing() {
	var content []fyne.CanvasObject
	content = []fyne.CanvasObject{
		this.settings.components.mainTestSelect.Container,
		widget.NewSeparator(),
		this.settings.components.label.Container,
		this.settings.components.fuzzingMode.Container,
		this.settings.components.scen.Container,
		this.settings.components.timeout.Container,
		this.settings.components.toFuzzing.Container,
		this.settings.components.maxFuzzingRun.Container,
		this.settings.components.maxNumberElements.Container,
		twoCheck(this.settings.components.measureTime.Container, this.settings.components.createStatistics.Container),
		twoCheck(this.settings.components.checkForNotExecuted.Container, this.settings.components.ignoreCriticalSections.Container),
		twoCheck(this.settings.components.ignoreAtomics.Container, this.settings.components.onlyAPanicAndLeak.Container),
		twoCheck(this.settings.components.noRewrite.Container, this.settings.components.deleteTrace.Container),
		twoCheck(this.settings.components.cont.Container, this.settings.components.noWarning.Container),
		twoCheck(this.settings.components.noInfo.Container, this.settings.components.noProgress.Container),
		twoCheck(this.settings.components.output.Container, this.settings.components.alwaysPanic.Container),
		twoCheck(this.settings.components.noMemorySup.Container, this.settings.components.cancelTestIfBugFound.Container),
	}

	this.setSettings(content)
}
