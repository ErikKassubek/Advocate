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
	"advocate/tool/toolchain"
	"advocate/utils/flags"
	"advocate/utils/types"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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

	modeSelect     *componentModeSelect
	projSelector   *componentPathSelector
	mainTestSelect *componentMainTestSelect
	traceSelect    *componentPathSelector

	runButton        *componentButton
	cancelButton     *componentButton
	traceButton      *componentButton
	output           *componentOutput
	progressBarOne   *componentProgress
	progressBarTotal *componentProgress
	settingsButton   *componentButton
	settings         *componentSetting

	worker *worker
}

func (this *window) create() {
	this.app = app.New()
	this.window = this.app.NewWindow("Advocate")

	// self.w.Resize(fyne.NewSize(800, 500))
	this.window.Resize(windowSize)
	this.window.CenterOnScreen()

	this.app.Settings().SetTheme(guiTheme{
		Theme: theme.DefaultTheme(),
	})

	this.handleClose()

	this.createComponents()

	this.startProgressComm()

}

func (this *window) build() {
	this.left = container.NewBorder(
		container.NewVBox(
			this.modeSelect.Container,
			widget.NewSeparator(),
			this.projSelector.Container,
			widget.NewSeparator(),
			this.mainTestSelect.Container,
			widget.NewSeparator(),
			this.traceSelect.Container,
			widget.NewSeparator(),
			widget.NewSeparator(),
		),
		container.NewVBox(
			this.settingsButton.Container,
			this.runButton.Container,
			this.cancelButton.Container,
			this.traceButton.Container,
		),
		nil,
		nil,
	)

	this.right = container.NewBorder(
		nil,
		container.NewVBox(
			this.progressBarOne.Container,
			this.progressBarTotal.Container,
		),
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
	this.mainTestSelect = creatMainTestSelector()
	this.traceSelect = createPathSelector("Trace", &flags.TracePath, getAllTestNames, win.window)

	this.runButton = createRunButton()
	this.cancelButton = createCancelButton()
	this.traceButton = createTraceButton()

	this.progressBarOne = createProgressBar()
	this.progressBarTotal = createProgressBar()

	this.settingsButton = createSettingsButton()
	this.settings = createSettings()

	this.modeSelect = createModeSelect() // must be created last

	this.cancelButton.disable()

	this.traceSelect.disable()

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

func (this *window) startProgressComm() {
	go func() {
		toolchain.GuiCom = make(chan types.Pair[string, float64], 100)
		for val := range toolchain.GuiCom {
			switch val.X {
			case "progressTotal":
				this.progressBarTotal.setValue(val.Y)
			case "progressOne":
				this.progressBarOne.setValue(val.Y)
			}
		}
	}()
}
