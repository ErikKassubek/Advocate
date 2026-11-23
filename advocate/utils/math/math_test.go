// Copyright (c) 2025 Erik Kassubek
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package math

import (
	"testing"
)

func TestClampInt(t *testing.T) {
	tests := []struct {
		name     string
		v        int
		low      int
		high     int
		expected int
	}{
		{"below lower bound", -5, 0, 10, 0},
		{"at lower bound", 0, 0, 10, 0},
		{"within range", 5, 0, 10, 5},
		{"at upper bound", 10, 0, 10, 10},
		{"above upper bound", 15, 0, 10, 10},
		{"negative range", -15, -10, -5, -10},
		{"negative value within range", -7, -10, -5, -7},
		{"lower bound equal upper bound", 5, 10, 10, 10},
		{"lower bound greater than upper bound, value lower", 5, 15, 10, 5},
		{"lower bound greater than upper bound, value greater", 5, 15, 10, 5},
		{"lower bound greater than upper bound, value greater", 12, 15, 10, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.v, tt.low, tt.high)
			if tt.expected != result {
				t.Fatalf("Clamp(%d, %d, %d), expected: %d, got: %d", tt.v, tt.low, tt.high, tt.expected, result)
			}
		})
	}
}

func TestClampInt64(t *testing.T) {
	tests := []struct {
		name     string
		v        int64
		low      int64
		high     int64
		expected int64
	}{
		{"below lower bound", -5, 0, 10, 0},
		{"at lower bound", 0, 0, 10, 0},
		{"within range", 5, 0, 10, 5},
		{"at upper bound", 10, 0, 10, 10},
		{"above upper bound", 15, 0, 10, 10},
		{"negative range", -15, -10, -5, -10},
		{"negative value within range", -7, -10, -5, -7},
		{"lower bound equal upper bound", 5, 10, 10, 10},
		{"lower bound greater than upper bound, value lower", 5, 15, 10, 5},
		{"lower bound greater than upper bound, value greater", 5, 15, 10, 5},
		{"lower bound greater than upper bound, value greater", 12, 15, 10, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.v, tt.low, tt.high)
			if tt.expected != result {
				t.Fatalf("Clamp(%d, %d, %d), expected: %d, got: %d", tt.v, tt.low, tt.high, tt.expected, result)
			}
		})
	}
}

func TestClampFloat32(t *testing.T) {
	tests := []struct {
		name     string
		v        float32
		low      float32
		high     float32
		expected float32
	}{
		{"below lower bound", -5, 0, 10, 0},
		{"at lower bound", 0, 0, 10, 0},
		{"within range", 5, 0, 10, 5},
		{"at upper bound", 10, 0, 10, 10},
		{"above upper bound", 15, 0, 10, 10},
		{"negative range", -15, -10, -5, -10},
		{"negative value within range", -7, -10, -5, -7},
		{"lower bound equal upper bound", 5, 10, 10, 10},
		{"lower bound greater than upper bound, value lower", 5, 15, 10, 5},
		{"lower bound greater than upper bound, value greater", 5, 15, 10, 5},
		{"lower bound greater than upper bound, value greater", 12, 15, 10, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.v, tt.low, tt.high)
			if tt.expected != result {
				t.Fatalf("Clamp(%f, %f, %f), expected: %f, got: %f", tt.v, tt.low, tt.high, tt.expected, result)
			}
		})
	}
}

func TestClampFloat64(t *testing.T) {
	tests := []struct {
		name     string
		v        float64
		low      float64
		high     float64
		expected float64
	}{
		{"below lower bound", -5, 0, 10, 0},
		{"at lower bound", 0, 0, 10, 0},
		{"within range", 5, 0, 10, 5},
		{"at upper bound", 10, 0, 10, 10},
		{"above upper bound", 15, 0, 10, 10},
		{"negative range", -15, -10, -5, -10},
		{"negative value within range", -7, -10, -5, -7},
		{"lower bound equal upper bound", 5, 10, 10, 10},
		{"lower bound greater than upper bound, value lower", 5, 15, 10, 5},
		{"lower bound greater than upper bound, value greater", 5, 15, 10, 5},
		{"lower bound greater than upper bound, value greater", 12, 15, 10, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.v, tt.low, tt.high)
			if tt.expected != result {
				t.Fatalf("Clamp(%f, %f, %f), expected: %f, got: %f", tt.v, tt.low, tt.high, tt.expected, result)
			}
		})
	}
}

func TestToNumInt(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected int
	}{
		{"zero", "0", 0},
		{"positive one digid", "4", 4},
		{"positive multi digid", "438", 438},
		{"negative one digid", "-4", -4},
		{"negative multi digid", "-438", -438},
		{"positive leading 0", "014", 14},
		{"negative leading 0", "-014", -14},
		{"not a number", "asd", 0},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToNum[int](tt.s)
			if tt.expected != result {
				t.Fatalf("ToNum(%s), expected: %d, got: %d", tt.s, tt.expected, result)
			}
		})
	}
}

func TestToNumInt64(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected int64
	}{
		{"zero", "0", 0},
		{"positive one digid", "4", 4},
		{"positive multi digid", "438", 438},
		{"negative one digid", "-4", -4},
		{"negative multi digid", "-438", -438},
		{"positive leading 0", "014", 14},
		{"negative leading 0", "-014", -14},
		{"not a number", "asd", 0},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToNum[int64](tt.s)
			if tt.expected != result {
				t.Fatalf("ToNum(%s), expected: %d, got: %d", tt.s, tt.expected, result)
			}
		})
	}
}

func TestToNumFloat32(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected float32
	}{
		{"zero", "0", 0},
		{"positive int one digid", "4", 4},
		{"positive int multi digid", "438", 438},
		{"negative int one digid", "-4", -4},
		{"negative int multi digid", "-438", -438},
		{"positive int leading 0", "014", 14},
		{"negative int leading 0", "-014", -14},

		{"positive float one digid", "4.123", 4.123},
		{"positive float multi digid", "438.001", 438.001},
		{"negative float one digid", "-4.1", -4.1},
		{"negative float multi digid", "-438.5", -438.5},
		{"positive float leading 0", "014.02", 14.02},
		{"negative float leading 0", "-014.4", -14.4},

		{"not a number", "asd", 0},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToNum[float32](tt.s)
			if tt.expected != result {
				t.Fatalf("ToNum(%s), expected: %f, got: %f", tt.s, tt.expected, result)
			}
		})
	}
}

func TestToNumFloat64(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected float64
	}{
		{"zero", "0", 0},
		{"positive int one digid", "4", 4},
		{"positive int multi digid", "438", 438},
		{"negative int one digid", "-4", -4},
		{"negative int multi digid", "-438", -438},
		{"positive int leading 0", "014", 14},
		{"negative int leading 0", "-014", -14},

		{"positive float one digid", "4.123", 4.123},
		{"positive float multi digid", "438.001", 438.001},
		{"negative float one digid", "-4.1", -4.1},
		{"negative float multi digid", "-438.5", -438.5},
		{"positive float leading 0", "014.02", 14.02},
		{"negative float leading 0", "-014.4", -14.4},

		{"not a number", "asd", 0},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToNum[float64](tt.s)
			if tt.expected != result {
				t.Fatalf("ToNum(%s), expected: %f, got: %f", tt.s, tt.expected, result)
			}
		})
	}
}

func TestToStringInt(t *testing.T) {
	tests := []struct {
		name     string
		v        int
		expected string
	}{
		{"zero", 0, "0"},
		{"positive int one digid", 4, "4"},
		{"positive int multi digid", 438, "438"},
		{"negative int one digid", -4, "-4"},
		{"negative int multi digid", -438, "-438"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.v)
			if tt.expected != result {
				t.Fatalf("ToString(%d), expected: %s, got: %s", tt.v, tt.expected, result)
			}
		})
	}
}

func TestToStringInt64(t *testing.T) {
	tests := []struct {
		name     string
		v        int64
		expected string
	}{
		{"zero", 0, "0"},
		{"positive int one digid", 4, "4"},
		{"positive int multi digid", 438, "438"},
		{"negative int one digid", -4, "-4"},
		{"negative int multi digid", -438, "-438"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.v)
			if tt.expected != result {
				t.Fatalf("ToString(%d), expected: %s, got: %s", tt.v, tt.expected, result)
			}
		})
	}
}

func TestToStringFloat32(t *testing.T) {
	tests := []struct {
		name     string
		v        float32
		expected string
	}{
		{"zero", 0, "0"},
		{"positive int one digid", 4, "4"},
		{"positive int multi digid", 438, "438"},
		{"negative int one digid", -4, "-4"},
		{"negative int multi digid", -438, "-438"},

		{"positive float one digid", 4.5, "4.5"},
		{"positive float multi digid", 438.5, "438.5"},
		{"negative float one digid", -4.5, "-4.5"},
		{"negative float multi digid", -438.5, "-438.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.v)
			if tt.expected != result {
				t.Fatalf("ToString(%f), expected: %s, got: %s", tt.v, tt.expected, result)
			}
		})
	}
}

func TestToStringFloat64(t *testing.T) {
	tests := []struct {
		name     string
		v        float64
		expected string
	}{
		{"zero", 0, "0"},
		{"positive int one digid", 4, "4"},
		{"positive int multi digid", 438, "438"},
		{"negative int one digid", -4, "-4"},
		{"negative int multi digid", -438, "-438"},

		{"positive float one digid", 4.5, "4.5"},
		{"positive float multi digid", 438.5, "438.5"},
		{"negative float one digid", -4.5, "-4.5"},
		{"negative float multi digid", -438.5, "-438.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.v)
			if tt.expected != result {
				t.Fatalf("ToString(%f), expected: %s, got: %s", tt.v, tt.expected, result)
			}
		})
	}
}

func TestToStringString(t *testing.T) {
	tests := []struct {
		name     string
		v        string
		expected string
	}{
		{"zero", "0", "0"},
		{"positive int one digid", "4", "4"},
		{"positive int multi digid", "438", "438"},
		{"negative int one digid", "-4", "-4"},
		{"negative int multi digid", "-438", "-438"},

		{"positive float one digid", "4.5", "4.5"},
		{"positive float multi digid", "438.5", "438.5"},
		{"negative float one digid", "-4.5", "-4.5"},
		{"negative float multi digid", "-438.5", "-438.5"},

		{"not a number", "asd", "asd"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.v)
			if tt.expected != result {
				t.Fatalf("ToString(%s), expected: %s, got: %s", tt.v, tt.expected, result)
			}
		})
	}
}
