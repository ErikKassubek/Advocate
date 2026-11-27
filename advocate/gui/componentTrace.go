// Copyright (c) 2026 Erik Kassubek
//
// File: componentTrace.go
// Brief: Trace Viewer
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type traceConnection struct {
	a string
	b string
}

type componentTraceViewer struct {
	widget.BaseWidget

	scroll *container.Scroll

	rows, cols int
	cells      map[string]fyne.CanvasObject
	grid       *fyne.Container

	links []traceConnection

	dragging bool
	lastPos  fyne.Position
}

func createTraceViewer(cols, rows int) *componentTraceViewer {
	grid := container.NewGridWithColumns(cols)

	scroll := container.NewScroll(grid)
	scroll.SetMinSize(fyne.NewSize(0, 0))

	v := &componentTraceViewer{
		scroll: scroll,
		rows:   rows,
		cols:   cols,
		cells:  make(map[string]fyne.CanvasObject),
		grid:   grid,
	}

	v.ExtendBaseWidget(v)
	return v
}

func (v *componentTraceViewer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.scroll)
}

func key(rout, row int) string {
	return fmt.Sprintf("%d:%d", rout, row)
}

// AddCell places content into a specific row/col
func (this *componentTraceViewer) AddCell(rout, row int, content string, noBox, removeTop, request bool) {
	this.cells[key(rout, row)] = this.buildCell(content, noBox, removeTop, request)
}

func (this *componentTraceViewer) buildCell(text string, noBox, removeTop, request bool) fyne.CanvasObject {
	return container.NewPadded(createBoxedLabel(text, noBox, removeTop, request))
}

func (this *componentTraceViewer) rebuild() {
	this.grid.Objects = nil

	for r := 0; r < this.rows+1; r++ {
		for c := 1; c < this.cols+1; c++ {
			k := key(c, r)

			if r == 0 { // routine
				this.grid.Add(this.buildCell(fmt.Sprint("Routine ", c), true, false, false))
			} else {
				if obj, ok := this.cells[k]; ok {
					this.grid.Add(obj)
				} else { // empty placeholder so grid stays aligned
					this.grid.Add(this.buildCell("", true, true, true))
				}
			}
		}
	}

	this.grid.Refresh()
}

func (v *componentTraceViewer) MouseDown(ev *desktop.MouseEvent) {
	if ev.Button != desktop.MouseButtonTertiary {
		return
	}

	v.dragging = true
	v.lastPos = ev.Position
}

func (v *componentTraceViewer) MouseUp(ev *desktop.MouseEvent) {
	if ev.Button == desktop.MouseButtonTertiary {
		v.dragging = false
	}
}

func (v *componentTraceViewer) MouseMoved(ev *desktop.MouseEvent) {
	if !v.dragging {
		return
	}

	delta := ev.Position.Subtract(v.lastPos)
	v.lastPos = ev.Position

	v.scroll.ScrollToOffset(fyne.NewPos(
		v.scroll.Offset.X-delta.X,
		v.scroll.Offset.Y-delta.Y,
	))
}

func (v *componentTraceViewer) MouseIn(*desktop.MouseEvent) {}
func (v *componentTraceViewer) MouseOut()                   { v.dragging = false }
