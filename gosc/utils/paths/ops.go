// Copyright (c) 2025 
//
// File: paths.go
// Brief: Operations on paths
//
//
// License: BSD-3-Clause

package paths

import (
	"gosc/utils/consts"
	"gosc/utils/log"
	"os"
	"path/filepath"
	"strings"
)

func Join(pre, post bool, elem ...string) string {
	res := filepath.Join(elem...)
	if pre {
		res = consts.Sep + res
	}
	if post {
		res = res + consts.Sep
	}
	return res
}

func ToLocal(path string) string {
	path = strings.ReplaceAll(path, "/", consts.Sep)
	path = strings.ReplaceAll(path, "\\", consts.Sep)
	return path
}

func ToUnix(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// MakePathLocal transforms a path into a local path by adding a ./ at the beginning it has non
//
// Parameter:
//   - path string: path
//
// Returns:
//   - string: path starting with ./
func MakePathLocal(path string) string {
	pathSep := string(os.PathSeparator)

	// ./path
	if strings.HasPrefix(path, "."+pathSep) {
		return path
	}

	// /path
	if strings.HasPrefix(path, pathSep) {
		return "." + path
	}

	// path
	return "." + pathSep + path
}

// CleanPathHome takes a path containing a ~ and replaces it with the
// path to the home folder
func CleanPathHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Error(err.Error())
	}

	return strings.Replace(path, "~", home, -1)
}
