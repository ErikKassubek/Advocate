// Copyright (c) 2026 Erik Kassubek
//
// File: traceViewer.go
// Brief: Create trace viewer
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"advocate/advoc/toolchain"
	"advocate/analysis/a_base"
	"advocate/trace"
	"advocate/utils/io"
	"advocate/utils/types"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type traceViewer struct {
	window fyne.Window

	pathSelector *componentPathSelector
	closeButton  *componentButton
	traceViewer  *componentTraceViewer

	path  string
	trace *trace.Trace
}

func (this *window) openTraceViewer() {
	tv := traceViewer{}

	tv.window = this.app.NewWindow("Trace")

	tv.pathSelector = createPathSelector("Trace", &tv.path, tv.updateTrace, tv.window)

	tv.closeButton = createButton("Close", tv.window.Close)

	tv.traceViewer = createTraceViewer(0, 0)

	content := container.NewBorder(
		tv.pathSelector.Container,
		tv.closeButton.Container,
		nil,
		nil,
		container.NewStack(tv.traceViewer.Scroll),
	)

	tv.window.SetContent(content)

	tv.window.Resize(fyne.NewSize(1728, 972))

	if toolchain.LastMovedTraceDest != "" {
		tv.pathSelector.setPath(toolchain.LastMovedTraceDest)
		tv.updateTrace(toolchain.LastMovedTraceDest)
	}

	tv.window.Show()
}

func (this *traceViewer) updateTrace(dir string) {
	routs, _, err := this.readTrace(dir)
	if err != nil {
		win.writeErr("Could not read trace: ", err)
		return
	}

	this.trace = &a_base.MainTrace

	elems := this.trace.AsRequestCommit()
	this.trace.NormalizeRequestCommit()
	this.traceViewer = createTraceViewer(routs, elems)

	content := container.NewBorder(
		this.pathSelector.Container,
		this.closeButton.Container,
		nil,
		nil,
		container.NewStack(this.traceViewer.Scroll),
	)

	this.window.SetContent(content)

	reqCom := make(map[int][]types.Pair[int, int])
	for id := 1; id < this.trace.GetNoRoutines()+1; id++ {
		reqCom[id] = make([]types.Pair[int, int], 0)
	}
	inOp := make(map[int]bool)

	traceIter := this.trace.AsIterator()
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		rout := elem.Routine().ID()
		row := elem.T(trace.Commit) + 1

		elem_name := elem.StringGui()

		removeTop := false
		request := false

		if elem.Request() {
			// elem_name += "?"
			request = true
			reqCom[rout] = append(reqCom[rout], types.NewPair(row, 0))
			inOp[rout] = true
		} else if elem.CanBeRequest() {
			removeTop = true
			elem_name = ""
			if inOp[rout] {
				inOp[rout] = false
				l := len(reqCom[rout]) - 1
				elem := reqCom[rout][l]
				elem.Y = row
				reqCom[rout][l] = elem
			}
		}

		this.AddEntry(rout, row, elem_name, request, removeTop)
	}

	// add sides between request and commit
	for rout, val := range reqCom {
		for _, op := range val {
			if op.Y == 0 {
				continue
			}
			for i := op.X + 1; i < op.Y; i++ {
				this.traceViewer.AddCell(rout, i, "", false, true, true)
			}
		}

	}

	this.rebuild()
}

func (this *traceViewer) AddEntry(rout, row int, text string, request, commit bool) {
	noBox := false
	this.traceViewer.AddCell(rout, row, text, noBox, request, commit)
}

func (this *traceViewer) rebuild() {
	this.traceViewer.rebuild()
}

func (this *traceViewer) readTrace(dir string) (cols int, rows int, err error) {
	return io.CreateTraceFromFiles(dir)
}
