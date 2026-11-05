// Copyright (c) 2026 Erik Kassubek
//
// File: code.go
// Brief: Directly interact with the code
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_code

import (
	"advocate/trace"
	"bufio"
	"fmt"
	"os"
)

func GetLineContent(pos trace.Position) (string, error) {
	f, err := os.Open(pos.File())
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(f)
	for current := 1; scanner.Scan(); current++ {
		if current == pos.Line() {
			return scanner.Text(), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("line %d not found in %s", pos.Line(), pos.File())
}
