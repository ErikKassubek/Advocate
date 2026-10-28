// Copyright (c) 2026 Erik Kassubek
//
// File: componentInput.go
// Brief: input components
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type componentSectionLabel struct {
	*fyne.Container

	label *canvas.Text
}

func createSectionLabel(text string) *componentSectionLabel {
	label := canvas.NewText(
		text,
		theme.Color(theme.ColorNameError),
	)

	return &componentSectionLabel{
		Container: container.NewHBox(label),
		label:     label,
	}
}

func createBoxedLabel(text string, noBox, request, commit bool) fyne.CanvasObject {

	label := widget.NewLabel(text)

	if noBox {
		return label
	}

	return container.NewStack(
		createBorder(request, commit),
		label,
	)
}

func createBorder(request, commit bool) fyne.CanvasObject {
	top := canvas.NewLine(color.White)
	bottom := canvas.NewLine(color.White)
	left := canvas.NewLine(color.White)
	right := canvas.NewLine(color.White)

	objects := []fyne.CanvasObject{}

	if !commit {
		objects = append(objects, top)
	}

	if !request {
		objects = append(objects, bottom)
	}

	objects = append(objects, left, right)

	border := container.NewWithoutLayout(objects...)

	border.Layout = &borderLayout{
		removeTop:    commit,
		removeBottom: request,
	}

	return border
}

type borderLayout struct {
	removeTop    bool
	removeBottom bool
}

func (l *borderLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	i := 0

	if !l.removeTop {
		objects[i].Move(fyne.NewPos(0, 0))
		objects[i].Resize(fyne.NewSize(size.Width, 1))
		i++
	}

	if !l.removeBottom {
		objects[i].Move(fyne.NewPos(0, size.Height-1))
		objects[i].Resize(fyne.NewSize(size.Width, 1))
		i++
	}

	// Left
	objects[i].Move(fyne.NewPos(0, 0))
	objects[i].Resize(fyne.NewSize(1, size.Height))
	i++

	// Right
	objects[i].Move(fyne.NewPos(size.Width-1, 0))
	objects[i].Resize(fyne.NewSize(1, size.Height))
}

func (l *borderLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}
