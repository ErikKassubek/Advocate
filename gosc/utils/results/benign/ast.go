// Copyright (c) 2025 
//
// File: ast.go
// Brief: Functions to work with the as
//
//
// License: BSD-3-Clause

package benign

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

// parseFile creates an ast for a given file
//
// Parameter:
//   - filename string: the file name
//
// Returns:
//   - *token.FileSet: the ast file set
//   - *ast.File: the ast
//   - *types.Info: the type info
//   - error
func parseFile(fileName string) (*token.FileSet, *ast.File, *types.Info, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fileName, nil, 0)
	if err != nil {
		return nil, nil, nil, err
	}

	// Prepare type info
	conf := &types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	conf.Check("", fset, []*ast.File{file}, info)
	// Type errors are not fatal—we can still proceed
	// if err != nil {
	// 	log.Error("Warning: type checking not perfect:", err)
	// }

	return fset, file, info, nil
}

// buildParentMap creates a map that, for each node in an ast contains the parent node
//
// Parameter:
//   - root *ast.File: the ast
//
// Returns:
//   - map[ast.Node]ast.Mode: maps each node to its parent in the ast
func buildParentMap(root ast.Node) map[ast.Node]ast.Node {
	parent := map[ast.Node]ast.Node{}

	var visit func(ast.Node)
	visit = func(n ast.Node) {
		ast.Inspect(n, func(child ast.Node) bool {
			if child != nil && child != n {
				parent[child] = n
			}
			return true
		})
	}

	visit(root)
	return parent
}

// findNodeAtLine searches for the node in the ast, that represents a specific line
//
// Parameter:
//   - fset *token.FileSet: the file set
//   - file *ast.File: the ast
//   - line int: the line to search for
//
// Returns:
//   - ast.Node: the ast Node that represents to operation in the given line
func findNodeAtLine(fset *token.FileSet, file *ast.File, line int) ast.Node {
	var result ast.Node

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		start := fset.Position(n.Pos()).Line
		end := fset.Position(n.End()).Line

		if start <= line && line <= end {
			// This node contains the line.
			// Keep it and continue to find a more specific one.
			result = n
			return true
		}

		// Prune traversal if the node starts after the line
		if start > line {
			return false
		}

		return true
	})

	return result
}

// finds the label attached to a loop: "<label>: for { ... }"
func findLabeledStmt(loop *ast.ForStmt, parentMap map[ast.Node]ast.Node) *ast.LabeledStmt {
	if parent, ok := parentMap[loop].(*ast.LabeledStmt); ok {
		return parent
	}
	return nil
}

// Checks if child belongs directly to the loop body (not a deeper nested loop)
func isDirectChildOf(parent, child ast.Node, parentMap map[ast.Node]ast.Node) bool {
	for n := child; n != nil; n = parentMap[n] {
		if n == parent {
			return true
		}
		// stop at other loop constructs
		switch n.(type) {
		case *ast.ForStmt, *ast.RangeStmt, *ast.SelectStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			if n != parent {
				return false
			}
		}
	}
	return false
}
