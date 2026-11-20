package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type outputTab struct {
	lines      []string
	outputList *widget.List
}

type componentResultViewer struct {
	*container.Scroll

	tabs    *container.AppTabs
	outputs []*outputTab
}

func createResultsViewer() *componentResultViewer {
	cr := &componentResultViewer{}

	cr.tabs = container.NewAppTabs()

	cr.Scroll = container.NewScroll(cr.tabs)
	cr.Scroll.SetMinSize(fyne.NewSize(0, 0))

	return cr
}

func (cr *componentResultViewer) AddTab(name string) int {
	tab := &outputTab{
		lines: []string{},
	}

	tab.outputList = widget.NewList(
		func() int {
			return len(tab.lines)
		},
		func() fyne.CanvasObject {
			return container.NewPadded(canvas.NewText("", color.White))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			txt := o.(*fyne.Container).Objects[0].(*canvas.Text)
			txt.Text = tab.lines[i]
			txt.TextSize = 16
			txt.Refresh()
		},
	)

	scroll := container.NewVScroll(tab.outputList)

	cr.tabs.Append(container.NewTabItem(name, scroll))
	cr.outputs = append(cr.outputs, tab)

	return len(cr.outputs) - 1
}

func (cr *componentResultViewer) AddLine(tabIndex int, line string) {
	if tabIndex < 0 || tabIndex >= len(cr.outputs) {
		return
	}

	tab := cr.outputs[tabIndex]
	tab.lines = append(tab.lines, line)
	tab.outputList.Refresh()
}
