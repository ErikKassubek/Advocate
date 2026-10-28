// Copyright (c) 2026 Erik Kassubek
//
// File: componentsMainTestSelector.go
// Brief: Create main/test selector
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"advocate/utils/flags"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	unitTest = "Unit Test"
	mainFunc = "Main"
	allTests = "All Tests"
	oneTest  = "One Test"
)

type componentMainTestSelect struct {
	*fyne.Container

	label *componentSectionLabel

	mainTestSel *widget.Select
	allOneSel   *widget.Select
	testNameSel *widget.Select

	showAllOne   bool
	showTestName bool

	replay bool

	testNames []string
}

func creatMainTestSelector() *componentMainTestSelect {
	csmt := &componentMainTestSelect{}

	csmt.mainTestSel = widget.NewSelect(
		[]string{
			mainFunc,
			unitTest,
		},
		func(value string) {
			if value == mainFunc {
				flags.ModeMain = true
				csmt.showAllOne = false
				csmt.showTestName = false
			} else {
				flags.ModeMain = false
				csmt.showAllOne = true
				csmt.showTestName = (csmt.allOneSel.Selected == oneTest)
				flags.ExecName = csmt.testNameSel.Selected

			}

			csmt.creatMainTestSelectorContainer()
		},
	)

	csmt.allOneSel = widget.NewSelect(
		[]string{
			allTests,
			oneTest,
		},
		func(s string) {
			if s == allTests {
				flags.ExecName = ""
				csmt.showTestName = false
			} else {

				csmt.showTestName = true
			}

			csmt.creatMainTestSelectorContainer()
		},
	)

	csmt.testNameSel = widget.NewSelect(
		[]string{},
		func(s string) {
			flags.ExecName = s
		},
	)

	csmt.Container = container.NewVBox()

	csmt.creatMainTestSelectorContainer()

	csmt.mainTestSel.SetSelected(mainFunc)
	csmt.allOneSel.SetSelected(allTests)

	if len(csmt.testNameSel.Options) == 0 {
		csmt.testNameSel.PlaceHolder = "No tests found"
	}

	return csmt
}

func (this *componentMainTestSelect) creatMainTestSelectorContainer() {
	this.label = createSectionLabel("Main/Test")
	objects := []fyne.CanvasObject{
		this.label.Container,
		this.mainTestSel,
	}

	if this.showAllOne {
		objects = append(objects, this.allOneSel)
	}

	if this.showTestName {
		objects = append(objects, this.testNameSel)
	}

	this.Container.Objects = objects
	this.Container.Refresh()
}

func (this *componentMainTestSelect) setTestNames(names *[]string) {
	this.testNames = *names
	this.testNameSel.Options = this.testNames

	if len(*names) != 0 {
		this.testNameSel.SetSelected(this.testNames[0])
	} else {
		this.testNameSel.ClearSelected()
	}
}

func (this *componentMainTestSelect) isReplay(r bool) {
	this.replay = r
	if r {
		this.allOneSel.SetSelected(oneTest)
		this.allOneSel.Hide()
	} else {
		this.allOneSel.SetSelected(allTests)
		this.allOneSel.Show()
	}
}

func (this *componentMainTestSelect) disable() {
	this.mainTestSel.Disable()
	this.allOneSel.Disable()
	this.testNameSel.Disable()
}

func (this *componentMainTestSelect) enable() {
	this.mainTestSel.Enable()
	this.allOneSel.Enable()
	this.testNameSel.Enable()
}
