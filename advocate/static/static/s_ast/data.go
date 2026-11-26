// Copyright (c) 2026 Erik Kassubek
//
// File: data.go
// Brief: Data for the ast analysis
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ast

import (
	"advocate/static/static/s_base"
	"advocate/utils/log"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"runtime"

	"golang.org/x/tools/go/packages"
)

type objId int

type FuncCall struct {
	Call     *ast.CallExpr
	Decl     *ast.FuncDecl
	Name     string
	CallType s_base.FuncName
}

type Operation struct {
	Obj objId
	Op  s_base.FuncName
}

type FuncInfo struct {
	decl *ast.FuncDecl

	// functions called in each function,
	// func -> call
	// ast.Expr.(type) -> *ast.Ident: direct function (foo())
	// ast.Expr.(type) -> *ast.SelectorExpr: methodCall (obj.Method())
	// ast.Expr.(type) -> *ast.FuncLit: function literal (func() {...}())
	FuncCalls map[FuncCall]struct{}

	// routine spawns from functions
	// *ast.GoStmt.Call.(type) -> *ast.Ident: direct function (go foo())
	// *ast.GoStmt.Call.(type) -> *ast.SelectorExpr: methodCall (go obj.Method())
	// *ast.GoStmt.Call.(type) -> *ast.FuncLit: function literal (go func() { ... }())
	GoCalls map[*ast.GoStmt]*ast.FuncDecl

	Ops map[Operation]map[ast.Expr]struct{}
}

type Data struct {
	dir string // path to analyzed program

	Pkgs []*packages.Package

	Fset *token.FileSet

	pkgInfo map[*packages.Package]*types.Info
	uses    map[*ast.Ident]types.Object
	defs    map[*ast.Ident]types.Object

	AstMap map[string][]*ast.File         // pkg path -> files
	ast    []*ast.File                    // flattened list
	npm    map[ast.Node]*packages.Package // node packages map

	FuncDeclMap map[token.Pos]*ast.FuncDecl
	FuncInfo    map[*ast.FuncDecl]FuncInfo
	RoutFunc    map[*ast.GoStmt]*ast.FuncDecl
	FuncLitDecl map[*ast.FuncLit]*ast.FuncDecl // dummy for func lit

	IfStmtMap map[token.Pos]*ast.IfStmt

	nextFuncLitId int
}

func BuildAst(dir string) (*Data, error) {
	data := &Data{
		dir:  dir,
		Fset: token.NewFileSet(),

		AstMap: make(map[string][]*ast.File),
		ast:    make([]*ast.File, 0),
		npm:    make(map[ast.Node]*packages.Package),

		FuncDeclMap: make(map[token.Pos]*ast.FuncDecl),
		FuncInfo:    make(map[*ast.FuncDecl]FuncInfo),
		RoutFunc:    make(map[*ast.GoStmt]*ast.FuncDecl),

		FuncLitDecl: make(map[*ast.FuncLit]*ast.FuncDecl),

		IfStmtMap: make(map[token.Pos]*ast.IfStmt),
	}

	err := data.loadPackages()
	if err != nil {
		log.Error(err.Error())
		return data, err
	}

	data.buildTypeInfo()

	data.buildAst()

	data.CollectOperations()

	return data, nil
}

// Determine the packages and type info
//
// Parameter:
//   - dir: string: root directory of project
func (self *Data) loadPackages() error {
	cfg := &packages.Config{
		Fset: self.Fset,
		Mode: packages.LoadAllSyntax,
		Dir:  self.dir,
		Env: append(os.Environ(),
			"GO111MODULE=on",
			"GOROOT="+runtime.GOROOT(),
		),
		Tests: true,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("static analysis: failed to load packages: %w", err)
	}

	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			return fmt.Errorf("static analysis: packages contain errors: %s", err.Error())
		}
	}

	self.Pkgs = pkgs
	return nil
}

func (self *Data) buildTypeInfo() {
	self.pkgInfo = make(map[*packages.Package]*types.Info)

	self.uses = make(map[*ast.Ident]types.Object)
	self.defs = make(map[*ast.Ident]types.Object)

	for _, pkg := range self.Pkgs {
		if pkg.TypesInfo == nil {
			continue
		}

		self.pkgInfo[pkg] = pkg.TypesInfo

		for ident, obj := range pkg.TypesInfo.Uses {
			self.uses[ident] = obj
		}

		for ident, obj := range pkg.TypesInfo.Defs {
			self.defs[ident] = obj
		}
	}
}

func (self *Data) PrintInfo() {
	fmt.Print("=================== Info ===================\n\n")
	for p, c := range self.FuncInfo {
		pos := self.getPos(p)
		if pos == "[internal]" {
			continue
		}
		fmt.Println(self.getName(p.Name), pos)

		fmt.Println("\tFuncs: ")
		for call := range self.FuncInfo[p].FuncCalls {
			fmt.Println("\t\t", call.Name, self.getPos(call.Call), self.getPos(call.Decl))
		}

		fmt.Println("\tGo: ")
		for ch, call := range self.FuncInfo[p].GoCalls {
			fmt.Println("\t\t", call.Name, self.getPos(ch), self.getPos(call))
		}

		fmt.Println("\tOps: ")
		for op, expr := range c.Ops {
			for e := range expr {
				fmt.Println("\t\t", op.Obj, op.Op, self.getPos(e))
			}
		}

		fmt.Println("")
	}

}
