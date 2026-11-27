// Copyright (c) 2026 Erik Kassubek
//
// File: componentsSetting.go
// Brief: Settings
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package gui

import (
	"advocate/fuzzing/f_base"
	"advocate/utils/flags"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type settingSec string

const (
	mode      settingSec = "Modes and Scenarios"
	limit     settingSec = "Timeout and Limits"
	statistic settingSec = "Statistics"
	recording settingSec = "Recording"
	behaviour settingSec = "Behaviour"
	logging   settingSec = "Logging"
)

type componentSetting struct {
	Container  *fyne.Container
	record     *settingsRecord
	components *settingComponents

	boxes map[settingSec]*fyne.Container
}

func createSettings() *componentSetting {
	comp := createSettingComponents()

	boxes := map[settingSec]*fyne.Container{
		mode:      container.NewVBox(widget.NewLabel("Column 1 - Row 1")),
		limit:     container.NewVBox(widget.NewLabel("Column 2 - Row 1")),
		statistic: container.NewVBox(widget.NewLabel("Column 3 - Row 1")),
		recording: container.NewVBox(widget.NewLabel("Column 1 - Row 2")),
		behaviour: container.NewVBox(widget.NewLabel("Column 2 - Row 2")),
		logging:   container.NewVBox(widget.NewLabel("Column 3 - Row 2")),
	}

	cs := &componentSetting{
		Container: container.NewGridWithColumns(3,
			boxes[mode],
			boxes[limit],
			boxes[logging],
			boxes[behaviour],
			boxes[statistic],
			boxes[recording],
		),
		record:     createSettingsRecord(comp),
		components: comp,
		boxes:      boxes,
	}

	return cs
}

func (this *window) setSettings(name settingSec, content []fyne.CanvasObject) {
	box, ok := this.settings.boxes[name]
	if !ok {
		return
	}

	label := canvas.NewText(string(name), color.RGBA{255, 0, 0, 255})

	content = append([]fyne.CanvasObject{widget.NewSeparator(), label}, content...)
	box.Objects = content
	box.Refresh()
}

func (this *componentSetting) disable() {
	this.components.disable()
}

func (this *componentSetting) enable() {
	this.components.enable()
}

// ------------------------------------------------------------------------------------
// Setting components
// ------------------------------------------------------------------------------------

type settingComponents struct {
	timeout   *textInput
	toFuzzing *textInput

	maxFuzzingRun     *textInput
	maxNumberElements *textInput

	measureTime         *checkInput
	createStatistics    *checkInput
	checkForNotExecuted *checkInput

	ignoreCriticalSections *checkInput
	ignoreAtomics          *checkInput
	onlyAPanicAndLeak      *checkInput

	noRewrite   *checkInput
	deleteTrace *checkInput
	cont        *checkInput

	runStatic *checkInput

	recordPlus *checkInput

	noWarning  *checkInput
	verbose    *checkInput
	noProgress *checkInput
	output     *checkInput

	alwaysPanic *checkInput
	noMemorySup *checkInput

	fuzzingMode *selectInput

	cancelTestIfBugFound *checkInput

	scen []*checkInput
}

const (
	inverted         = true
	direct           = false
	canBeDisabled    = true
	canNotBeDisabled = false
)

func createSettingComponents() *settingComponents {
	to := settingComponents{
		timeout:                createInputNumeric("Timeout Run [s]: ", &flags.Timeout, canBeDisabled),
		toFuzzing:              createInputNumeric("Timeout Fuzzing [s]: ", &flags.TimeoutFuzzing, canBeDisabled),
		maxFuzzingRun:          createInputNumeric("Max. Fuzzing Runs: ", &flags.MaxFuzzingRun, canBeDisabled),
		maxNumberElements:      createInputNumeric("Max. Number Elements: ", &flags.MaxNumberElements, canBeDisabled),
		measureTime:            createInputCheck("Measure Time: ", &flags.MeasureTime, direct),
		createStatistics:       createInputCheck("Create Statistics: ", &flags.CreateStatistics, direct),
		checkForNotExecuted:    createInputCheck("Check for not Executed Ops: ", &flags.NotExecuted, direct),
		runStatic:              createInputCheck("Run Static Blocking Analysis: ", &flags.Static, direct),
		ignoreCriticalSections: createInputCheck("Ignore Critical Sections: ", &flags.IgnoreCriticalSection, direct),
		ignoreAtomics:          createInputCheck("Ignore Atomics: ", &flags.IgnoreAtomics, direct),
		onlyAPanicAndLeak:      createInputCheck("Disable Prediction: ", &flags.OnlyAPanicAndLeak, direct),
		noRewrite:              createInputCheck("Skip Rewrite: ", &flags.NoRewrite, direct),
		deleteTrace:            createInputCheck("Delete Trace:", &flags.DeleteTraces, direct),
		cont:                   createInputCheck("Continue: ", &flags.Continue, direct),
		noWarning:              createInputCheck("Disable Warning Massages: ", &flags.NoWarning, direct),
		verbose:                createInputCheck("Verbose Messages: ", &flags.Verbose, direct),
		noProgress:             createInputCheck("Disable Progress Massages: ", &flags.NoProgress, direct),
		output:                 createInputCheck("Show Program/Test Output: ", &flags.Output, direct),
		alwaysPanic:            createInputCheck("Always Panic: ", &flags.AlwaysPanic, direct),
		noMemorySup:            createInputCheck("Disable Memory Supervisor", &flags.NoMemorySupervisor, direct),
		fuzzingMode:            createInputSelect("Fuzzing Mode: ", &flags.FuzzingMode, f_base.Modes, f_base.Full, canNotBeDisabled),
		cancelTestIfBugFound:   createInputCheck("Cancel Fuzzing If Bug Found: ", &flags.CancelTestIfBugFound, direct),
		recordPlus:             createInputCheck("Record Function and Controll Flow", &flags.RecordPlus, direct),
		scen:                   createScen(),
	}

	// TODO: Only for testing select. Remove
	to.runStatic.check.SetChecked(true)

	return &to
}

func createScen() []*checkInput {
	res := make([]*checkInput, 8)

	checked := true
	res[0] = createInputCheckFunc("Scenario - Send on Closed Channel: ", scenChange(), "s", checked)
	res[1] = createInputCheckFunc("Scenario - Close of Closed Channel: ", scenChange(), "c", checked)
	res[2] = createInputCheckFunc("Scenario - Negative WaitGroup Counter: ", scenChange(), "w", checked)
	res[3] = createInputCheckFunc("Scenario - Unlock before Lock: ", scenChange(), "u", checked)
	res[4] = createInputCheckFunc("Scenario - Resource Deadlock: ", scenChange(), "d", checked)
	res[5] = createInputCheckFunc("Scenario - Mixed Deadlock: ", scenChange(), "m", checked)
	res[6] = createInputCheckFunc("Scenario - Leak: ", scenChange(), "l", checked)
	res[7] = createInputCheckFunc("Scenario - Concurrent Recveive on same Channel: ", scenChange(), "r", checked)

	return res
}

func scenChange() func(b bool, letter string) {
	return func(b bool, letter string) {
		if b {
			if flags.Scenarios == "-" {
				flags.Scenarios = ""
			}
			if !strings.Contains(flags.Scenarios, letter) {
				flags.Scenarios += letter
			}
		} else {
			flags.Scenarios = strings.ReplaceAll(flags.Scenarios, letter, "")
			if flags.Scenarios == "" {
				flags.Scenarios = "-"
			}
		}

		err := flags.ParseAnalysisCases()
		if err != nil {
			panic(err)
		}
	}
}

func (this *settingComponents) disable() {
	this.timeout.disable()
	this.toFuzzing.disable()

	this.maxFuzzingRun.disable()
	this.maxNumberElements.disable()

	this.measureTime.disable()
	this.createStatistics.disable()
	this.checkForNotExecuted.disable()

	this.runStatic.disable()

	this.ignoreCriticalSections.disable()
	this.ignoreAtomics.disable()
	this.onlyAPanicAndLeak.disable()

	this.noRewrite.disable()
	this.deleteTrace.disable()
	this.cont.disable()

	this.noWarning.disable()
	this.verbose.disable()
	this.noProgress.disable()
	this.output.disable()

	this.alwaysPanic.disable()
	this.noMemorySup.disable()

	this.fuzzingMode.disable()

	this.cancelTestIfBugFound.disable()

	for _, s := range this.scen {
		s.disable()
	}

	win.modeSelect.disable()
	win.projSelector.disable()
}

func (this *settingComponents) enable() {
	this.timeout.enable()
	this.toFuzzing.enable()

	this.maxFuzzingRun.enable()
	this.maxNumberElements.enable()

	this.measureTime.enable()
	this.createStatistics.enable()
	this.checkForNotExecuted.enable()

	this.runStatic.enable()

	this.ignoreCriticalSections.enable()
	this.ignoreAtomics.enable()
	this.onlyAPanicAndLeak.enable()

	this.noRewrite.enable()
	this.deleteTrace.enable()
	this.cont.enable()

	this.noWarning.enable()
	this.verbose.enable()
	this.noProgress.enable()
	this.output.enable()

	this.alwaysPanic.enable()
	this.noMemorySup.enable()

	this.fuzzingMode.enable()

	this.cancelTestIfBugFound.enable()

	for _, s := range this.scen {
		s.enable()
	}
}
