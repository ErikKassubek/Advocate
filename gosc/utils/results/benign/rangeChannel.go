// Copyright (c) 2025 
//
// File: rangeChannel.go
// Brief: Check if the blocking bug is in a range over channel
//
//
// License: BSD-3-Clause

package benign

import (
	"go/ast"
	"go/token"
	"go/types"
)

func isRangeOverChan(node ast.Node, parents map[ast.Node]ast.Node, info *types.Info) (bool, error) {
	for n := node; n != nil; n = parents[n] {
		rs, ok := n.(*ast.RangeStmt)
		if !ok {
			continue
		}

		// Use type checking (most reliable)
		if info != nil {
			if tv, ok := info.Types[rs.X]; ok {
				if _, ok := tv.Type.Underlying().(*types.Chan); ok {
					return true, nil
				}
			}
		}

		// Fallback AST-based heuristics
		if exprLooksLikeChannel(rs.X) {
			return true, nil
		}
	}

	return false, nil
}

func exprLooksLikeChannel(expr ast.Expr) bool {
	switch e := expr.(type) {

	case *ast.ChanType:
		return true

	case *ast.UnaryExpr:
		return e.Op == token.ARROW // <-ch

	case *ast.Ident:
		// Cannot guarantee anything without type info
		return false

	case *ast.CallExpr:
		// Weak heuristic for names ending with "Chan"/"Channel"
		if ident, ok := e.Fun.(*ast.Ident); ok {
			if containsChanLikeName(ident.Name) {
				return true
			}
		}
	}

	return false
}

func containsChanLikeName(name string) bool {
	return name == "chan" ||
		name == "makeChan" ||
		name == "newChan" ||
		len(name) > 4 && name[len(name)-4:] == "Chan"
}
