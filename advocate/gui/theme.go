// Copyright (c) 2026 Erik Kassubek
//
// File: theme.go
// Brief: Theme
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type guiTheme struct {
	fyne.Theme
}

func (t guiTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameSeparator {
		return color.Transparent
	}

	return t.Theme.Color(name, variant)
}
