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
)

type traceConnection struct {
	a string
	b string
}

type componentTraceViewer struct {
	*container.Scroll

	rows, cols int
	cells      map[string]fyne.CanvasObject
	grid       *fyne.Container

	links []traceConnection
}

func createTraceViewer(cols, rows int) *componentTraceViewer {
	grid := container.NewGridWithColumns(cols)

	scroll := container.NewScroll(grid)
	scroll.SetMinSize(fyne.NewSize(0, 0))

	return &componentTraceViewer{
		Scroll: scroll,
		rows:   rows,
		cols:   cols,
		cells:  make(map[string]fyne.CanvasObject),
		grid:   grid,
	}
}

func key(rout, row int) string {
	return fmt.Sprintf("%d:%d", rout, row)
}

// AddCell places content into a specific row/col
func (this *componentTraceViewer) AddCell(rout, row int, content string, noBox, removeTop, request bool) {
	this.cells[key(rout, row)] = this.buildCell(content, noBox, removeTop, request)
}

// TODO: color/fix text
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
