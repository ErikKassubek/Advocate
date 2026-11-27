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
	"fyne.io/fyne/v2/container"
)

type settingsRecord struct {
	*fyne.Container
}

func createSettingsRecord(comp *settingComponents) *settingsRecord {
	sr := settingsRecord{}

	sr.Container = container.NewVBox()

	return &sr
}

func (this *window) setRecord() {
	this.traceSelect.disable()

	modes := []fyne.CanvasObject{}

	limits := []fyne.CanvasObject{
		this.settings.components.timeout.Container,
		this.settings.components.maxNumberElements.Container,
	}

	statistics := []fyne.CanvasObject{
		this.settings.components.createStatistics.Container,
		this.settings.components.measureTime.Container,
		this.settings.components.checkForNotExecuted.Container,
	}

	recordings := []fyne.CanvasObject{
		this.settings.components.recordPlus.Container,
		this.settings.components.ignoreAtomics.Container,
	}

	behaviours := []fyne.CanvasObject{
		this.settings.components.deleteTrace.Container,
		this.settings.components.cont.Container,
		this.settings.components.alwaysPanic.Container,
		this.settings.components.noMemorySup.Container,
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
