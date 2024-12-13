package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// main initializes and runs a multichecker, combining multiple static analysis tools
// to analyze Go code for potential issues.
//
// This program includes:
// - Standard analyzers from the golang.org/x/tools/go/analysis/passes package
// - SA analyzers from staticcheck.io
// - Additional analyzers from staticcheck.io, such as Simple and Stylecheck
// - A custom analyzer that checks for direct calls to os.Exit in the main function
//
// The goal is to provide a comprehensive analysis tool to ensure code quality,
// correctness, and maintainability.
func main() {
	multichecker.Main(getAnalyzers()...)
}

func getAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer

	// Add standard analyzers from golang.org/x/tools/go/analysis/passes
	analyzers = append(analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
	)

	// Add SA analyzers from staticcheck.io
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Add additional analyzers from staticcheck.io
	analyzers = append(analyzers, simple.Analyzers[0].Analyzer, stylecheck.Analyzers[0].Analyzer)

	// Add the custom analyzer
	analyzers = append(analyzers, Analyzer)

	return analyzers
}
