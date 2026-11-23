// Copyright (c) 2026 Erik Kassubek
//
// File: maths.go
// Brief: Maths functions
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package math

import "strconv"

type NumberType interface {
	~int | ~int64 | ~float32 | ~float64
}

// If low <= high:
//
//	low if v is less than low,
//	or high if high is less than v,
//	otherwise v.
//
// Else:
//
//	v
func Clamp[T NumberType](v, low, high T) T {
	if low > high {
		return v
	}
	return max(min(v, high), low)
}

func ToNum[T NumberType](s string) T {
	var zero T

	switch any(zero).(type) {
	case int:
		v, err := strconv.Atoi(s)
		if err != nil {
			return T(0)
		}
		return T(v)

	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return T(0)
		}
		return T(v)

	case float32:
		v, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return T(0)
		}
		return T(v)

	case float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return T(0)
		}
		return T(v)
	}

	return T(-1)
}

func ToString[T NumberType | string](v T) string {
	switch x := any(v).(type) {
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case string:
		return x
	default:
		panic("unsupported type")
	}
}
