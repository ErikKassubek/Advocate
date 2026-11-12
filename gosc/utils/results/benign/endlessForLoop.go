// Copyright (c) 2025 
//
// File: falsePos.go
// Brief: Check for false positives
//
//
// License: BSD-3-Clause

package benign

import (
	"go/ast"
	"go/token"
)

func isInsideEndlessLoop(node ast.Node, parents map[ast.Node]ast.Node) bool {
	for n := node; n != nil; n = parents[n] {
		fs, ok := n.(*ast.ForStmt)
		if !ok {
			continue
		}

		if fs.Init != nil || fs.Cond != nil || fs.Post != nil {
			continue
		}

		if !loopHasExit(fs, parents) {
			return true
		}
	}

	return false
}

// loopHasExit takes for loop node and checks if it contains a return or break
//
// Parameter:
//   - loop *ast.ForStmt: the for loop node
//   - parents map[ast.Node]ast.Node: the parents map
//
// Returns:
//   - bool: true if there is no return or break, false otherwise
func loopHasExit(loop *ast.ForStmt, parents map[ast.Node]ast.Node) bool {
	foundExit := false

	ast.Inspect(loop.Body, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		// Return always exits the loop
		if _, ok := n.(*ast.ReturnStmt); ok {
			foundExit = true
			return false
		}

		// Break statements sometimes exit this loop
		if bs, ok := n.(*ast.BranchStmt); ok && bs.Tok == token.BREAK {
			if breakExitsLoop(bs, loop, parents) {
				foundExit = true
				return false
			}
		}

		// goto may leave the loop
		if bs, ok := n.(*ast.BranchStmt); ok && bs.Tok == token.GOTO {
			foundExit = true
			return false
		}

		return true
	})

	return foundExit
}

// breakExitsLoop determines whether `break` exits the given loop.
func breakExitsLoop(b *ast.BranchStmt, loop *ast.ForStmt, parents map[ast.Node]ast.Node) bool {
	// Unlabeled break → breaks nearest enclosing loop/switch/select
	if b.Label == nil {
		// If directly inside loop, it breaks the loop
		return isDirectChildOf(loop, b, parents)
	}

	// Labeled break → check if that label belongs to the loop
	label := findLabeledStmt(loop, parents)
	if label != nil && label.Label.Name == b.Label.Name {
		return true
	}

	return false
}
