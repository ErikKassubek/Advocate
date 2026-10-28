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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type componentSetting struct {
	*fyne.Container

	components *settingComponents

	record *settingsRecord
}

func createSettings() *componentSetting {
	comp := createSettingComponents()

	cs := &componentSetting{
		Container: container.NewVBox(),

		record:     createSettingsRecord(comp),
		components: comp,
	}

	return cs
}

func (this *window) setSettings(content []fyne.CanvasObject) {
	this.settings.Container.Objects = content
	this.settings.Refresh()
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
	label          *componentSectionLabel
	mainTestSelect *componentMainTestSelect

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

	noWarning  *checkInput
	noInfo     *checkInput
	noProgress *checkInput
	output     *checkInput

	alwaysPanic *checkInput
	noMemorySup *checkInput

	fuzzingMode *selectInput

	cancelTestIfBugFound *checkInput

	scen *textInput

	tracePath *componentPathSelector
}

const (
	inverted         = true
	direct           = false
	canBeDisabled    = true
	canNotBeDisabled = false
)

func createSettingComponents() *settingComponents {
	to := settingComponents{
		label:                  createSectionLabel("Settings"),
		mainTestSelect:         creatMainTestSelector(),
		timeout:                createInputNumeric("Timeout Replay [s]: ", &flags.Timeout, canBeDisabled),
		toFuzzing:              createInputNumeric("Timeout Fuzzing [s]: ", &flags.TimeoutFuzzing, canBeDisabled),
		maxFuzzingRun:          createInputNumeric("Max. Fuzzing Runs: ", &flags.MaxFuzzingRun, canBeDisabled),
		maxNumberElements:      createInputNumeric("Max. Number Elements: ", &flags.MaxNumberElements, canBeDisabled),
		measureTime:            createInputCheck("Measure Time: ", &flags.MeasureTime, direct),
		createStatistics:       createInputCheck("Create Statistics: ", &flags.CreateStatistics, direct),
		checkForNotExecuted:    createInputCheck("Check for not Executed Ops: ", &flags.NotExecuted, direct),
		ignoreCriticalSections: createInputCheck("Ignore Critical Sections: ", &flags.IgnoreCriticalSection, direct),
		ignoreAtomics:          createInputCheck("Ignore Atomics: ", &flags.IgnoreAtomics, direct),
		onlyAPanicAndLeak:      createInputCheck("Disable Prediction: ", &flags.OnlyAPanicAndLeak, direct),
		noRewrite:              createInputCheck("Skip Rewrite: ", &flags.NoRewrite, direct),
		deleteTrace:            createInputCheck("Delete Trace:", &flags.DeleteTraces, direct),
		cont:                   createInputCheck("Continue: ", &flags.Continue, direct),
		noWarning:              createInputCheck("Disable Warning Massages: ", &flags.NoWarning, direct),
		noInfo:                 createInputCheck("Disable Info Messages: ", &flags.NoInfo, direct),
		noProgress:             createInputCheck("Disable Progress Massages: ", &flags.NoProgress, direct),
		output:                 createInputCheck("Show Program/Test Output: ", &flags.Output, direct),
		alwaysPanic:            createInputCheck("Always Panic: ", &flags.AlwaysPanic, direct),
		noMemorySup:            createInputCheck("Disable Memory Supervisor", &flags.NoMemorySupervisor, direct),
		fuzzingMode:            createInputSelect("Fuzzing Mode: ", &flags.FuzzingMode, f_base.Modes, canNotBeDisabled),
		cancelTestIfBugFound:   createInputCheck("Cancel Fuzzing If Bug Found: ", &flags.CancelTestIfBugFound, direct),
		scen: createInputTextFunc("Scenations (disabled = all): ", "",
			func(e bool, s string) {
				if e {
					flags.Scenarios = s
				} else {
					flags.Scenarios = ""
				}
				flags.ParseAnalysisCases()
			}, true, true),
		tracePath: createPathSelector("Trace", &flags.TracePath, getAllTestNames, win.window),
	}

	return &to
}

func (this *settingComponents) disable() {
	this.mainTestSelect.disable()

	this.timeout.disable()
	this.toFuzzing.disable()

	this.maxFuzzingRun.disable()
	this.maxNumberElements.disable()

	this.measureTime.disable()
	this.createStatistics.disable()
	this.checkForNotExecuted.disable()

	this.ignoreCriticalSections.disable()
	this.ignoreAtomics.disable()
	this.onlyAPanicAndLeak.disable()

	this.noRewrite.disable()
	this.deleteTrace.disable()
	this.cont.disable()

	this.noWarning.disable()
	this.noInfo.disable()
	this.noProgress.disable()
	this.output.disable()

	this.alwaysPanic.disable()
	this.noMemorySup.disable()

	this.fuzzingMode.disable()

	this.cancelTestIfBugFound.disable()

	this.scen.disable()

	this.tracePath.disable()

	win.modeSelect.disable()
	win.projSelector.disable()
}

func (this *settingComponents) enable() {
	this.mainTestSelect.enable()

	this.timeout.enable()
	this.toFuzzing.enable()

	this.maxFuzzingRun.enable()
	this.maxNumberElements.enable()

	this.measureTime.enable()
	this.createStatistics.enable()
	this.checkForNotExecuted.enable()

	this.ignoreCriticalSections.enable()
	this.ignoreAtomics.enable()
	this.onlyAPanicAndLeak.enable()

	this.noRewrite.enable()
	this.deleteTrace.enable()
	this.cont.enable()

	this.noWarning.enable()
	this.noInfo.enable()
	this.noProgress.enable()
	this.output.enable()

	this.alwaysPanic.enable()
	this.noMemorySup.enable()

	this.fuzzingMode.enable()

	this.cancelTestIfBugFound.enable()

	this.scen.enable()

	this.tracePath.enable()
}
