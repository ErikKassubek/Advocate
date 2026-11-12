// Copyright (c) 2025 
//
// File: io.go
// Brief: write the constraints to a file
//
//
// License: BSD-3-Clause

package a_constraints

import (
	"bufio"
	"fmt"
	"os"
)

// Write the constraints to a file that can be used for the execution
//
// Parameter:
//   - path string: path for the file
//
// Returns:
//   - error
func writeToSting(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", path, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range allConstraints {
		_, err := writer.WriteString(line.toString() + "\n")
		if err != nil {
			return fmt.Errorf("could not write line: %w", err)
		}
	}
	return writer.Flush()
}
