// Copyright (c) 2026 Erik Kassubek
//
// File: values.go
// Brief: get os dependent values
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package osinfo

// os dependent
var (
	Os     = GetOs()
	GoExec = map[bool]string{
		true:  "go.exe",
		false: "go",
	}[Os == Windows]
)
