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
	"advocate/utils/flags"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentButton struct {
	*fyne.Container

	runButton *widget.Button
}

func createButton(label string, f func()) *componentButton {
	rb := &componentButton{}

	rb.runButton = widget.NewButton(label, f)

	rb.Container = container.NewVBox(
		rb.runButton,
	)

	return rb
}

func createRunButton() *componentButton {
	return createButton("Run", win.start)
}

func createCancelButton() *componentButton {
	return createButton("Cancel", func() { win.WriteGui("Cancel..."); win.worker.cancel() })
}

func createTraceButton() *componentButton {
	return createButton("Trace Viewer", win.openTraceViewer)
}

func (this *componentButton) disable() {
	this.runButton.Disable()
}

func (this *componentButton) enable() {
	this.runButton.Enable()
}

func validInput() bool {
	if flags.ProgPath == "" {
		win.writeErr("Program path not set")
		win.writeErr("Abort...")
		return false
	}

	return true
}
