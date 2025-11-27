package executor

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// MockExecutor implements Executor for testing
type MockExecutor struct {
	Output      string
	Error       error
	CallCount   int
	LastCode    string
	LastOrg     string
	ShouldDelay bool
}

func (m *MockExecutor) Run(apexCode string, org string) (string, error) {
	m.CallCount++
	m.LastCode = apexCode
	m.LastOrg = org
	if m.Error != nil {
		return "", m.Error
	}
	return m.Output, nil
}

func (m *MockExecutor) ExecuteParallel(apexCode string, runs int, maxConcurrent int, org string) ([]string, error) {
	results := make([]string, runs)
	for i := 0; i < runs; i++ {
		output, err := m.Run(apexCode, org)
		if err != nil {
			return nil, err
		}
		results[i] = output
	}
	return results, nil
}

func TestCLIExecutor_Run_CreatesTempFile(t *testing.T) {
	// This test would require mocking exec.Command
	// For now, we'll skip actual execution tests
	// In a real scenario, you'd use a package like github.com/golang/mock
	t.Skip("Integration test - requires sf CLI")
}

func TestExecuteParallel_InvalidRuns(t *testing.T) {
	executor := &CLIExecutor{}
	_, err := executor.ExecuteParallel("String s = 'test';", 0, 1, "")
	if err == nil {
		t.Error("Expected error for zero runs")
	}

	_, err = executor.ExecuteParallel("String s = 'test';", -1, 1, "")
	if err == nil {
		t.Error("Expected error for negative runs")
	}
}

func TestMockExecutor_Run(t *testing.T) {
	mock := &MockExecutor{
		Output: "DEBUG|BENCH_RESULT:{...}",
		Error:  nil,
	}

	code := "String s = 'test';"
	org := "my-org"

	output, err := mock.Run(code, org)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output != mock.Output {
		t.Errorf("Expected output %q, got %q", mock.Output, output)
	}

	if mock.LastCode != code {
		t.Errorf("Expected last code %q, got %q", code, mock.LastCode)
	}

	if mock.LastOrg != org {
		t.Errorf("Expected last org %q, got %q", org, mock.LastOrg)
	}

	if mock.CallCount != 1 {
		t.Errorf("Expected call count 1, got %d", mock.CallCount)
	}
}

func TestMockExecutor_Run_WithError(t *testing.T) {
	expectedErr := fmt.Errorf("execution failed")
	mock := &MockExecutor{
		Error: expectedErr,
	}

	_, err := mock.Run("String s = 'test';", "")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockExecutor_ExecuteParallel(t *testing.T) {
	mock := &MockExecutor{
		Output: "DEBUG|BENCH_RESULT:{...}",
	}

	runs := 5
	results, err := mock.ExecuteParallel("String s = 'test';", runs, 2, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != runs {
		t.Errorf("Expected %d results, got %d", runs, len(results))
	}

	for i, result := range results {
		if result != mock.Output {
			t.Errorf("Result %d: expected %q, got %q", i, mock.Output, result)
		}
	}

	if mock.CallCount != runs {
		t.Errorf("Expected call count %d, got %d", runs, mock.CallCount)
	}
}

func TestCreateTempApexFile(t *testing.T) {
	code := "String s = 'hello';"

	tempFile, err := createTempApexFile(code)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	// Verify file exists and contains the code
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if string(content) != code {
		t.Errorf("File content mismatch. Expected %q, got %q", code, string(content))
	}

	// Verify file has .apex extension
	if !strings.HasSuffix(tempFile, ".apex") {
		t.Errorf("Expected .apex extension, got %s", tempFile)
	}
}

func TestCheckSalesforceCLI(t *testing.T) {
	// This is an integration test that requires sf CLI
	// We'll skip it in normal test runs
	t.Skip("Integration test - requires sf CLI installation")

	err := CheckSalesforceCLI()
	if err != nil {
		t.Logf("sf CLI check failed (expected if not installed): %v", err)
	}
}

func TestGetDefaultOrg(t *testing.T) {
	// This is an integration test that requires sf CLI and org setup
	t.Skip("Integration test - requires sf CLI and authenticated org")

	org, err := GetDefaultOrg()
	if err != nil {
		t.Logf("Get default org failed (expected if not configured): %v", err)
	} else {
		t.Logf("Default org: %s", org)
	}
}
