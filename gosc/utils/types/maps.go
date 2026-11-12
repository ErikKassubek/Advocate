// Copyright (c) 2025 
//
// File: maps.go
// Brief: Helper functions for maps
//
//
// License: BSD-3-Clause

package types

func MergeMaps[K comparable, V any](dst, src map[K]V) {
	for k, v := range src {
		dst[k] = v
	}
}
