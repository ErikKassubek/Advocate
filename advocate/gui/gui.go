// Copyright (c) 2026 Erik Kassubek
//
// File: gui.go
// Brief: Entry point for graphical interface
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

var win = window{}

func Run() {
	win.create()
	win.build()
	win.showAndRun()
}
