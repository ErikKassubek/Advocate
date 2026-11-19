// Copyright (c) 2024 Erik Kassubek
//
// File: output.go
// Brief: Handle output file
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package files

import (
	"advocate/utils/paths"
	"os"
)

func OpenOutput() (*os.File, error) {
	return os.OpenFile(paths.NameOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func IsOpen(f *os.File) bool {
	if f == nil {
		return false
	}

	_, err := f.Stat()
	return err == nil
}
