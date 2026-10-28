// Copyright (c) 2026 Erik Kassubek
//
// File: modeSelect.go
// Brief: Create mode selector
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"advocate/utils/log"
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// internal storage for log lines
type logLine struct {
	text string
	lv   log.InfoLevel
}

type componentOutput struct {
	*fyne.Container

	lines         []logLine
	outputList    *widget.List
	outputChannel chan log.GuiInfo
}

func createOutput() *componentOutput {
	co := &componentOutput{}

	// color mapping
	getColor := func(mode log.InfoLevel) color.Color {
		switch mode {
		case log.ImportantLv, log.DebugLv:
			return yellow
		case log.Debug2Lv:
			return pink
		case log.ResultLv:
			return green
		case log.ProgressLv:
			return blue
		case log.TimeoutLv:
			return purple
		case log.ErrorLv:
			return red
		case log.OutputLv:
			return gray
		case log.GuiLv:
			return pink
		default:
			return color.White
		}
	}

	// create list
	co.outputList = widget.NewList(
		func() int { return len(co.lines) },
		func() fyne.CanvasObject {
			return container.NewPadded(canvas.NewText("", color.White))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			txt := o.(*fyne.Container).Objects[0].(*canvas.Text)

			txt.Text = co.lines[i].text
			txt.Color = getColor(co.lines[i].lv)
			txt.TextSize = 14
		},
	)

	// self.output.SetMinSize(fyne.NewSize(700, 400))

	co.outputChannel = log.GetGuiChan()

	co.Container = container.NewBorder(nil, nil, nil, nil,
		container.NewVScroll(co.outputList),
	)

	co.write(log.GuiLv, "Application started")

	go func() {
		for c := range co.outputChannel {
			fyne.Do(func() {
				co.write(c.Lv, c.Msg)
			})
		}
	}()

	return co
}

func (this *componentOutput) write(lv log.InfoLevel, text string) {
	timestamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s] %s", timestamp, text)

	this.lines = append(this.lines, logLine{
		text: line,
		lv:   lv,
	})

	this.outputList.Refresh()

	this.outputList.ScrollToBottom()
}

func (this *window) write(lv log.InfoLevel, msg ...any) {
	this.output.write(lv, fmt.Sprint(msg))
}

func (this *window) writef(lv log.InfoLevel, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	this.output.write(lv, msg)
}

func (this *window) WriteGui(msg ...any) {
	this.output.write(log.GuiLv, fmt.Sprint(msg))
}

func (this *window) writeGuif(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	this.output.write(log.GuiLv, msg)
}

func (this *window) writeErr(msg ...any) {
	this.output.write(log.ErrorLv, fmt.Sprint(msg))
}

func (this *window) writeErrf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	this.output.write(log.ErrorLv, msg)
}
