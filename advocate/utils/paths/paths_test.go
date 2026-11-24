// Copyright (c) 2025 Erik Kassubek
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package paths

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProjectRoot(t *testing.T) {
	root, err := projectRoot()
	if err != nil {
		t.Fatal("Could not get project root")
	}

	if !strings.HasSuffix(root, "/Advocate/advocate") {
		t.Fatal("Incorrect project root: expected suffix /Advocate/advocate, got ", root)
	}
}

func TestPathsAdvocate(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename))))

	PathsAdvocate()

	tests := []struct {
		name     string
		val      string
		expected string
	}{
		{"Advocate", Advocate, root},
		{"GoPatch", GoPatch, root + "/goPatch"},
		{"Go", Go, root + "/goPatch/bin/go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected != tt.val {
				t.Fatalf("PathsAdvocate(), expected: %s, got: %s", tt.expected, tt.val)
			}
		})
	}
}

// Cont
