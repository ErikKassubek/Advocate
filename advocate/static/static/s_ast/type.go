// Copyright (c) 2026 Erik Kassubek
//
// File: type.go
// Brief: Build type info
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ast

import (
	"advocate/static/static/s_base"
	"go/ast"
	"go/token"
	"go/types"
)

func (self *Data) getName(id ast.Expr) string {
	if id == nil {
		return "NIL"
	}

	return self.getPackage(id) + ":" + self.getNameRec(id)
}

func (self *Data) getNameRec(id ast.Expr) string {
	switch e := id.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return self.getNameRec(e.X) + "." + e.Sel.Name
	case *ast.FuncLit:
		return "FuncLit"
	case *ast.CallExpr:
		return self.getNameRec(e.Fun)
	default:
		return "Other" // TODO: write for others
		// panic(fmt.Sprintf("Unknown expr type %T", e))
	}
}

func (self *Data) getPackage(expr ast.Expr) string {
	pkg, ok := self.npm[expr]
	if !ok {
		return "[unknown]"
	}
	return pkg.Name
}

// parse the files the determine the type information
// TODO: this takes very long, maybe only apply to the files/packages where it is necessary
func (self *Data) CollectOperations() {
	for _, file := range self.ast {
		self.detOpsInFile(file)
	}
}

func (self *Data) detOpsInPackage(pkgName string) {
	for _, file := range self.ast {
		if file.Name.Name != pkgName {
			continue
		}

		self.detOpsInFile(file)
	}
}

func (self *Data) detOpsInPAth(path string) {
	for _, file := range self.ast {
		filePath := self.Fset.Position(file.Pos()).Filename
		if filePath != path {
			continue
		}

		self.detOpsInFile(file)
	}
}

func (self *Data) detOpsInFile(file *ast.File) {
	if file == nil {
		return
	}

	for _, decl := range file.Decls {
		fdecl, ok := decl.(*ast.FuncDecl)
		if !ok || fdecl.Body == nil {
			continue
		}

		self.detOpsInFunc(fdecl)
	}
}

func (self *Data) detOpsInFunc(fdecl *ast.FuncDecl) {
	if fdecl == nil || fdecl.Body == nil {
		return
	}

	self.addFuncIfNotExists(fdecl)

	self.FuncDeclMap[fdecl.Name.Pos()] = fdecl

	ast.Inspect(fdecl.Body, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch x := n.(type) {
		case *ast.IfStmt:
			self.IfStmtMap[x.Pos()] = x

		case *ast.GoStmt:
			self.recordGoStatement(fdecl, x)

		case *ast.SendStmt:
			self.recordOperation(fdecl, x.Chan, s_base.ChanSend)

		case *ast.UnaryExpr:
			if x.Op == token.ARROW {
				self.recordOperation(fdecl, x.X, s_base.ChanRecv)
			}

		case *ast.CallExpr:
			ft := s_base.UnknownFunc

			if ident, ok := x.Fun.(*ast.Ident); ok {
				switch ident.Name {
				case "close":
					if len(x.Args) == 1 && self.isChannel(x.Args[0]) {
						ft = s_base.ChanClose
					}

				case "make":
					if len(x.Args) > 0 {
						if _, ok := x.Args[0].(*ast.ChanType); ok {
							ft = s_base.MakeChan
						}
					}
				}

				if ft != s_base.UnknownFunc {
					self.recordOperation(fdecl, x.Args[0], ft)
					return true
				}
			}

			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if self.isMutex(x) {
					switch sel.Sel.Name {
					case "Lock":
						ft = s_base.MutexLock
					case "TryLock":
						ft = s_base.MutexTryLock
					case "RLock":
						ft = s_base.MutexRLock
					case "TryRLock":
						ft = s_base.MutexTryRLock
					case "Unlock":
						ft = s_base.MutexUnlock
					case "RUnlock":
						ft = s_base.MutexRUnlock
					}

				} else if self.isCondVar(x) {
					switch sel.Sel.Name {
					case "Wait":
						ft = s_base.CondWait
					case "Signal":
						ft = s_base.CondSignal
					case "Broadcast":
						ft = s_base.CondBroadcast
					}

				} else if self.isWaitGroup(x) {
					switch sel.Sel.Name {
					case "Wait":
						ft = s_base.WgWait
					case "Add":
						ft = s_base.WgAdd
					case "Done":
						ft = s_base.WgDone
					case "Go":
						ft = s_base.WgGo
					}
				}

				if ft != s_base.UnknownFunc {
					self.recordOperation(fdecl, sel.X, ft)
					return true
				}
			}

			self.recordFunctionCall(fdecl, x)
		}

		return true
	})
}

func (self *Data) isConcOp(id *ast.CallExpr) string {
	if self.isMutex(id) {
		return "Mutex"
	} else if self.isCondVar(id) {
		return "CondVar"
	} else if self.isWaitGroup(id) {
		return "WaitGroup"
	} else {
		return "?"
	}
}

func (self *Data) isMutex(id *ast.CallExpr) bool {
	return self.isConcObj(id, s_base.Mutex)
}

func (self *Data) isCondVar(id *ast.CallExpr) bool {
	return self.isConcObj(id, s_base.CondVar)
}

func (self *Data) isWaitGroup(id *ast.CallExpr) bool {
	return self.isConcObj(id, s_base.Wg)
}

func (self *Data) isConcObj(call *ast.CallExpr, on s_base.ObjName) bool {
	if call == nil {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg := self.npm[call]
	if pkg == nil {
		return false
	}

	info := self.pkgInfo[pkg]
	if info == nil {
		return false
	}

	selection := info.Selections[sel]
	if selection == nil {
		return false
	}

	t := selection.Recv()

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	switch on {
	case s_base.Mutex:
		return (obj.Name() == "Mutex" || obj.Name() == "RWMutex") &&
			obj.Pkg().Path() == "sync"

	case s_base.Wg:
		return obj.Name() == "WaitGroup" &&
			obj.Pkg().Path() == "sync"

	case s_base.CondVar:
		return obj.Name() == "Cond" &&
			obj.Pkg().Path() == "sync"
	}

	return false
}

func (self *Data) isChannel(expr ast.Expr) bool {
	// Find the package containing this node.
	pkg := self.npm[expr]
	if pkg == nil {
		return false
	}

	info := self.pkgInfo[pkg]
	if info == nil {
		return false
	}

	tv, ok := info.Types[expr]
	if !ok || tv.Type == nil {
		return false
	}

	_, ok = tv.Type.Underlying().(*types.Chan)
	return ok
}

func (self *Data) getConcFuncName(id ast.Expr) s_base.FuncName {
	// if x, ok := id.(*ast.SelectorExpr); ok { // X.Sel is func name
	// 	// fmt.Println(self.isMutex(x.Fun))
	// 	// } else {
	// 	// 	fmt.Println("Not Ident")
	// 	fmt.Println("is call ", x.Sel, self.isMutex(x.Sel))
	// } else {
	// 	fmt.Println("is not")
	// }
	switch x := id.(type) {
	// case *ast.SendStmt: // channel send
	// 	return "<chan:send>"
	case *ast.UnaryExpr: // channel recv
		if x.Op == token.ARROW {
			return s_base.ChanRecv
		}
	case *ast.CallExpr:
		// channel close
		if ident, ok := x.Fun.(*ast.Ident); ok {
			switch ident.Name {
			case "close":
				if len(x.Args) != 1 {
					return s_base.UnknownFunc
				}
				if self.isChannel(x.Args[0]) {
					return s_base.ChanClose
				}
				return s_base.UnknownFunc
			case "make":
				if len(x.Args) == 0 {
					return s_base.UnknownFunc
				}

				switch x.Args[0].(type) {
				case *ast.ChanType:
					return s_base.MakeChan
				}
				return s_base.MakeFunc
			}
		}

		if sel, ok := x.Fun.(*ast.SelectorExpr); ok { // a.x()
			if self.isMutex(x) {
				switch sel.Sel.Name {
				case "Lock":
					return s_base.MutexLock
				case "TryLock":
					return s_base.MutexTryLock
				case "RLock":
					return s_base.MutexRLock
				case "TryRLock":
					return s_base.MutexTryRLock
				case "Unlock":
					return s_base.MutexUnlock
				case "RUnlock":
					return s_base.MutexRUnlock
				}
			} else if self.isCondVar(x) {
				switch sel.Sel.Name {
				case "Wait":
					return s_base.CondWait
				case "Signal":
					return s_base.CondSignal
				case "Broadcast":
					return s_base.CondBroadcast
				}
			} else if self.isWaitGroup(x) {
				switch sel.Sel.Name {
				case "Wait":
					return s_base.WgWait
				case "Add":
					return s_base.WgAdd
				case "Done":
					return s_base.WgDone
				case "Go":
					return s_base.WgGo
				}
			}
		}
	}

	return s_base.UnknownFunc
}
