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
)

func (this *window) setFuzzing() {
	this.traceSelect.disable()

	l := len(this.settings.components.scen)
	modes := make([]fyne.CanvasObject, l+3)
	modes[0] = this.settings.components.fuzzingMode.Container
	for i, s := range this.settings.components.scen {
		modes[i+1] = s.Container
	}
	modes[l+1] = this.settings.components.runStatic.Container
	modes[l+2] = this.settings.components.onlyAPanicAndLeak.Container

	limits := []fyne.CanvasObject{
		this.settings.components.timeout.Container,
		this.settings.components.toFuzzing.Container,
		this.settings.components.maxFuzzingRun.Container,
		this.settings.components.maxNumberElements.Container,
	}

	statistics := []fyne.CanvasObject{
		this.settings.components.createStatistics.Container,
		this.settings.components.measureTime.Container,
		this.settings.components.checkForNotExecuted.Container,
	}

	recordings := []fyne.CanvasObject{
		this.settings.components.recordPlus.Container,
		this.settings.components.ignoreCriticalSections.Container,
		this.settings.components.ignoreAtomics.Container,
	}

	behaviours := []fyne.CanvasObject{
		this.settings.components.noRewrite.Container,
		this.settings.components.deleteTrace.Container,
		this.settings.components.cont.Container,
		this.settings.components.alwaysPanic.Container,
		this.settings.components.noMemorySup.Container,
		this.settings.components.cancelTestIfBugFound.Container,
	}

	loggings := []fyne.CanvasObject{
		this.settings.components.verbose.Container,
		this.settings.components.noWarning.Container,
		this.settings.components.noProgress.Container,
		this.settings.components.output.Container,
	}

	this.setSettings(mode, modes)
	this.setSettings(limit, limits)
	this.setSettings(recording, recordings)
	this.setSettings(statistic, statistics)
	this.setSettings(behaviour, behaviours)
	this.setSettings(logging, loggings)
}
