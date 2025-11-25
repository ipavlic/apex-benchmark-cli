package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

// mockExecutor is a mock implementation of executor.Executor for testing
type mockExecutor struct {
	runFunc             func(apexCode string, org string) (string, error)
	executeParallelFunc func(apexCode string, runs int, maxConcurrent int, org string) ([]string, error)
}

func (m *mockExecutor) Run(apexCode string, org string) (string, error) {
	if m.runFunc != nil {
		return m.runFunc(apexCode, org)
	}
	return mockSuccessfulBenchResultFromCode(apexCode), nil
}

func (m *mockExecutor) ExecuteParallel(apexCode string, runs int, maxConcurrent int, org string) ([]string, error) {
	if m.executeParallelFunc != nil {
		return m.executeParallelFunc(apexCode, runs, maxConcurrent, org)
	}
	results := make([]string, runs)
	for i := 0; i < runs; i++ {
		results[i] = mockSuccessfulBenchResultFromCode(apexCode)
	}
	return results, nil
}

func mockSuccessfulBenchResult() string {
	return `USER_DEBUG|[DEBUG]
USER_DEBUG|BENCH_RESULT:{"name":"TestBench","iterations":10,"avgCpuMs":5.5,"minCpuMs":5.0,"maxCpuMs":6.0,"avgWallMs":5.5,"minWallMs":5.0,"maxWallMs":6.0}
USER_DEBUG|[DEBUG]`
}

func mockSuccessfulBenchResultFromCode(apexCode string) string {
	// Extract benchmark name from generated code
	// The generated code contains: "name":"BenchmarkName" in the JSON
	name := "TestBench"
	if strings.Contains(apexCode, `"name":"`) {
		start := strings.Index(apexCode, `"name":"`) + len(`"name":"`)
		end := strings.Index(apexCode[start:], `"`)
		if end > 0 {
			name = apexCode[start : start+end]
		}
	}
	return fmt.Sprintf(`USER_DEBUG|[DEBUG]
USER_DEBUG|BENCH_RESULT:{"name":"%s","iterations":10,"avgCpuMs":5.5,"minCpuMs":5.0,"maxCpuMs":6.0,"avgWallMs":5.5,"minWallMs":5.0,"maxWallMs":6.0}
USER_DEBUG|[DEBUG]`, name)
}

func TestRunBenchmarkWithExecutor_Success(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mock := &mockExecutor{}
	spec := types.CodeSpec{
		Name:       "TestBench",
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     2,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "json", 1, 1)

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if !strings.Contains(output, "TestBench") {
		t.Errorf("Expected output to contain benchmark name, got: %s", output)
	}
	if !strings.Contains(output, "avgCpuMs") {
		t.Errorf("Expected JSON output with avgCpuMs field, got: %s", output)
	}
}

func TestRunBenchmarkWithExecutor_TableOutput(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mock := &mockExecutor{}
	spec := types.CodeSpec{
		Name:       "TableTest",
		UserCode:   "Integer x = 1;",
		Iterations: 5,
		Warmup:     1,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "table", 1, 1)

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Table output should contain table characters
	if !strings.Contains(output, "│") && !strings.Contains(output, "┌") {
		t.Errorf("Expected table output, got: %s", output)
	}
}

func TestRunBenchmarkWithExecutor_MultipleRuns(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mock := &mockExecutor{}
	spec := types.CodeSpec{
		Name:       "MultiRun",
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     2,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "json", 3, 2)

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestRunBenchmarkWithExecutor_ExecutionError(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{
		runFunc: func(apexCode string, org string) (string, error) {
			return "", fmt.Errorf("execution failed")
		},
	}

	spec := types.CodeSpec{
		Name:       "ErrorTest",
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     2,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "json", 1, 1)

	if err == nil {
		t.Error("Expected error, got success")
	}
	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected execution error, got: %v", err)
	}
}

func TestRunBenchmarkWithExecutor_ParallelExecutionError(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{
		executeParallelFunc: func(apexCode string, runs int, maxConcurrent int, org string) ([]string, error) {
			return nil, fmt.Errorf("parallel execution failed")
		},
	}

	spec := types.CodeSpec{
		Name:       "ParallelErrorTest",
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     2,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "json", 3, 2)

	if err == nil {
		t.Error("Expected error, got success")
	}
	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("Expected parallel execution error, got: %v", err)
	}
}

func TestRunBenchmarkWithExecutor_InvalidOutputFormat(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{}
	spec := types.CodeSpec{
		Name:       "InvalidFormat",
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     2,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "xml", 1, 1)

	if err == nil {
		t.Error("Expected error for invalid output format")
	}
	if !strings.Contains(err.Error(), "unknown output format") {
		t.Errorf("Expected format error, got: %v", err)
	}
}

func TestRunBenchmarkWithExecutor_InvalidBenchmarkSpec(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	mock := &mockExecutor{}
	spec := types.CodeSpec{
		Name:       "", // Invalid: empty name
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     2,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "json", 1, 1)

	if err == nil {
		t.Error("Expected error for invalid spec")
	}
	if !strings.Contains(err.Error(), "failed to generate code") {
		t.Errorf("Expected generation error, got: %v", err)
	}
}

func TestRunBenchmarkWithExecutor_ParseError(t *testing.T) {
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

	spec := types.CodeSpec{
		Name:       "ParseError",
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     2,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "json", 1, 1)

	if err == nil {
		t.Error("Expected parse error")
	}
	if !strings.Contains(err.Error(), "failed to parse results") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestRunBenchmarkWithExecutor_WithTrackingOptions(t *testing.T) {
	// Redirect stderr to suppress log output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	os.Stderr, _ = os.Open(os.DevNull)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	mock := &mockExecutor{
		runFunc: func(apexCode string, org string) (string, error) {
			// Verify that heap and DB tracking code is in the generated code
			if !strings.Contains(apexCode, "Limits.getHeapSize()") {
				return "", fmt.Errorf("expected heap tracking code")
			}
			if !strings.Contains(apexCode, "Limits.getDmlStatements()") {
				return "", fmt.Errorf("expected DB tracking code")
			}
			return mockSuccessfulBenchResult(), nil
		},
	}

	spec := types.CodeSpec{
		Name:       "TrackingTest",
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     2,
		TrackHeap:  true,
		TrackDB:    true,
	}

	err := runBenchmarkWithExecutor(mock, "test-org", spec, "json", 1, 1)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}
