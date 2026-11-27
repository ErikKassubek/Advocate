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
	"advocate/utils/flags"
	"advocate/utils/types"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type settings struct {
	window fyne.Window

	closeButton *componentButton
}

func (this *window) openSettings() {
	se := settings{}
	se.window = this.app.NewWindow("Settings " + types.CapitalizeFirst(flags.Mode))

	se.closeButton = createButton("Close", se.window.Close)

	content := container.NewBorder(
		nil, se.closeButton.Container, nil, nil,
		this.settings.Container,
	)

	se.window.SetContent(content)
	se.window.Resize(fyne.NewSize(1728, 972))

	se.window.Show()
}
