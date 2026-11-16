// Copyright (c) 2026 Erik Kassubek
//
// File: settings.go
// Brief: Settings window
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type settings struct {
	window fyne.Window
}

func (this *window) openSettings() {
	se := settings{}
	se.window = this.app.NewWindow("Settings")

	content := container.NewBorder(
		nil, nil, nil, nil,
		container.NewScroll(this.settings.Container),
	)

	se.window.SetContent(content)
	se.window.Resize(fyne.NewSize(1728, 972))

	se.window.Show()

}
