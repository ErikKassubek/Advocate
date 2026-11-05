// ADVOCATE-FILE-START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate.go
// Brief: Insert recording for mutex, cond var and wait group creation
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package walk

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
	"strings"
)

// ==================================================
// MARK: instrument
// ==================================================

func instrumentBody(fn *ir.Func) {
	isAdvocate := base.Flag.AdvocateTrace || base.Flag.AdvocateReplay || base.Flag.AdvocateFuzzing
	if !isAdvocate {
		return
	}

	if isUserMain(fn) {
		fn.Body = addInit(fn.Body, fn.Pos())
	}

	if !shouldAdvocate(fn) {
		return
	}

	instrumentParameterCopy(fn)

	fn.Body = instrumentStmtList(fn.Body)
}

func addInit(body ir.Nodes, pos src.XPos) ir.Nodes {
	out := make(ir.Nodes, 0)

	if base.Flag.AdvocateTrace {
		fn := typecheck.LookupRuntime("AdvocateInitTracing")
		out.Append(typecheck.Call(
			pos,
			fn,
			[]ir.Node{
				ir.NewInt(pos, int64(base.Flag.AdvocateTimeout)),
				ir.NewBool(pos, false),
			},
			false,
		))
	} else if base.Flag.AdvocateReplay {
		fn := typecheck.LookupRuntime("AdvocateInitReplay")
		out.Append(typecheck.Call(
			pos,
			fn,
			[]ir.Node{
				ir.NewString(pos, base.Flag.AdvocatePath),
				ir.NewInt(pos, int64(base.Flag.AdvocateTimeout)),
				ir.NewBool(pos, true),
			},
			false,
		))
	} else if base.Flag.AdvocateFuzzing {
		fn := typecheck.LookupRuntime("AdvocateInitFuzzing")
		out.Append(typecheck.Call(
			pos,
			fn,
			[]ir.Node{
				ir.NewString(pos, base.Flag.AdvocatePath),
				ir.NewInt(pos, int64(base.Flag.AdvocateTimeout)),
				ir.NewBool(pos, false),
			},
			false,
		))
	} else {
		return body
	}

	out.Append(body...)

	return out
}

func instrumentStmtList(body ir.Nodes) ir.Nodes {
	out := make(ir.Nodes, 0, len(body)*2)

	for _, stmt := range body {
		instrumentStmtRecursive(stmt)

		out.Append(stmt)

		if n := addAlloc(stmt); n != nil {
			out.Append(n)
		}
	}

	return out
}

func instrumentStmtRecursive(n ir.Node) {
	if n == nil {
		return
	}

	switch x := n.(type) {
	case *ir.BlockStmt:
		x.List = instrumentStmtList(x.List)

	case *ir.IfStmt:
		instrumentIfChain(x)

	case *ir.ForStmt:
		x.Cond = instrumentExprRecursive(x.Cond)
		x.Body = instrumentStmtList(x.Body)
		x.Body = instrumentLoop(x.Body, x.Pos())

	case *ir.RangeStmt:
		x.X = instrumentExprRecursive(x.X)
		x.Body = instrumentStmtList(x.Body)
		x.Body = instrumentLoop(x.Body, x.Pos())

	case *ir.SwitchStmt:
		instrumentSwitch(x)

	case *ir.SelectStmt:
		for _, c := range x.Compiled {
			if c, ok := c.(*ir.CommClause); ok {
				c.Body = instrumentStmtList(c.Body)
			}
		}

	default:
		// Let expression traversal handle assignments, calls, literals, etc.
		instrumentExprRecursive(n)
	}
}

func instrumentExprRecursive(n ir.Node) ir.Node {
	if n == nil {
		return nil
	}

	switch n := n.(type) {

	case *ir.CallExpr:
		n.Fun = instrumentExprRecursive(n.Fun)

		for i, arg := range n.Args {
			n.Args[i] = instrumentExprRecursive(arg)
		}

	case *ir.BinaryExpr:
		n.X = instrumentExprRecursive(n.X)
		n.Y = instrumentExprRecursive(n.Y)

	case *ir.UnaryExpr:
		n.X = instrumentExprRecursive(n.X)

	case *ir.ConvExpr:
		n.X = instrumentExprRecursive(n.X)

	case *ir.AddrExpr:
		n.X = instrumentExprRecursive(n.X)

	case *ir.SelectorExpr:
		n.X = instrumentExprRecursive(n.X)

	case *ir.IndexExpr:
		n.X = instrumentExprRecursive(n.X)
		n.Index = instrumentExprRecursive(n.Index)

	case *ir.SliceExpr:
		n.X = instrumentExprRecursive(n.X)
		n.Low = instrumentExprRecursive(n.Low)
		n.High = instrumentExprRecursive(n.High)
		n.Max = instrumentExprRecursive(n.Max)

	case *ir.CompLitExpr:
		for i, e := range n.List {
			n.List[i] = instrumentExprRecursive(e)
		}

	case *ir.StructKeyExpr:
		n.Value = instrumentExprRecursive(n.Value)

	}

	return n
}

// ==================================================
// MARK: Alloc
// ==================================================

func addAlloc(n ir.Node) ir.Node {
	t, obj := allocatedSync(n)
	if t == nil || obj == nil {
		return nil
	}

	var runtimeName string

	switch {
	case isSyncType(t, "Mutex") || isSyncType(t, "RWMutex"):
		runtimeName = "AdvocateAllocMutex"

	case isSyncType(t, "Cond"):
		runtimeName = "AdvocateAllocCondVar"

	case isSyncType(t, "WaitGroup"):
		runtimeName = "AdvocateAllocWG"

	case isSyncType(t, "Once"):
		runtimeName = "AdvocateAllocOnce"

	default:
		return nil
	}

	fn := typecheck.LookupRuntime(runtimeName)

	addr := ir.NewAddrExpr(
		n.Pos(),
		obj,
	)

	addr.SetType(types.NewPtr(obj.Type()))
	addr.SetTypecheck(1)

	unsafeAddr := makeUnsafePointer(addr, n.Pos())

	return typecheck.Call(
		n.Pos(),
		fn,
		[]ir.Node{unsafeAddr},
		false,
	)
}

func makeUnsafePointer(addr ir.Node, pos src.XPos) ir.Node {
	ptrType := types.Types[types.TUNSAFEPTR]

	return ir.NewConvExpr(
		pos,
		ir.OCONV,
		ptrType,
		addr,
	)
}

func allocatedSync(n ir.Node) (*types.Type, ir.Node) {
	switch x := n.(type) {

	case *ir.AssignStmt:
		// m := sync.Mutex{}
		if x.X == nil {
			return nil, nil
		}

		return x.X.Type(), x.X

	// TODO: what if only var m sync.Mutex
	case *ir.Decl:
		// var m sync.Mutex
		// if x.X == nil {
		// 	return nil, nil
		// }

		// return x.X.Type(), x.X
	}

	return nil, nil
}

// func isSyncType(t *types.Type, name string) bool {
// 	if t == nil {
// 		return false
// 	}

// 	sym := t.Sym()
// 	if sym == nil {
// 		return false
// 	}

// 	if sym.Pkg.Path == "sync" {
// 		fmt.Println(sym.Pkg.Path, sym.Name)
// 	}

// 	return sym.Pkg != nil &&
// 		sym.Pkg.Path == "sync" &&
// 		sym.Name == name
// }

func isSyncType(t *types.Type, name string) bool {
	if t == nil {
		return false
	}

	sym := t.Sym()
	if sym == nil || sym.Pkg == nil {
		return false
	}

	return sym.Pkg.Path == "sync" &&
		sym.Name == name
}

// ==================================================
// MARK: If
// ==================================================

func addControllRec(body ir.Nodes, pos src.XPos, numCases, caseNum int, t string, start bool) ir.Nodes {
	if !base.Flag.AdvocateRecordPlus {
		return body
	}

	p := base.Ctxt.PosTable.Pos(pos)
	if advocateIgnore(p.Filename()) {
		return body
	}

	fn := typecheck.LookupRuntime("advocateControllFlow")

	call := typecheck.Call(
		pos,
		fn,
		[]ir.Node{
			ir.NewString(pos, t),
			ir.NewInt(pos, int64(numCases)),
			ir.NewInt(pos, int64(caseNum)),
		},
		false,
	)

	out := make(ir.Nodes, 0, len(body)+1)
	out.Append(call)
	out.Append(body...)

	return out
}

func instrumentIfChain(n *ir.IfStmt) {
	if !base.Flag.AdvocateRecordPlus {
		return
	}

	numCases := countIfCases(n)

	caseNum := 0
	for cur := n; cur != nil; {

		cur.Body = instrumentStmtList(cur.Body)
		cur.Body = addControllRec(
			cur.Body,
			cur.Pos(),
			numCases,
			caseNum,
			"I",
			true,
		)

		caseNum++

		// else-if
		if len(cur.Else) == 1 {
			if next, ok := cur.Else[0].(*ir.IfStmt); ok {
				cur = next
				continue
			}
		}

		// final else
		if cur.Else != nil {
			cur.Else = instrumentStmtList(cur.Else)
			cur.Else = addControllRec(
				cur.Else,
				cur.Pos(),
				numCases,
				caseNum,
				"I",
				true,
			)
		}

		break
	}
}

func countIfCases(n *ir.IfStmt) int {
	count := 1

	for {
		if len(n.Else) == 1 {
			if next, ok := n.Else[0].(*ir.IfStmt); ok {
				count++
				n = next
				continue
			}
		}

		if n.Else != nil {
			count++
		}

		return count
	}
}

// ==================================================
// MARK: Switch
// ==================================================

func instrumentSwitch(n *ir.SwitchStmt) {
	if !base.Flag.AdvocateRecordPlus {
		return
	}

	numCases := len(n.Cases)

	for i, c := range n.Cases {
		c.Body = instrumentStmtList(c.Body)

		c.Body = addControllRec(
			c.Body,
			c.Pos(),
			numCases,
			i,
			"S",
			true,
		)
	}
}
func countSwitchCases(n *ir.SwitchStmt) int {
	count := 0

	for _, c := range n.Compiled {
		if _, ok := c.(*ir.CaseClause); ok {
			count++
		}
	}

	return count
}

// ==================================================
// MARK: Loop
// ==================================================

func instrumentLoop(body ir.Nodes, pos src.XPos) ir.Nodes {
	if !base.Flag.AdvocateRecordPlus {
		return body
	}

	p := base.Ctxt.PosTable.Pos(pos)
	if advocateIgnore(p.Filename()) {
		return body
	}

	fn := typecheck.LookupRuntime("advocateControllFlow")

	call := typecheck.Call(
		pos,
		fn,
		[]ir.Node{
			ir.NewString(pos, "L"),
			ir.NewInt(pos, int64(0)),
			ir.NewInt(pos, int64(0)),
		},
		false,
	)

	out := make(ir.Nodes, 0, len(body)+1)
	out.Append(call)
	out.Append(body...)

	return out
}

// ==================================================
// MARK: Parameter
// ==================================================

func instrumentParameterCopy(fn *ir.Func) {
	calls := make(ir.Nodes, 0)
	for _, n := range fn.Dcl {
		if n.Class != ir.PPARAM {
			continue
		}

		// Skip *sync.Mutex.
		if n.Type().IsPtr() {
			continue
		}

		var runtimeName string

		switch {
		case isSyncType(n.Type(), "Mutex") || isSyncType(n.Type(), "RWMutex"):
			runtimeName = "AdvocateAllocMutex"

		case isSyncType(n.Type(), "Cond"):
			runtimeName = "AdvocateAllocCondVar"

		case isSyncType(n.Type(), "WaitGroup"):
			runtimeName = "AdvocateAllocWG"

		case isSyncType(n.Type(), "Once"):
			runtimeName = "AdvocateAllocOnce"

		default:
			continue
		}

		f := typecheck.LookupRuntime(runtimeName)

		addr := ir.NewAddrExpr(n.Pos(), n)
		addr.SetType(types.NewPtr(n.Type()))
		addr.SetTypecheck(1)

		call := typecheck.Call(
			n.Pos(),
			f,
			[]ir.Node{makeUnsafePointer(addr, n.Pos())},
			false,
		)

		calls = append(calls, call)
	}

	if len(calls) > 0 {
		body := append(calls, fn.Body...)
		fn.Body = body
	}
}

// ==================================================
// MARK: shouldAdvocate
// ==================================================

func shouldAdvocate(n ir.Node) bool {
	if n == nil {
		return false
	}

	if !(base.Flag.AdvocateTrace || base.Flag.AdvocateReplay || base.Flag.AdvocateFuzzing) {
		return false
	}

	pkgName := n.Sym().Pkg.Path
	if pkgName == "runtime" ||
		pkgName == "syscall" ||
		pkgName == "os" ||
		pkgName == "fmt" ||
		strings.HasPrefix(pkgName, "internal") {

		return false
	}

	return !isAdvocateCall(n)
}

func isAdvocateCall(n ir.Node) bool {
	call, ok := n.(*ir.CallExpr)
	if !ok {
		return false
	}

	if call.Op() != ir.OCALLFUNC {
		return false
	}

	name, ok := call.Fun.(*ir.Name)
	if !ok {
		return false
	}

	return name.Sym() != nil &&
		(fmt.Sprint(name.Sym()) == "AdvocateAllocMutex" ||
			fmt.Sprint(name.Sym()) == "AdvocateAllocCondVar" ||
			fmt.Sprint(name.Sym()) == "AdvocateAllocWG" ||
			fmt.Sprint(name.Sym()) == "AdvocateAllocOnce" ||
			fmt.Sprint(name.Sym()) == "advocateTraceControllFlow")
}

// ==================================================
// MARK: helper
// ==================================================

func printFunc(fn *ir.Func) {
	fmt.Printf("FUNC: %v\n", fn.Sym())
	ir.DumpList("body", fn.Body)
}

func isUserMain(fn *ir.Func) bool {
	if fn == nil {
		return false
	}

	sym := fn.Sym()

	if sym.Name != "main" {
		return false
	}

	pkg := sym.Pkg
	if pkg == nil {
		return false
	}

	return pkg.Name == "main" && pkg.Path == "main"
}

func advocateIgnore(file string) bool {
	return (strings.Contains(file, "goPatch/src/") || strings.Contains(file, "go/pkg/mod")) &&
		!strings.Contains(file, "goPatch/src/time/tick.go") &&
		!strings.Contains(file, "goPatch/src/context/context.go")
}
