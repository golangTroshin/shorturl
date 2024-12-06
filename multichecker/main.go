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
	var analyzers []*analysis.Analyzer

	// Add standard analyzers from golang.org/x/tools/go/analysis/passes
	analyzers = append(analyzers,
		asmdecl.Analyzer,   // Checks assembly declarations
		assign.Analyzer,    // Detects suspicious assignments
		atomic.Analyzer,    // Flags incorrect atomic operations
		bools.Analyzer,     // Identifies common mistakes with boolean operations
		buildtag.Analyzer,  // Validates build tags
		cgocall.Analyzer,   // Detects unsafe cgo calls
		composite.Analyzer, // Flags composite literals that use unkeyed fields
		copylock.Analyzer,  // Detects locks passed by value
	)

	// Add SA analyzers from staticcheck.io
	for _, v := range staticcheck.Analyzers {
		// Include only analyzers with names starting with "SA"
		if v.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Add additional analyzers from staticcheck.io
	// - Simple: Finds simple issues (e.g., redundant code)
	// - Stylecheck: Enforces coding style guidelines
	analyzers = append(analyzers, simple.Analyzers[0].Analyzer, stylecheck.Analyzers[0].Analyzer)

	// Add the custom analyzer (defined elsewhere in the same package)
	// This custom analyzer checks for direct calls to os.Exit in the main function.
	analyzers = append(analyzers, Analyzer)

	// Run the multichecker with the selected analyzers
	multichecker.Main(analyzers...)
}
