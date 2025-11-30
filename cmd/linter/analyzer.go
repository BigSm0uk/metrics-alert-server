package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer проверяет использование panic, os.Exit и log.Fatal вне main.
var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "проверяет использование panic, os.Exit и log.Fatal вне функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				checkCall(pass, call)
			}
			return true
		})
	}
	return nil, nil
}

func checkCall(pass *analysis.Pass, call *ast.CallExpr) {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if ident.Name == "panic" {
			if obj := pass.TypesInfo.Uses[ident]; obj != nil {
				if builtin, ok := obj.(*types.Builtin); ok && builtin.Name() == "panic" {
					pass.Reportf(call.Pos(), "использование встроенной функции panic")
				}
			}
		}
		return
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}

	obj := pass.TypesInfo.Uses[pkgIdent]
	if obj == nil {
		return
	}

	pkgName, ok := obj.(*types.PkgName)
	if !ok {
		return
	}

	if pkgName.Imported().Path() == "os" && sel.Sel.Name == "Exit" {
		if !isInMainOfMain(pass, call) {
			pass.Reportf(call.Pos(), "вызов os.Exit вне функции main пакета main")
		}
	}

	if pkgName.Imported().Path() == "log" {
		if sel.Sel.Name == "Fatal" || sel.Sel.Name == "Fatalf" || sel.Sel.Name == "Fatalln" {
			if !isInMainOfMain(pass, call) {
				pass.Reportf(call.Pos(), "вызов log.%s вне функции main пакета main", sel.Sel.Name)
			}
		}
	}
}

func isInMainOfMain(pass *analysis.Pass, n ast.Node) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}

	var inMainFunc bool
	for _, file := range pass.Files {
		var found bool

		ast.Inspect(file, func(node ast.Node) bool {
			if found {
				return false
			}

			if fn, ok := node.(*ast.FuncDecl); ok {
				if fn.Name.Name == "main" {
					if nodeContains(fn, n) {
						inMainFunc = true
						found = true
						return false
					}
				}
			}

			return true
		})

		if found {
			break
		}
	}

	return inMainFunc
}

func nodeContains(parent, child ast.Node) bool {
	var contains bool
	ast.Inspect(parent, func(n ast.Node) bool {
		if n == child {
			contains = true
			return false
		}
		return true
	})
	return contains
}
