// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: runCommand.go
// Brief: Functions to run commands
//
// Author: Erik Kassubek, Mario Occhinegro
//
// License: BSD-3-Clause

package command

import (
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

const (
	NoTimeout = -1
)

var count = 0

// RunCommand runs a command line (shell) commands
//
// Parameter:
//   - osOut *os.File: file/output to write to not being what os.Stdout points to
//   - osErr *os.File: file/output to write to not being what os.Stdout points to
//   - timeout int: timeout in seconds, -1 for no timeout
//   - name string: main command
//   - args ...string: command line parameters
//
// Returns:
//   - error
func RunCommand(osOut, osErr *os.File, timeout int, name string, args ...string) error {
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	id := control.AddRunningCom(cancel)
	defer control.RemoveRunningCom(id)

	cmd := exec.CommandContext(ctx, name, args...)

	log.Command(cmd.String())

	if flags.Output {
		if osOut != nil {
			multiOut := io.MultiWriter(os.Stdout, osOut)
			cmd.Stdout = multiOut
		}
		if osErr != nil {
			multiErr := io.MultiWriter(os.Stderr, osErr)
			cmd.Stderr = multiErr
		}
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	count++

	return cmd.Run()
}

func RunGoModTidy() {
	log.Info("Run go mod tidy")

	err := os.Setenv("GOROOT", paths.GoPatch)
	if err == nil {
		defer os.Unsetenv("GOROOT")
	}
	RunCommand(nil, nil, NoTimeout, "go", "mod", "tidy")
}

// func runCommandWithOutput(name, outputFile string, args ...string) (string, error) {
// 	cmd := exec.Command(name, args...)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return "", err
// 	}

// 	// Write output to the specified file
// 	return string(output), os.WriteFile(outputFile, output, 0644)
// }

// // runCommandWithTee runs a command and writes output to a file
// func runCommandWithTee(name, outputFile string, args ...string) error {
// 	cmd := exec.Command(name, args...)
// 	outfile, err := os.Create(outputFile)
// 	if err != nil {
// 		return err
// 	}
// 	defer outfile.Close()
// 	cmd.Stdout = outfile
// 	cmd.Stderr = outfile
// 	return cmd.Run()
// }
