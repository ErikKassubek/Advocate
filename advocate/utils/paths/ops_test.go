// Copyright (c) 2025 Erik Kassubek
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package paths

import (
	"os"
	"testing"
)

func TestJoin(t *testing.T) {
	tests := []struct {
		name     string
		pre      bool
		post     bool
		elem     []string
		expected string
	}{
		{"valid, no pre, no post, no seperators", false, false, []string{"a", "b", "c"}, "a/b/c"},
		{"valid, pre, no post, no seperators", true, false, []string{"a", "b", "c"}, "/a/b/c"},
		{"valid, no pre, post, no seperators", false, true, []string{"a", "b", "c"}, "a/b/c/"},
		{"valid, pre, post, no seperators", true, true, []string{"a", "b", "c"}, "/a/b/c/"},

		{"valid, pre, post, seperators in between", true, true, []string{"a/", "/b/", "/c"}, "/a/b/c/"},
		{"valid, pre, post, seperators at start", true, true, []string{"/a", "b", "c"}, "/a/b/c/"},
		{"valid, pre, post, seperators at end", true, true, []string{"a", "b", "c/"}, "/a/b/c/"},
		{"valid, pre, post, seperators at start and end", true, true, []string{"/a", "b", "c/"}, "/a/b/c/"},

		{"empty, no pre, no post", false, false, []string{}, ""},
		{"empty, pre, no post", true, false, []string{}, "/"},
		{"empty, no pre, post", false, true, []string{}, "/"},
		{"empty, pre, post", true, true, []string{}, "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Join(tt.pre, tt.post, tt.elem...)
			if tt.expected != result {
				t.Fatalf("Join(%t, %t, %s), expected: %s, got: %s", tt.pre, tt.post, tt.elem, tt.expected, result)
			}
		})
	}
}

func TestToLocal(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		sep      string
		expected string
	}{
		{"windows to unix", "\\a\\b\\c\\", "/", "/a/b/c/"},
		{"windows to unix, no sep", "a", "/", "a"},
		{"windows to unix, empty", "", "/", ""},

		{"unix to windows", "/a/b/c/", "\\", "\\a\\b\\c\\"},
		{"unix to windows, no sep", "a", "\\", "a"},
		{"unix to windows, empty", "", "\\", ""},

		{"windows to windows", "\\a\\b\\c\\", "\\", "\\a\\b\\c\\"},
		{"windows to windows, no sep", "a", "\\", "a"},
		{"windows to windows, empty", "", "\\", ""},

		{"unix to unix", "/a/b/c/", "/", "/a/b/c/"},
		{"unix to unix, no sep", "a", "/", "a"},
		{"unix to unix, empty", "", "/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Seperator = tt.sep
			result := ToLocal(tt.path)
			if tt.expected != result {
				t.Fatalf("ToLocal(%s), expected: %s, got: %s", tt.path, tt.expected, result)
			}
		})
	}
}

func TestToUnix(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"windows to unix", "\\a\\b\\c\\", "/a/b/c/"},
		{"windows to unix, no sep", "a", "a"},
		{"windows to unix, empty", "", ""},

		{"unix to unix", "/a/b/c/", "/a/b/c/"},
		{"unix to unix, no sep", "a", "a"},
		{"unix to unix, empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUnix(tt.path)
			if tt.expected != result {
				t.Fatalf("ToUnix(%s), expected: %s, got: %s", tt.path, tt.expected, result)
			}
		})
	}
}

func TestAddLocalPathIndicator(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"with local", "./a/b", "./a/b"},
		{"with seperator", "/a/b", "./a/b"},
		{"without", "a/b", "./a/b"},
		{"empty", "", "./"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddLocalPathIndicator(tt.path)
			if tt.expected != result {
				t.Fatalf("AddLocalPathIndicator(%s), expected: %s, got: %s", tt.path, tt.expected, result)
			}
		})
	}
}

func TestCleanPathHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"normal path", "~/a/b", home + "/a/b"},
		{"without ~", "/a/b", "/a/b"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanPathHome(tt.path)
			if tt.expected != result {
				t.Fatalf("CleanPathHome(%s), expected: %s, got: %s", tt.path, tt.expected, result)
			}
		})
	}
}
