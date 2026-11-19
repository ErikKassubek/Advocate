// Copyright (c) 2024 Erik Kassubek
//
// File: main.go
// Brief: Main file and starting point for the toolchain
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

//go:build !gui

package main

import (
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
		log.Error("Failed CL Check. Cancel...")
		return
	}

	tool.Initialize()

	flags.ContainsGui = false

	if flags.Mode == "gui" {
		log.Error("Gui mode is only available if advocate was build with -tags gui (go build -tags gui).")
		return
	}

	err := tool.Run()
	if err != nil {
		log.Error(err)
	}
}
