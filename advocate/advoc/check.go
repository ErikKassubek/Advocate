// Copyright (c) 2026 Erik Kassubek
//
// File: check.go
// Brief: run some checks
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package advoc

import (
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CheckBin() error {
	path := filepath.Join(paths.GoPatch, "bin", "go")
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		pathMake := filepath.Join(paths.GoPatch, "src")
		return fmt.Errorf("Could not find %s. Run make.bash or make.bat in %s before running advocate.", path, pathMake)
	}

	return nil
}

// CheckProg checks the version of the program to be analyzed and finds the exec name
// Advocate is implemented in and for go1.27. It the analyzed program has another
// version, especially if the other version is also installed on the machine,
// this can lead to problems. checkGoMod therefore reads the version of the
// analyzed program and if its not 1.27, a warning and information is printed
// to the terminal
// Additionally it reads the module name from the go.mod file.
// If -main is set, but -exec is not set it will try to set the
// execname value. If no module value is found, the program will panic
//
// Returns:
//   - string: exec name, or empty if not found
func CheckProg() string {
	var goModPath string

	if flags.ProgPath == "" {
		return flags.ExecName
	}

	var err error
	flags.ProgPath, err = paths.CheckPath(flags.ProgPath)
	if err != nil {
		log.Errorf("Could not find %s", flags.ProgPath)
		return ""
	}

	if flags.RootPath == "" {
		flags.RootPath = paths.GetDirectory(flags.ProgPath)
	}

	// Search for go.mod
	err = filepath.WalkDir(paths.GetDirectory(flags.ProgPath), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == "go.mod" {
			goModPath = path
			return filepath.SkipAll // Stop searching after finding the first one
		}
		return nil
	})

	if goModPath == "" {
		log.Error("Could not find go.mod")
		return flags.ExecName
	}

	// Open and read go.mod
	file, err := os.Open(goModPath)
	if err != nil {
		log.Error("Could not find go.mod")
		return flags.ExecName
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		timeLine := strings.TrimSpace(line)

		// check for module name
		if flags.ModeMain && flags.ExecName == "" && strings.HasPrefix(timeLine, "module") {
			s := strings.Split(timeLine, " ")
			if len(s) >= 2 {
				flags.ExecName = s[1]
			}
		}

		// check for version
		if strings.HasPrefix(timeLine, "go ") {
			version := strings.TrimSpace(strings.TrimPrefix(timeLine, "go "))

			versionSplit := strings.Split(version, ".")

			if len(versionSplit) < 2 {
				log.Error("Invalid go version")
			}

			if len(versionSplit) > 2 {
				version = versionSplit[0] + "." + versionSplit[1]
				line = "go " + version
				log.Importantf("Updated Go version in go.mod to %s", version)
			}

			if versionSplit[0] != "1" || versionSplit[1] != "27" {
				errString := "ADVOCATE is implemented for go version 1.27.\n"
				errString += fmt.Sprintf("Found version %s in go.mod.\n", version)
				errString += fmt.Sprintf("This may result in the analysis not working correctly, especially if go %s.%s is installed on the computer.\n", versionSplit[0], versionSplit[1])
				errString += "The message 'package advocate is not in std' in the output.log file may indicate this.\n"
				errString += "If this message appears it may help to remove the minor version number in the go.mod file of the analyzed program (e.g. in go 1.27.0 -> 1.27)\n"
				// errString += `'/home/.../go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/advocate' or 'package advocate is not in std' in the output files may indicate an incompatible go version.`
				log.Important(errString)
			}
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error reading go.mod:", err)
		return flags.ExecName
	}

	// rewrite the file with updated lines
	err = os.WriteFile(goModPath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	if err != nil {
		log.Error("Failed to update go.mod:", err)
	}

	return flags.ExecName
}
