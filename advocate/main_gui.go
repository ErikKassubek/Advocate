// Copyright (c) 2024 Erik Kassubek
//
// File: main.go
// Brief: Main file and starting point for the toolchain
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build gui

package main

import (
	"advocate/gui"
	"advocate/tool"
	"advocate/utils/flags"
	"advocate/utils/log"
)

var (
	help bool
)

// Main function
func main() {

	cont := tool.CommandLine()
	if !cont {
		return
	}

	tool.Initialize()

	flags.ContainsGui = true

	if flags.Mode == flags.ModeGui {
		flags.Mode = flags.ModeRecord
		log.Info("Start Gui.")
		gui.Run()
	} else {
		err := tool.Run()
		if err != nil {
			log.Error(err)
		}
	}
}
