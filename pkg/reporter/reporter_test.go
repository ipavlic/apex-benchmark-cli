package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

func TestPrintJSON_AggregatedResult(t *testing.T) {
	result := types.AggregatedResult{
		Name:         "TestBench",
		Runs:         5,
		Iterations:   100,
		Warmup:       10,
		AvgCpuMs:     1.234,
		StdDevCpuMs:  0.123,
		MinCpuMs:     1.100,
		MaxCpuMs:     1.400,
		AvgWallMs:    1.345,
		StdDevWallMs: 0.145,
		MinWallMs:    1.200,
		MaxWallMs:    1.500,
	}

	var buf bytes.Buffer
	err := PrintJSON(result, &buf)
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}

	output := buf.String()

	// Check that JSON contains expected fields
	expectedFields := []string{
		`"name"`,
		`"TestBench"`,
		`"runs"`,
		`"iterations"`,
		`"avgCpuMs"`,
		`"stdDevCpuMs"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("JSON output missing expected field: %s\nOutput: %s", field, output)
		}
	}
}

func TestPrintJSON_Result(t *testing.T) {
	result := types.Result{
		Name:       "SingleBench",
		Iterations: 100,
		AvgWallMs:  1.5,
		AvgCpuMs:   1.2,
		MinWallMs:  1.0,
		MaxWallMs:  2.0,
		MinCpuMs:   1.0,
		MaxCpuMs:   1.5,
	}

	var buf bytes.Buffer
	err := PrintJSON(result, &buf)
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"SingleBench"`) {
		t.Errorf("JSON output missing benchmark name\nOutput: %s", output)
	}
}

func TestPrintTable(t *testing.T) {
	result := types.AggregatedResult{
		Name:        "TestBench",
		Runs:        5,
		Iterations:  100,
		AvgCpuMs:    1.234,
		StdDevCpuMs: 0.123,
		MinCpuMs:    1.100,
		MaxCpuMs:    1.400,
	}

	var buf bytes.Buffer
	err := PrintTable(result, &buf)
	if err != nil {
		t.Fatalf("PrintTable failed: %v", err)
	}

	output := buf.String()

	// Check that table contains expected values
	expectedStrings := []string{
		"TestBench",
		"1.234 ms",
		"1.100 ms",
		"1.400 ms",
		"0.123 ms",
		"NAME",        // Headers are uppercased
		"AVG CPU",     // Headers are uppercased
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Table output missing expected string: %s\nOutput: %s", expected, output)
		}
	}
}

func TestPrintComparison(t *testing.T) {
	results := []types.AggregatedResult{
		{
			Name:     "Method A",
			AvgCpuMs: 1.0,
			MinCpuMs: 0.9,
			MaxCpuMs: 1.1,
		},
		{
			Name:     "Method B",
			AvgCpuMs: 2.0,
			MinCpuMs: 1.8,
			MaxCpuMs: 2.2,
		},
		{
			Name:     "Method C",
			AvgCpuMs: 1.5,
			MinCpuMs: 1.4,
			MaxCpuMs: 1.6,
		},
	}

	var buf bytes.Buffer
	err := PrintComparison(results, &buf)
	if err != nil {
		t.Fatalf("PrintComparison failed: %v", err)
	}

	output := buf.String()

	// Check that comparison table contains all methods
	expectedStrings := []string{
		"Method A",
		"Method B",
		"Method C",
		"1.00x",
		"2.00x",
		"1.50x",
		"Fastest: Method A",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Comparison output missing expected string: %s\nOutput: %s", expected, output)
		}
	}
}

func TestPrintComparison_Empty(t *testing.T) {
	results := []types.AggregatedResult{}

	var buf bytes.Buffer
	err := PrintComparison(results, &buf)
	if err == nil {
		t.Error("Expected error for empty results")
	}
}

func TestPrintComparison_IdentifiesFastest(t *testing.T) {
	results := []types.AggregatedResult{
		{
			Name:     "Slow",
			AvgCpuMs: 5.0,
			MinCpuMs: 4.5,
			MaxCpuMs: 5.5,
		},
		{
			Name:     "Fast",
			AvgCpuMs: 1.0,
			MinCpuMs: 0.9,
			MaxCpuMs: 1.1,
		},
		{
			Name:     "Medium",
			AvgCpuMs: 3.0,
			MinCpuMs: 2.8,
			MaxCpuMs: 3.2,
		},
	}

	var buf bytes.Buffer
	err := PrintComparison(results, &buf)
	if err != nil {
		t.Fatalf("PrintComparison failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Fastest: Fast") {
		t.Errorf("Expected 'Fastest: Fast' in output, got: %s", output)
	}

	// Check that Fast has 1.00x and star
	if !strings.Contains(output, "1.00x") {
		t.Error("Expected '1.00x' for fastest method")
	}
}
