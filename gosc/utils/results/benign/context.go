// Copyright (c) 2025 
//
// File: context.go
// Brief: check if the block is a context event though the cancel existed
//
//
// License: BSD-3-Clause

package benign

import (
	"go/ast"
	"go/types"
)

// isContextDoneWithCancel checks if the given node is a context done and if
// the trace contains the cancel that would have released it
//
// Parameter:
//   - n ast.Node: the node
//   - info *types.Info: the type info
//   - fileName string: the file name of the blocked element
//   - line int: the line of the blocked element
//
// Returns:
//   - bool: true if it is a context done with a cancel, false otherwise
func isContextDoneWithCancel(n ast.Node, info *types.Info, fileName string, line int,
	contextCancel map[int]struct{}, contextDone map[string]map[int]int) bool {
	if !isContextDone(n, info) {
		return false
	}

	if hasCancel(fileName, line, contextCancel, contextDone) {
		return true
	}

	return false
}

// isContextDoneCall reports whether n represents a call to context.Context.Done().
func isContextDone(n ast.Node, info *types.Info) bool {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "Done" {
		return false
	}

	tv, ok := info.Types[sel]
	if !ok || tv.Type == nil {
		return false
	}

	obj := info.Uses[sel.Sel]
	if obj == nil {
		return false
	}

	// Check method’s receiver type
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return false
	}

	recv := sig.Recv()
	if recv == nil {
		return false
	}

	return recv.Type().String() == "context.Context"
}

func hasCancel(file string, line int, contextCancel map[int]struct{}, contextDone map[string]map[int]int) bool {
	if id, okDone := contextDone[file][line]; okDone {
		if _, okCancel := contextCancel[id]; okCancel {
			return true
		}
	}

	return false
}
