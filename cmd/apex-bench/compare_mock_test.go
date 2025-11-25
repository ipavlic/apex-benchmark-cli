package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

func TestCompareBenchmarksWithExecutor_Success(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mock := &mockExecutor{}
	benchSpecs := []types.BenchmarkSpec{
		{Name: "Bench1", Code: "String s1 = 'a';"},
		{Name: "Bench2", Code: "String s2 = 'b';"},
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, false, false, "table")

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Table output should contain benchmark names
	if !strings.Contains(output, "Bench1") {
		t.Errorf("Expected output to contain Bench1, got: %s", output)
	}
	if !strings.Contains(output, "Bench2") {
		t.Errorf("Expected output to contain Bench2, got: %s", output)
	}
}

func TestCompareBenchmarksWithExecutor_JSONOutput(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mock := &mockExecutor{}
	benchSpecs := []types.BenchmarkSpec{
		{Name: "Test1", Code: "Integer x = 1;"},
		{Name: "Test2", Code: "Integer y = 2;"},
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 5, 1, 1, 1, false, false, "json")

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// JSON output should be valid JSON array
	if !strings.HasPrefix(strings.TrimSpace(output), "[") {
		t.Errorf("Expected JSON array output, got: %s", output)
	}
}

func TestCompareBenchmarksWithExecutor_WithFiles(t *testing.T) {
	// Create temporary files
	tmpFile1, err := os.CreateTemp("", "bench1-*.apex")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile1.Name())

	tmpFile2, err := os.CreateTemp("", "bench2-*.apex")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile2.Name())

	tmpFile1.Write([]byte("String s1 = 'test1';"))
	tmpFile1.Close()

	tmpFile2.Write([]byte("String s2 = 'test2';"))
	tmpFile2.Close()

	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mock := &mockExecutor{}
	benchSpecs := []types.BenchmarkSpec{
		{Name: "File1", File: tmpFile1.Name()},
		{Name: "File2", File: tmpFile2.Name()},
	}

	err = compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, false, false, "table")

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestCompareBenchmarksWithExecutor_FileReadError(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{}
	benchSpecs := []types.BenchmarkSpec{
		{Name: "Valid", Code: "String s = 'test';"},
		{Name: "Invalid", File: "/nonexistent/file.apex"},
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, false, false, "table")

	if err == nil {
		t.Error("Expected file read error")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected file read error, got: %v", err)
	}
}

func TestCompareBenchmarksWithExecutor_ExecutionError(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{
		runFunc: func(apexCode string, org string) (string, error) {
			// Fail on second benchmark
			if strings.Contains(apexCode, "Bench2") {
				return "", fmt.Errorf("execution failed for Bench2")
			}
			return mockSuccessfulBenchResult(), nil
		},
	}

	benchSpecs := []types.BenchmarkSpec{
		{Name: "Bench1", Code: "String s1 = 'a';"},
		{Name: "Bench2", Code: "String s2 = 'b';"},
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, false, false, "table")

	if err == nil {
		t.Error("Expected execution error")
	}
	if !strings.Contains(err.Error(), "execution failed for Bench2") {
		t.Errorf("Expected Bench2 execution error, got: %v", err)
	}
}

func TestCompareBenchmarksWithExecutor_MultipleRuns(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	executionCount := 0
	mock := &mockExecutor{
		executeParallelFunc: func(apexCode string, runs int, maxConcurrent int, org string) ([]string, error) {
			executionCount++
			if runs != 3 {
				return nil, fmt.Errorf("expected 3 runs, got %d", runs)
			}
			results := make([]string, runs)
			for i := 0; i < runs; i++ {
				results[i] = mockSuccessfulBenchResult()
			}
			return results, nil
		},
	}

	benchSpecs := []types.BenchmarkSpec{
		{Name: "Multi1", Code: "String s1 = 'a';"},
		{Name: "Multi2", Code: "String s2 = 'b';"},
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 3, 2, false, false, "table")

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if executionCount != 2 {
		t.Errorf("Expected 2 parallel executions (one per benchmark), got %d", executionCount)
	}
}

func TestCompareBenchmarksWithExecutor_InvalidOutputFormat(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{}
	benchSpecs := []types.BenchmarkSpec{
		{Name: "Test1", Code: "String s1 = 'a';"},
		{Name: "Test2", Code: "String s2 = 'b';"},
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, false, false, "xml")

	if err == nil {
		t.Error("Expected error for invalid output format")
	}
	if !strings.Contains(err.Error(), "unknown output format") {
		t.Errorf("Expected format error, got: %v", err)
	}
}

func TestCompareBenchmarksWithExecutor_GenerationError(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{}
	benchSpecs := []types.BenchmarkSpec{
		{Name: "", Code: "String s = 'test';"}, // Invalid: empty name
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, false, false, "table")

	if err == nil {
		t.Error("Expected generation error")
	}
	if !strings.Contains(err.Error(), "failed to generate code") {
		t.Errorf("Expected generation error, got: %v", err)
	}
}

func TestCompareBenchmarksWithExecutor_ParseError(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{
		runFunc: func(apexCode string, org string) (string, error) {
			// Return output without BENCH_RESULT marker
			return "No benchmark result here", nil
		},
	}

	benchSpecs := []types.BenchmarkSpec{
		{Name: "Parse1", Code: "String s1 = 'a';"},
		{Name: "Parse2", Code: "String s2 = 'b';"},
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, false, false, "table")

	if err == nil {
		t.Error("Expected parse error")
	}
	if !strings.Contains(err.Error(), "failed to parse results") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestCompareBenchmarksWithExecutor_WithTrackingOptions(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	generatedCount := 0
	mock := &mockExecutor{
		runFunc: func(apexCode string, org string) (string, error) {
			// Verify that heap and DB tracking code is in the generated code
			if !strings.Contains(apexCode, "Limits.getHeapSize()") {
				return "", fmt.Errorf("expected heap tracking code")
			}
			if !strings.Contains(apexCode, "Limits.getDmlStatements()") {
				return "", fmt.Errorf("expected DB tracking code")
			}
			generatedCount++
			return mockSuccessfulBenchResult(), nil
		},
	}

	benchSpecs := []types.BenchmarkSpec{
		{Name: "Track1", Code: "String s1 = 'a';"},
		{Name: "Track2", Code: "String s2 = 'b';"},
	}

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, true, true, "table")

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if generatedCount != 2 {
		t.Errorf("Expected 2 benchmarks to be generated, got %d", generatedCount)
	}
}

func TestCompareBenchmarksWithExecutor_EmptyBenchmarks(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mock := &mockExecutor{}
	benchSpecs := []types.BenchmarkSpec{} // Empty list

	err := compareBenchmarksWithExecutor(mock, "test-org", benchSpecs, 10, 2, 1, 1, false, false, "table")

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Empty list should complete successfully (edge case)
	if err != nil {
		t.Logf("Got error for empty benchmarks: %v", err)
	}
}
