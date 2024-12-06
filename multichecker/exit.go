// Package main implements a static analysis tool that identifies
// direct calls to os.Exit within the main function of the main package.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer defines a static analysis tool that flags direct calls
// to os.Exit in the main function of the main package.
var Analyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "checks for direct calls to os.Exit in main",
	Run:  run,
}

// run executes the analysis logic. It iterates through all files
// in the package, identifies the main function in the main package,
// and inspects its body for direct calls to os.Exit.
//
// If a direct call to os.Exit is found, a diagnostic message is reported.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Check if the file belongs to the main package
		if pass.Pkg.Name() != "main" {
			continue
		}

		// Iterate over all declarations in the file
		for _, decl := range file.Decls {
			// Look for the main function
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" || funcDecl.Recv != nil {
				continue
			}

			// Inspect the body of the main function
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check for calls to os.Exit
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "os" && selector.Sel.Name == "Exit" {
						pass.Reportf(call.Pos(), "direct call to os.Exit is not allowed in main.main")
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
