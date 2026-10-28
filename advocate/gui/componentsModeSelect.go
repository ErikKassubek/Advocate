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
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentModeSelect struct {
	*fyne.Container

	label *componentSectionLabel

	modeSelectWidget *widget.Select
}

const (
	record   = "Record"
	replay   = "Replay"
	analysis = "Analysis"
	fuzzing  = "Fuzzing"
	static   = "Static"
)

func createModeSelect() *componentModeSelect {
	cms := &componentModeSelect{}

	cms.modeSelectWidget = widget.NewSelect(
		[]string{
			record,
			replay,
			analysis,
			fuzzing,
			static,
		},
		func(value string) {
			flags.Mode = strings.ToLower(value)

			if value == replay {
				win.settings.components.mainTestSelect.isReplay(true)
			} else {
				win.settings.components.mainTestSelect.isReplay(false)
			}

			switch value {
			case record:
				win.setRecord()
			case analysis:
				win.setAnalysis()
			case replay:
				win.setReplay()
			case fuzzing:
				win.setFuzzing()
			case static:
				win.setStatic()
			}

		},
	)

	cms.label = createSectionLabel("Mode")

	cms.Container = container.NewVBox(
		cms.label.Container,
		cms.modeSelectWidget,
	)

	cms.modeSelectWidget.SetSelected(static)

	return cms
}

func (this *componentModeSelect) disable() {
	this.modeSelectWidget.Disable()
}

func (this *componentModeSelect) enable() {
	this.modeSelectWidget.Enable()
}

func getAllTestNames(path string) {
	if path == "" {
		return
	}

	var testNames []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}

			if strings.HasPrefix(fn.Name.Name, "Test") {
				testNames = append(testNames, fn.Name.Name)
			}
		}

		return nil
	})

	if err != nil {
		win.writeErr(err.Error())
	}

	sort.Strings(testNames)

	win.settings.components.mainTestSelect.setTestNames(&testNames)
}
