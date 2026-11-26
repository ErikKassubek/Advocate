// Copyright (c) 2026 Erik Kassubek
//
// File: analysis.go
// Brief: Read the ssa and pass into a format that is simpler to work with
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

func (this *Data) runSSAAnalysis() {
	seen := make(map[string]bool)

	this.funcs = make(map[string]*Function)

	for _, pkg := range this.ssaPkgs {
		if pkg == nil {
			continue
		}

		path := pkg.Pkg.Path()
		if seen[path] {
			continue
		}
		seen[path] = true

		var addFunc func(*ssa.Function)
		addFunc = func(fn *ssa.Function) {
			if fn == nil {
				return
			}
			f := this.analysisFunction(fn)
			this.funcs[f.Name()] = &f

			if isMain(fn) {
				this.mainFunc = &f
			} else if isInit(fn) {
				this.initFunc = &f
			}

			for _, anon := range fn.AnonFuncs {
				addFunc(anon)
			}
		}

		for _, mem := range pkg.Members {
			if fn, ok := mem.(*ssa.Function); ok {
				addFunc(fn)
			}
		}

	}

	// for _, fn := range self.funcs {
	// 	fmt.emln(fn.string())
	// 	fmt.Print("\n\n\n==================================================\n\n\n")
	// }
}
