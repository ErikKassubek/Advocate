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
	"advocate/advoc"
	"advocate/utils/flags"
	"advocate/utils/log"
)

var (
	help bool
)

// Main function
func main() {

	cont := advoc.CommandLine()
	if !cont {
		return
	}

	advoc.Initialize()

	flags.ContainsGui = false

	if flags.Mode == "gui" {
		log.Error("Gui mode is only available if advocate was build with -tags gui (go build -tags gui).")
		return
	}

	err := advoc.Run()
	if err != nil {
		log.Error(err)
	}
}
