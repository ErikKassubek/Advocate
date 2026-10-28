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
	"fyne.io/fyne/v2/widget"
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
	objects := []fyne.CanvasObject{
		this.settings.components.mainTestSelect.Container,
		widget.NewSeparator(),
		this.settings.components.label.Container,
		this.settings.components.label,
		this.settings.components.timeout.Container,
		this.settings.components.maxNumberElements.Container,
		twoCheck(this.settings.components.measureTime.Container, this.settings.components.createStatistics.Container),
		twoCheck(this.settings.components.checkForNotExecuted.Container, this.settings.components.ignoreAtomics.Container),
		twoCheck(this.settings.components.deleteTrace.Container, this.settings.components.cont.Container),
		twoCheck(this.settings.components.noWarning.Container, this.settings.components.noInfo.Container),
		twoCheck(this.settings.components.noProgress.Container, this.settings.components.output.Container),
		twoCheck(this.settings.components.alwaysPanic.Container, this.settings.components.noMemorySup.Container),
	}
	this.setSettings(objects)
}
