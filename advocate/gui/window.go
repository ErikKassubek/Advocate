// Copyright (c) 2026 Erik Kassubek
//
// File: window.go
// Brief: Create main window
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"advocate/utils/flags"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	red    = color.RGBA{255, 99, 99, 255}
	green  = color.RGBA{80, 220, 120, 255}
	blue   = color.RGBA{100, 170, 255, 255}
	purple = color.RGBA{200, 140, 255, 255}
	yellow = color.RGBA{255, 220, 90, 255}
	pink   = color.RGBA{255, 120, 200, 255}
	gray   = color.RGBA{200, 200, 200, 255}
)

var windowSize = fyne.NewSize(1920, 1080)

type window struct {
	app    fyne.App
	window fyne.Window

	left  *fyne.Container
	right *fyne.Container

	modeSelect   *componentModeSelect
	projSelector *componentPathSelector
	runButton    *componentButton
	cancelButton *componentButton
	traceButton  *componentButton
	output       *componentOutput
	progressBar  *componentProgress
	settings     *componentSetting

	worker *worker
}

func (this *window) create() {
	this.app = app.New()
	this.window = this.app.NewWindow("Advocate")

	// self.w.Resize(fyne.NewSize(800, 500))
	this.window.Resize(windowSize)
	this.window.CenterOnScreen()

	this.handleClose()

	this.createComponents()
}

func (this *window) build() {
	this.left = container.NewBorder(
		container.NewVBox(
			this.modeSelect.Container,
			widget.NewSeparator(),
			this.projSelector.Container,
		),
		container.NewVBox(
			this.runButton.Container,
			this.cancelButton.Container,
			this.traceButton.Container,
		),
		nil,
		nil,
		container.NewScroll(this.settings.Container),
	)

	this.right = container.NewBorder(
		nil,
		this.progressBar.Container,
		nil,
		nil,
		container.NewScroll(this.output.Container),
	)

	content := container.NewHSplit(this.left, this.right)
	content.SetOffset(0.33)

	this.window.SetContent(content)
}

func (this *window) createComponents() {
	this.output = createOutput()

	this.projSelector = createPathSelector("Project", &flags.ProgPath, getAllTestNames, win.window)
	this.runButton = createRunButton()
	this.cancelButton = createCancelButton()
	this.traceButton = createTraceButton()
	this.progressBar = createProgressBar()
	this.settings = createSettings()

	this.modeSelect = createModeSelect() // must be created last

	this.cancelButton.disable()

	this.projSelector.setPath(flags.ProgPath)
}

func (this *window) showAndRun() {
	// without additional show, the window content is shifted up
	this.window.Show()
	this.window.ShowAndRun()
}

func (this *window) handleClose() {
	this.window.SetCloseIntercept(func() {
		this.WriteGui("Application shutting down...")
		this.window.Close()
		os.Exit(0)
	})
}
