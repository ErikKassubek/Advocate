// Copyright (c) 2026 Erik Kassubek
//
// File: osinfo.go
// Brief: get os info
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package osinfo

import (
	"runtime"
)

// os dependent
var (
	Os     = GetOs()
	GoExec = map[bool]string{
		true:  "go.exe",
		false: "go",
	}[Os == Windows]
)

type OsName string

const (
	Linux   OsName = "linux"
	Windows OsName = "windows"
	MacOs   OsName = "darwin"
	FreeBSD OsName = "freebsd"
	Other   OsName = ""
)

func GetOs() OsName {
	switch runtime.GOOS {
	case "linux":
		return Linux
	case "windows":
		return Windows
	case "darwin":
		return MacOs
	case "freebsd":
		return FreeBSD
	default:
		return Other
	}
}
