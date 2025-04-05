package osexit

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "checks for os.Exit call from main",
	Run:  run,
}

var osPackageName = "os"

func run(pass *analysis.Pass) (interface{}, error) {
	osPackageDecl := fmt.Sprintf(`"%s"`, osPackageName)
	for _, f := range pass.Files {
		currentOsPackageName := osPackageName
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncDecl:
				if n.Name.Name == "main" {
					return true
				}
			case *ast.File:
				if n.Name.Name == "main" {
					return true
				}
			case *ast.GenDecl:
				return true
			case *ast.ImportSpec:
				if n.Path.Value == osPackageDecl && n.Name != nil {
					currentOsPackageName = n.Name.Name
				}
			case *ast.BlockStmt:
				return true
			case *ast.ExprStmt:
				if call, ok := n.X.(*ast.CallExpr); ok {
					if s, ok := call.Fun.(*ast.SelectorExpr); ok {
						if i, ok := s.X.(*ast.Ident); ok {
							if i.Name == currentOsPackageName && s.Sel.Name == "Exit" {
								pass.Reportf(i.NamePos, "os.Exit call from main")
							}
						}
					}
				}
			}
			return false
		})
	}
	return nil, nil
}
