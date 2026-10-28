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
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type componentPathSelector struct {
	*fyne.Container

	label *componentSectionLabel

	selectPath        *fyne.Container
	selectedPathLabel *widget.Label
	openPathSelButton *widget.Button

	path string
}

func createPathSelector(label string, valToSet *string, onChange func(path string), parent fyne.Window) *componentPathSelector {
	cps := &componentPathSelector{}

	cps.selectedPathLabel = widget.NewLabel(fmt.Sprintf("No %s selected", strings.ToLower(label)))

	cps.openPathSelButton = widget.NewButtonWithIcon(
		"Select",
		theme.FolderOpenIcon(),
		func() {
			fileDialog := dialog.NewFolderOpen(
				func(uri fyne.ListableURI, err error) {
					if err != nil {
						win.writeErr("Error opening folder dialog")
						return
					}

					if uri == nil {
						return
					}

					path := uri.Path()
					cps.selectedPathLabel.SetText(filepath.Base(path))

					cps.path = path
					if valToSet != nil {
						*valToSet = path
					}
					if onChange != nil {
						onChange(path)
					}
				},
				parent,
			)

			fileDialog.Show()
		},
	)

	cps.label = createSectionLabel(label)

	cps.Container = container.NewVBox(
		cps.label.Container,
		cps.openPathSelButton,
		cps.selectedPathLabel,
	)

	return cps
}

func (this *componentPathSelector) getPath() string {
	return this.path
}

func (this *componentPathSelector) setPath(path string) {
	this.path = path
	this.selectedPathLabel.SetText(path)
}

func (this *componentPathSelector) disable() {
	this.openPathSelButton.Disable()
}

func (this *componentPathSelector) enable() {
	this.openPathSelButton.Enable()
}
