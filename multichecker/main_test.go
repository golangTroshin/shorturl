package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestGetAnalyzers(t *testing.T) {
	analyzers := getAnalyzers()

	// Ensure the list of analyzers is not empty
	assert.NotEmpty(t, analyzers, "Expected analyzers list to be non-empty")

	// Check for specific analyzers
	expectedAnalyzerNames := []string{
		"asmdecl",
		"assign",
		"atomic",
		"bools",
		"buildtag",
		"cgocall",
		"composites",
		"copylocks",
		"SA1000",
		"SA1001",
		"SA1002",
		"SA1003",
		"SA1004",
		"SA1005",
		"SA1006",
		"SA1007",
		"SA1008",
		"SA1010",
		"SA1011",
		"SA1012",
		"SA1013",
		"SA1014",
		"SA1015",
		"SA1016",
		"SA1017",
		"SA1018",
		"SA1019",
		"SA1020",
		"SA1021",
		"SA1023",
		"SA1024",
		"SA1025",
		"SA1026",
		"SA1027",
		"SA1028",
		"SA1029",
		"SA1030",
		"SA1031",
		"SA1032",
		"SA2000",
		"SA2001",
		"SA2002",
		"SA2003",
		"SA3000",
		"SA3001",
		"SA4000",
		"SA4001",
		"SA4003",
		"SA4004",
		"SA4005",
		"SA4006",
		"SA4008",
		"SA4009",
		"SA4010",
		"SA4011",
		"SA4012",
		"SA4013",
		"SA4014",
		"SA4015",
		"SA4016",
		"SA4017",
		"SA4018",
		"SA4019",
		"SA4020",
		"SA4021",
		"SA4022",
		"SA4023",
		"SA4024",
		"SA4025",
		"SA4026",
		"SA4027",
		"SA4028",
		"SA4029",
		"SA4030",
		"SA4031",
		"SA4032",
		"SA5000",
		"SA5001",
		"SA5002",
		"SA5003",
		"SA5004",
		"SA5005",
		"SA5007",
		"SA5008",
		"SA5009",
		"SA5010",
		"SA5011",
		"SA5012",
		"SA6000",
		"SA6001",
		"SA6002",
		"SA6003",
		"SA6005",
		"SA6006",
		"SA9001",
		"SA9002",
		"SA9003",
		"SA9004",
		"SA9005",
		"SA9006",
		"SA9007",
		"SA9008",
		"SA9009",
		"S1000",
		"ST1000",
		"exit",
	}

	foundAnalyzers := make(map[string]bool)
	for _, analyzer := range analyzers {
		foundAnalyzers[analyzer.Name] = true
	}

	for _, expectedName := range expectedAnalyzerNames {
		assert.True(t, foundAnalyzers[expectedName], "Expected analyzer %s to be present", expectedName)
	}
}

func TestMultichecker(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}