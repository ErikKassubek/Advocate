// Copyright (c) 2024 
//
// File: readProg.go
// Brief: Functions to read in a program an extract all relevant operations
//
//
// License: BSD-3-Clause

package complete

import (
	"gosc/utils/log"
	"gosc/utils/paths"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

// Pass over all go files in a program and get the positions of all relevant
// primitives and operation on those primitives
//
// Parameter:
//   - progPath string: path to the project
//
// Returns:
//   - map[string][]int: all lines in the code that contain relevant operations.
//   - The map has the form filePath -> list of lines in this file
//   - error
func getProgramElements(progPath string) (map[string][]int, error) {
	progElems := make(map[string][]int)

	files, err := collectGoFiles(progPath)
	if err != nil {
		log.Error("Error in collecting files")
		return nil, err
	}

	pkg, err := analyzeFiles(files)
	if err != nil {
		log.Error("Error in analyzing files")
		return nil, err
	}

	// traverse all .go files in the directory recursively
	err = filepath.Walk(progPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("Error in walking")
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			content, err := os.ReadFile(path)
			if err != nil {
				log.Error("Error in reading file")
				return err
			}

			elems, err := getElemsFromContent(path, string(content), pkg)

			if err == nil && len(elems) > 0 {
				if _, ok := progElems[path]; !ok {
					progElems[path] = make([]int, 0)
				}
				for _, elem := range elems {
					progElems[path] = append(progElems[path], elem)
				}
			}
		}
		return nil
	})

	return progElems, err
}

// Given a directory, recursively collect all go files
//
// Parameter:
//   - dir string: path to the directory
//
// Returns:
//   - []string: paths to all go file in dir
//   - error
func collectGoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("Error when collecting %q: %v\n", path, err)
			return err
		}
		if info == nil {
			return nil
		}

		if info.IsDir() && info.Name() == paths.NameResult {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// analyzeFiles type-checks a package and returns the resulting package object
//
// Parameter:
//   - files []string: list of files to analyze
//
// Returns:
//   - *types.Package: the resulting types package
//   - error
func analyzeFiles(files []string) (*types.Package, error) {
	fset := token.NewFileSet()
	var astFiles []*ast.File

	for _, file := range files {
		parsedFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		astFiles = append(astFiles, parsedFile)
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("mypackage", fset, astFiles, &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	})

	if err != nil {
		log.Error(err.Error())
	}

	return pkg, nil
}

// getElemsFromContent passes the ast tree and for a given file, finds the
// line numbers of all relevant elements
//
// Parameter:
//   - path string: the path to the analyzed file
//   - content string: the content of the file
//   - pkg *types.Package: The type information for the ast
//
// Returns:
//   - []int: the line numbers containing relevant operations
//   - error
func getElemsFromContent(path string, content string, pkg *types.Package) ([]int, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return []int{}, err
	}

	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
	}

	imports := pkg.Imports()

	var syncPkg *types.Package
	for _, imp := range imports {
		if imp.Name() == "sync" {
			syncPkg = imp
			break
		}
	}

	v := &visitor{fset: fset, pkg: pkg, info: info, syncPkg: syncPkg,
		selectCases: make(map[string]struct{}), elements: make([]int, 0)}
	ast.Walk(v, node)

	return v.elements, nil
}

// visitor ist eine Struktur, die das ast.Visitor Interface implementiert.
// Sie wird verwendet, um den AST zu durchlaufen.
type visitor struct {
	fset        *token.FileSet
	pkg         *types.Package
	info        *types.Info
	syncPkg     *types.Package
	selectCases map[string]struct{}
	elements    []int // line numbers
}

// Visit is called for each node when passing the ast. It determines if the
// node represents a relevant operations and if so records the line number
// of the element
//
// Parameter:
//   - n ast.Node: the currently visited node
//
// Returns:
//   - ast.Visitor
func (this *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch x := n.(type) {
	case *ast.GoStmt:
		this.recordElement(this.fset.Position(n.Pos()))
	case *ast.SendStmt: // send
		if _, ok := this.selectCases[this.fset.Position(n.Pos()).String()]; ok {
			delete(this.selectCases, this.fset.Position(n.Pos()).String())
		} else {
			this.recordElement(this.fset.Position(n.Pos()))
		}
	case *ast.UnaryExpr: // recv
		if x.Op == token.ARROW {
			if _, ok := this.selectCases[this.fset.Position(n.Pos()).String()]; ok {
				delete(this.selectCases, this.fset.Position(n.Pos()).String())
			} else {
				this.recordElement(this.fset.Position(n.Pos()))
			}
		}
	case *ast.CallExpr:
		// close
		if fun, ok := x.Fun.(*ast.Ident); ok && fun.Name == "close" {
			this.recordElement(this.fset.Position(n.Pos()))
		}
		if fun, ok := x.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := fun.X.(*ast.Ident); ok {
				obj := this.info.Uses[ident]

				if obj != nil && this.syncPkg != nil {
					typ := obj.Type()

					// Überprüfen Sie, ob der Typ zu einem der spezifischen Typen gehört
					mutexType := this.syncPkg.Scope().Lookup("Mutex").Type()
					rwMutexType := this.syncPkg.Scope().Lookup("RWMutex").Type()
					wgType := this.syncPkg.Scope().Lookup("WaitGroup").Type()
					condType := this.syncPkg.Scope().Lookup("Cond").Type()
					onceType := this.syncPkg.Scope().Lookup("Once").Type()

					switch {
					case types.AssignableTo(typ, mutexType):
						this.recordElement(this.fset.Position(n.Pos()))
					case types.AssignableTo(typ, rwMutexType):
						this.recordElement(this.fset.Position(n.Pos()))
					case types.AssignableTo(typ, wgType):
						this.recordElement(this.fset.Position(n.Pos()))
					case types.AssignableTo(typ, condType):
						this.recordElement(this.fset.Position(n.Pos()))
					case types.AssignableTo(typ, onceType):
						this.recordElement(this.fset.Position(n.Pos()))
					}
				}
			}
		}
	case *ast.SelectStmt:
		this.recordElement(this.fset.Position(n.Pos()))
		for _, stmt := range x.Body.List {
			caseClause, ok := stmt.(*ast.CommClause)
			if !ok {
				continue // Nicht ein case-timeouteil, weitermachen
			}
			switch comm := caseClause.Comm.(type) {
			case *ast.SendStmt:
				// store to not record the send statement
				this.selectCases[this.fset.Position(comm.Pos()).String()] = struct{}{}
			case *ast.ExprStmt:
				// store to not record the recv statement
				if unaryExpr, ok := comm.X.(*ast.UnaryExpr); ok && unaryExpr.Op == token.ARROW {
					this.selectCases[this.fset.Position(unaryExpr.Pos()).String()] = struct{}{}
				}
			}
		}
	case *ast.RangeStmt:
		rangeExpr := x.X
		rangeExprType := this.info.Types[rangeExpr].Type
		// Check if the range expression is a channel
		if _, ok := rangeExprType.(*types.Chan); ok {
			fmt.Printf("Range über Kanal gefunden bei %s\n", this.fset.Position(n.Pos()))
		}
	}

	return this
}

// recordElement stores the line of a node in the visitor
//
// Parameter:
//   - pos (token.Position): the code position of the node for which the
//     function is called
func (this *visitor) recordElement(pos token.Position) {
	this.elements = append(this.elements, pos.Line)
}
