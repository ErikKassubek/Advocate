// Copyright (c) 2025 Erik Kassubek
//
// File: settings.go
// Brief: Constants that can be set via the setting flag
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package settings

import (
	"advocate/utils/flags"
	"advocate/utils/log"
	"math"
	"strconv"
	"strings"
)

// constants for fuzzing calculations
var (
	GFuzzW1 = 10.0
	GFuzzW2 = 10.0
	GFuzzW3 = 10.0
	GFuzzW4 = 10.0

	GFuzzFlipP    = 0.99
	GFuzzFlipPMin = 0.1

	GoPieW1          = 1.0
	GoPieW2          = 1.0
	GoPieMaxSCLength = 9
	ChainMutabound   = 10

	GoPieSCStart = 10

	SameElementTypeInSC = false
	WithoutReplay       = false
)

// SetSettings sets different constants and settings used in the program
// from the settings flag
func SetSettings() {
	if flags.FuzzingMode != "GoPie" {
		ChainMutabound = 2 * min(int(flags.MaxFuzzingRun/GoPieSCStart), ChainMutabound)
	}

	if flags.Settings == "" {
		return
	}

	sets := strings.Split(flags.Settings, ",")

	for _, elem := range sets {
		elemParts := strings.Split(elem, "=")
		if len(elemParts) != 2 {
			log.Errorf("Invalid setting '%s'. Skip.", elem)
			continue
		}

		value, err := strconv.ParseFloat(elemParts[1], 64)
		if err != nil {
			log.Errorf("Value in %s cannot be converted to number. Skip.", elem)
			continue
		}

		switch elemParts[0] {
		case "GFuzzW1":
			GFuzzW1 = value
		case "GFuzzW2":
			GFuzzW2 = value
		case "GFuzzW3":
			GFuzzW3 = value
		case "GFuzzW4":
			GFuzzW4 = value
		case "GFuzzFlipP":
			GFuzzFlipP = clamp(value, 0, 1)
		case "GFuzzFlipPMin":
			GFuzzFlipPMin = clamp(value, 0, 1)
		case "GoPieW1":
			GoPieW1 = value
		case "GoPieW2":
			GoPieW2 = value
		case "GoPieBound":
			GoPieMaxSCLength = int(clamp(value, 2.0, math.MaxFloat64))
		case "GoPieMutabound":
			ChainMutabound = int(clamp(value, 1, math.MaxFloat64))
		case "GoPieSCStart":
			GoPieSCStart = int(clamp(value, 1, math.MaxFloat64))
		case "SameElementTypeInSC":
			if value == 1 {
				SameElementTypeInSC = true
			}
		case "WithoutReplay":
			if value == 1 {
				WithoutReplay = true
			}

		default:
			log.Errorf("Unknown name in setting %s", elemParts[0])
			continue
		}
	}
}

// Return the value. If it is bigger than maxVal, return maxVal. If it is smaller
// than minVal, return minVal
//
// Parameter:
//   - value float64: the value
//   - minVal flat64: the minimum value
//   - maxVal flat64: the maximum value
//
// Returns:\
//   - the value clamped between minVal and maxVal
func clamp(value, minVal, maxVal float64) float64 {
	if value >= maxVal {
		return maxVal
	} else if value <= minVal {
		return minVal
	}
	return value
}
