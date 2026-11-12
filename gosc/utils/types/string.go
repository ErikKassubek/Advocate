// Copyright (c) 2025 
//
// File: pair.go
// Brief: Implement a set data type
//
//
// License: BSD-3-Clause

package types

import "strings"

func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
