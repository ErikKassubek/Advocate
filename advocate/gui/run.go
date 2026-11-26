// Copyright (c) 2026 Erik Kassubek
//
// File: run.go
// Brief: Run advocate
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"advocate/tool"
	"advocate/tool/toolchain"
	"context"

	"fyne.io/fyne/v2"
)

type worker struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newWorker() *worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &worker{ctx, cancel}

}

func (this *worker) IsCanceled() bool {
	return this.ctx.Err() != nil
}

func (this *window) startRunMode() {
	this.settings.disable()
	win.modeSelect.disable()
	win.projSelector.disable()
	win.runButton.disable()
	win.cancelButton.enable()
	win.settingsButton.disable()
}

func (this *window) endRunMode() {
	this.settings.enable()
	win.modeSelect.enable()
	win.projSelector.enable()
	win.runButton.enable()
	win.cancelButton.disable()
	win.settingsButton.enable()
}

func (this *window) start() {
	this.worker = newWorker()

	go func() {
		if !validInput() {
			return
		}

		fyne.Do(func() {
			win.startRunMode()
			win.WriteGui("Start Run")
		})

		err := tool.Run()
		if err != nil {
			if !this.worker.IsCanceled() {
				fyne.Do(func() { win.writeErr(err.Error()) })
			}
		}

		fyne.Do(func() {
			win.endRunMode()
			if !this.worker.IsCanceled() {
				win.WriteGui("Finish Run")
			}
		})
	}()
}

func (this *window) cancel() {
	toolchain.Cancel()
	win.worker.cancel()
	win.WriteGui("Cancel...")
	win.endRunMode()
}
