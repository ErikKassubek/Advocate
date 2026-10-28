// Copyright (c) 2026 Erik Kassubek
//
// File: experiments.go
// Brief: Entry point for experiments
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package experiments

import (
	"advocate/experiments/indexing"
	"advocate/utils/flags"
	"fmt"
)

func Run() error {
	if flags.Experiment == "" {
		return fmt.Errorf("No experiment selected. Select with `-experiment [name]`")
	}

	switch flags.ExecName {
	case "indexing":
		indexing.Run()
	default:
		return fmt.Errorf("Unknown experiment: %s", flags.ExecName)
	}

	return nil
}
