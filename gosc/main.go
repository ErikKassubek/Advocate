// Copyright (c) 2024
//
// File: main.go
// Brief: Main file and starting point for the toolchain
//
//
// License: BSD-3-Clause

package main

import (
	"gosc/tool"
	"gosc/utils/flags"
	"gosc/utils/log"
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

	err := tool.Run()
	if err != nil {
		log.Error(err)
	}
}
