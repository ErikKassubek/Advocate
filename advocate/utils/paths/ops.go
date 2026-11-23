// Copyright (c) 2025 Erik Kassubek
//
// File: paths.go
// Brief: Operations on paths
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package paths

import (
	"advocate/utils/log"
	"os"
	"path/filepath"
	"strings"
)

func Join(pre, post bool, elem ...string) string {

	if len(elem) == 0 {
		if pre || post {
			return Seperator
		}
		return ""
	}

	res := filepath.Join(elem...)
	if pre {
		if !strings.HasPrefix(res, Seperator) {
			res = Seperator + res
		}
	}
	if post {
		if !strings.HasSuffix(res, Seperator) {
			res = res + Seperator
		}
	}
	return res
}

func ToLocal(path string) string {
	path = strings.ReplaceAll(path, "/", Seperator)
	path = strings.ReplaceAll(path, "\\", Seperator)
	return path
}

func ToUnix(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// AddLocalPathIndicator transforms a path into a local path by adding a ./ at the beginning it has non
//
// Parameter:
//   - path string: path
//
// Returns:
//   - string: path starting with ./
func AddLocalPathIndicator(path string) string {
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
