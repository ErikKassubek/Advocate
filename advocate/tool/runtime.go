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
	"advocate/utils/log"
	"advocate/utils/osinfo"
	"advocate/utils/paths"
	"fmt"
	"os"
	"path/filepath"
)

func buildRuntime() error {
	path := filepath.Join(paths.GoPatch, "bin", osinfo.GoExec)

	if _, err := os.Stat(path); err == nil {
		var rebuild string

		for rebuild != "Y" {
			log.WriteInput("Runtime already exists. Rebuild? (Y/N): ")
			fmt.Scan(&rebuild)
			switch rebuild {
			case "Y":
			case "N":
				log.Important("Skip Build. Exit...")
				return nil
			default:
				log.Important("Invalid input.")
			}
		}
	}

	pathSrc := filepath.Join(paths.GoPatch, "src")
	if _, err := os.Stat(pathSrc); err != nil && os.IsNotExist(err) {
		log.Errorf("Could not find goPatch at: %s", pathSrc)
		return err
	}

	pathName := "./make.bash"
	if osinfo.GetOs() == osinfo.Windows {
		pathName = "make.bat"
	}

	pathMake := filepath.Join(pathSrc, pathName)
	if _, err := os.Stat(pathMake); err != nil && os.IsNotExist(err) {
		log.Errorf("Could not find goPatch %s at: %s", pathName, pathMake)
		return err
	}

	if err := os.Chdir(pathSrc); err != nil {
		log.Errorf("Failed to change directory: %v", err)
		return err
	}

	return command.RunCommand(nil, nil, command.NoTimeout, pathName)
}
