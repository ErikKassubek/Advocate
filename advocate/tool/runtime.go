// Copyright (c) 2026 Erik Kassubek
//
// File: runtime.go
// Brief: Build the runtime
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package tool

import (
	"advocate/utils/command"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/osinfo"
	"advocate/utils/paths"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func buildRuntime() error {
	pathGo := filepath.Join(paths.GoPatch, "bin", osinfo.GoExec)

	if flags.Mode != flags.ModeGui {
		if _, err := os.Stat(pathGo); err == nil {
			var rebuild string

			for rebuild != "Y" {
				log.WriteInput("Runtime already exists. Rebuild? (Y/N): ")
				fmt.Scan(&rebuild)

				rebuild = strings.ToUpper(rebuild)

				switch rebuild {
				case "Y":
				case "N":
					log.Info("Skip Build. Exit...")
					return nil
				default:
					log.Important("Invalid input.")
				}
			}
		}
	}

	log.Info("Building Runtime...")

	pathSrc := filepath.Join(paths.GoPatch, "src")
	if _, err := os.Stat(pathSrc); err != nil && os.IsNotExist(err) {
		log.Errorf("Could not find goPatch at: %s", pathSrc)
		return err
	}

	makeName := "./make.bash"
	if osinfo.GetOs() == osinfo.Windows {
		makeName = "make.bat"
	}

	pathMake := filepath.Join(pathSrc, makeName)
	if _, err := os.Stat(pathMake); err != nil && os.IsNotExist(err) {
		log.Errorf("Could not find goPatch %s at: %s", makeName, pathMake)
		return err
	}

	if err := os.Chdir(pathSrc); err != nil {
		log.Errorf("Failed to change directory: %v", err)
		return err
	}

	// required for windows
	cwd, _ := os.Getwd()
	makePath := filepath.Join(cwd, makeName)

	err := command.RunCommand(nil, nil, command.NoTimeout, makePath)
	if err != nil {
		log.Errorf("Failed to build runtime.")
		return err
	}

	// check that it worked
	pathBin := filepath.Join(paths.GoPatch, "bin")
	if _, err := os.Stat(pathGo); err != nil && os.IsNotExist(err) {
		log.Errorf("Could not find goPatch/bin/go at: %s", pathGo)
		log.Errorf("Runtime build failed.")
		return err
	}

	if err := os.Chdir(pathBin); err != nil {
		log.Errorf("Failed to change directory: %v", err)
		return err
	}

	// required for windows
	cwd, _ = os.Getwd()
	pathBin = filepath.Join(cwd, makeName)
	err = command.RunCommand(nil, nil, command.NoTimeout, pathGo, "version")
	if err != nil {
		log.Errorf("Failed to build runtime.")
		return err
	}

	log.Info("Runtime build successfully.")

	return nil
}
