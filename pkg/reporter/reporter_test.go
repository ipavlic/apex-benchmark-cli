package reporter

import (
	"bytes"
	"fmt"
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

func TestPrintJSON_NilWriter(t *testing.T) {
	result := types.AggregatedResult{
		Name:     "Test",
		AvgCpuMs: 1.0,
	}

	// Should not panic with nil writer (defaults to stdout)
	// We can't easily test stdout output, but at least verify no error
	err := PrintJSON(result, nil)
	if err != nil {
		t.Errorf("Expected no error with nil writer, got: %v", err)
	}
}

func TestPrintTable_NilWriter(t *testing.T) {
	result := types.AggregatedResult{
		Name:     "Test",
		AvgCpuMs: 1.0,
	}

	err := PrintTable(result, nil)
	if err != nil {
		t.Errorf("Expected no error with nil writer, got: %v", err)
	}
}

func TestPrintComparison_NilWriter(t *testing.T) {
	results := []types.AggregatedResult{
		{Name: "A", AvgCpuMs: 1.0},
		{Name: "B", AvgCpuMs: 2.0},
	}

	err := PrintComparison(results, nil)
	if err != nil {
		t.Errorf("Expected no error with nil writer, got: %v", err)
	}
}

func TestPrintTable_WithAllFields(t *testing.T) {
	result := types.AggregatedResult{
		Name:         "TestWithAllFields",
		AvgCpuMs:     1.234,
		MinCpuMs:     1.100,
		MaxCpuMs:     1.400,
		StdDevCpuMs:  0.123,
		AvgWallMs:    1.5,
		MinWallMs:    1.3,
		MaxWallMs:    1.7,
		StdDevWallMs: 0.2,
		Runs:         10,
		Iterations:   100,
		Warmup:       10,
	}

	var buf bytes.Buffer
	err := PrintTable(result, &buf)
	if err != nil {
		t.Fatalf("PrintTable failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "TestWithAllFields") {
		t.Errorf("Table output missing name: %s", output)
	}
}

func TestPrintJSON_SliceOfResults(t *testing.T) {
	results := []types.AggregatedResult{
		{Name: "Test1", AvgCpuMs: 1.0},
		{Name: "Test2", AvgCpuMs: 2.0},
	}

	var buf bytes.Buffer
	err := PrintJSON(results, &buf)
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test1") || !strings.Contains(output, "Test2") {
		t.Errorf("JSON output missing expected data: %s", output)
	}
}

func TestPrintComparison_SingleResult(t *testing.T) {
	results := []types.AggregatedResult{
		{Name: "OnlyOne", AvgCpuMs: 1.0, MinCpuMs: 0.9, MaxCpuMs: 1.1},
	}

	var buf bytes.Buffer
	err := PrintComparison(results, &buf)
	if err != nil {
		t.Fatalf("PrintComparison should work with single result: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "OnlyOne") {
		t.Error("Output should contain the single result name")
	}
	if !strings.Contains(output, "1.00x") {
		t.Error("Single result should be marked as fastest with 1.00x")
	}
}

// errorWriter is a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write error")
}

func TestPrintJSON_WriteError(t *testing.T) {
	result := types.AggregatedResult{
		Name:     "Test",
		AvgCpuMs: 1.0,
	}

	writer := &errorWriter{}
	err := PrintJSON(result, writer)

	if err == nil {
		t.Error("Expected error when writer fails")
	}
}

func TestPrintTable_LargeValues(t *testing.T) {
	// Test with very large values
	result := types.AggregatedResult{
		Name:         "LargeValues",
		AvgCpuMs:     9999.999,
		MinCpuMs:     9000.000,
		MaxCpuMs:     10999.999,
		StdDevCpuMs:  1500.500,
		AvgWallMs:    10000.000,
		MinWallMs:    9500.000,
		MaxWallMs:    11000.000,
		StdDevWallMs: 1600.000,
	}

	var buf bytes.Buffer
	err := PrintTable(result, &buf)
	if err != nil {
		t.Fatalf("PrintTable failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "LargeValues") {
		t.Error("Table should contain name")
	}
	if !strings.Contains(output, "9999.999") {
		t.Error("Table should contain large avg CPU value")
	}
}

func TestPrintComparison_MultipleResults(t *testing.T) {
	// Test with many results
	results := []types.AggregatedResult{
		{Name: "Test1", AvgCpuMs: 1.0, MinCpuMs: 0.9, MaxCpuMs: 1.1},
		{Name: "Test2", AvgCpuMs: 2.0, MinCpuMs: 1.9, MaxCpuMs: 2.1},
		{Name: "Test3", AvgCpuMs: 1.5, MinCpuMs: 1.4, MaxCpuMs: 1.6},
		{Name: "Test4", AvgCpuMs: 3.0, MinCpuMs: 2.8, MaxCpuMs: 3.2},
	}

	var buf bytes.Buffer
	err := PrintComparison(results, &buf)
	if err != nil {
		t.Fatalf("PrintComparison failed: %v", err)
	}

	output := buf.String()
	// Check all results are present
	for _, r := range results {
		if !strings.Contains(output, r.Name) {
			t.Errorf("Output missing result: %s", r.Name)
		}
	}

	// Fastest should be Test1 (1.0)
	if !strings.Contains(output, "Fastest: Test1") {
		t.Error("Should identify Test1 as fastest")
	}
}
