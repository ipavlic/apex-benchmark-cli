package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestHelperProcess is used by TestMain to provide mock command execution
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Get the command arguments
	args := os.Args

	// Find where our mock args start (after "--")
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "not enough arguments")
		os.Exit(1)
	}

	// First arg is the command name (sf)
	// Second arg is the subcommand
	cmd := args[0]

	if cmd != "sf" {
		fmt.Fprintf(os.Stderr, "unknown command: %s", cmd)
		os.Exit(1)
	}

	subcommand := args[1]

	switch subcommand {
	case "--version":
		fmt.Fprintf(os.Stdout, "@salesforce/cli/2.0.0 darwin-arm64 node-v18.0.0")
		os.Exit(0)

	case "apex":
		if len(args) > 2 && args[2] == "run" {
			// Mock apex run success with JSON response
			jsonResponse := `{
  "status": 0,
  "result": {
    "success": true,
    "compiled": true,
    "compileProblem": "",
    "exceptionMessage": "",
    "exceptionStackTrace": "",
    "line": -1,
    "column": -1,
    "logs": "15:45:09.123 (456789)|USER_DEBUG|[1]|DEBUG|BENCH_RESULT:{\"cpuMs\":10.5,\"elapsedMs\":15.2}"
  }
}`
			fmt.Fprint(os.Stdout, jsonResponse)
			os.Exit(0)
		}

	case "config":
		if len(args) > 3 && args[2] == "get" && args[3] == "target-org" {
			// Mock config get target-org
			if os.Getenv("MOCK_NO_DEFAULT_ORG") == "1" {
				fmt.Fprintf(os.Stdout, `{"status":0,"result":[]}`)
			} else {
				fmt.Fprintf(os.Stdout, `{"status":0,"result":[{"name":"target-org","value":"test-org"}]}`)
			}
			os.Exit(0)
		}

	case "org":
		if len(args) > 2 && args[2] == "display" {
			// Mock org display success with JSON response
			if os.Getenv("MOCK_ORG_AUTH_FAIL") == "1" {
				fmt.Fprintf(os.Stderr, "org not authenticated")
				os.Exit(1)
			}
			jsonResponse := `{
  "status": 0,
  "result": {
    "id": "00D000000000000",
    "instanceUrl": "https://test.salesforce.com",
    "username": "test@example.com",
    "connectedStatus": "Connected",
    "alias": "test-org"
  }
}`
			fmt.Fprint(os.Stdout, jsonResponse)
			os.Exit(0)
		}
	}

	fmt.Fprintf(os.Stderr, "unknown subcommand: %s %v", subcommand, args)
	os.Exit(2)
}

// mockCommand creates a mock exec.Command that uses TestHelperProcess
func mockCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestCLIExecutor_Run_Success(t *testing.T) {
	// Override exec.Command temporarily
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	output, err := executor.Run("String s = 'test';", "test-org")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(output, "BENCH_RESULT") {
		t.Errorf("Expected output to contain BENCH_RESULT, got: %s", output)
	}
}

func TestCLIExecutor_Run_WithoutOrg(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	output, err := executor.Run("String s = 'test';", "")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(output, "BENCH_RESULT") {
		t.Errorf("Expected output to contain BENCH_RESULT, got: %s", output)
	}
}

func TestCheckSalesforceCLI_Success(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	err := CheckSalesforceCLI()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestCheckSalesforceCLI_NotInstalled(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		cmd := mockCommand(command, args...)
		cmd.Env = append(cmd.Env, "PATH=/nonexistent")
		return exec.Command("nonexistent-command")
	}
	defer func() { execCommand = oldExecCommand }()

	err := CheckSalesforceCLI()
	if err == nil {
		t.Error("Expected error when sf CLI not installed")
	}
}

func TestGetDefaultOrg_Success(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	org, err := GetDefaultOrg()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if org != "test-org" {
		t.Errorf("Expected org 'test-org', got: %s", org)
	}
}

func TestGetDefaultOrg_NoDefaultOrg(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		cmd := mockCommand(command, args...)
		cmd.Env = append(cmd.Env, "MOCK_NO_DEFAULT_ORG=1")
		return cmd
	}
	defer func() { execCommand = oldExecCommand }()

	_, err := GetDefaultOrg()
	if err == nil {
		t.Error("Expected error when no default org configured")
	}

	if !strings.Contains(err.Error(), "no default org") {
		t.Errorf("Expected 'no default org' error, got: %v", err)
	}
}

func TestGetOrg_WithSpecified(t *testing.T) {
	org, err := GetOrg("my-specified-org")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if org != "my-specified-org" {
		t.Errorf("Expected org 'my-specified-org', got: %s", org)
	}
}

func TestGetOrg_WithoutSpecified_UsesDefault(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	org, err := GetOrg("")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if org != "test-org" {
		t.Errorf("Expected org 'test-org', got: %s", org)
	}
}

func TestGetOrg_NoDefault_ReturnsError(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		cmd := mockCommand(command, args...)
		cmd.Env = append(cmd.Env, "MOCK_NO_DEFAULT_ORG=1")
		return cmd
	}
	defer func() { execCommand = oldExecCommand }()

	_, err := GetOrg("")
	if err == nil {
		t.Error("Expected error when no org specified and no default")
	}
}

func TestCheckOrgAuth_Success(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	err := CheckOrgAuth("test-org")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestCheckOrgAuth_NotAuthenticated(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		cmd := mockCommand(command, args...)
		cmd.Env = append(cmd.Env, "MOCK_ORG_AUTH_FAIL=1")
		return cmd
	}
	defer func() { execCommand = oldExecCommand }()

	err := CheckOrgAuth("bad-org")
	if err == nil {
		t.Error("Expected error for unauthenticated org")
	}

	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("Expected 'not authenticated' error, got: %v", err)
	}
}

func TestCheckOrgAuth_EmptyOrg(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	err := CheckOrgAuth("")
	if err != nil {
		t.Errorf("Expected no error for empty org, got: %v", err)
	}
}

func TestExecuteParallel_Success(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	results, err := executor.ExecuteParallel("String s = 'test';", 3, 2, "test-org")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if !strings.Contains(result, "BENCH_RESULT") {
			t.Errorf("Result %d: expected to contain BENCH_RESULT, got: %s", i, result)
		}
	}
}

func TestExecuteParallel_DefaultMaxConcurrent(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = mockCommand
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	// Test with maxConcurrent = 0, should default to 1
	results, err := executor.ExecuteParallel("String s = 'test';", 2, 0, "test-org")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestCreateTempApexFile_WriteError(t *testing.T) {
	// Test write error by passing extremely large string (if we wanted to simulate)
	// For now, just test that it works with normal input
	code := "String s = 'test';"

	tempFile, err := createTempApexFile(code)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer os.Remove(tempFile)

	// Verify file content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if string(content) != code {
		t.Errorf("Expected content %q, got %q", code, string(content))
	}
}

func TestCLIExecutor_Run_Error(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return a command that will fail
		return exec.Command("false")
	}
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	_, err := executor.Run("String s = 'test';", "test-org")

	if err == nil {
		t.Error("Expected error when command fails")
	}
}

func TestExecuteParallel_SingleError(t *testing.T) {
	oldExecCommand := execCommand
	callCount := 0
	execCommand = func(command string, args ...string) *exec.Cmd {
		callCount++
		if callCount == 2 {
			// Make the second execution fail
			return exec.Command("false")
		}
		return mockCommand(command, args...)
	}
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	_, err := executor.ExecuteParallel("String s = 'test';", 3, 1, "test-org")

	if err == nil {
		t.Error("Expected error when one execution fails")
	}

	if !strings.Contains(err.Error(), "execution errors") {
		t.Errorf("Expected 'execution errors' in error message, got: %v", err)
	}
}

func TestCheckSalesforceCLI_UnexpectedOutput(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		cmd := mockCommand(command, args...)
		// Override output to not contain expected string
		cmd.Env = append(cmd.Env, "MOCK_INVALID_VERSION=1")
		// Create a command that returns different output
		return exec.Command("echo", "wrong output")
	}
	defer func() { execCommand = oldExecCommand }()

	err := CheckSalesforceCLI()
	if err == nil {
		t.Error("Expected error for unexpected CLI output")
	}
}

func TestGetDefaultOrg_InvalidJSON(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return invalid JSON
		return exec.Command("echo", "invalid json{")
	}
	defer func() { execCommand = oldExecCommand }()

	_, err := GetDefaultOrg()
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected 'failed to parse' error, got: %v", err)
	}
}

func TestGetDefaultOrg_EmptyValue(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return JSON with null value
		return exec.Command("echo", `{"status":0,"result":[{"name":"target-org","value":"null"}]}`)
	}
	defer func() { execCommand = oldExecCommand }()

	_, err := GetDefaultOrg()
	if err == nil {
		t.Error("Expected error for null org value")
	}

	if !strings.Contains(err.Error(), "no default org") {
		t.Errorf("Expected 'no default org' error, got: %v", err)
	}
}

func TestGetDefaultOrg_CommandError(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = oldExecCommand }()

	_, err := GetDefaultOrg()
	if err == nil {
		t.Error("Expected error when command fails")
	}

	if !strings.Contains(err.Error(), "failed to get default org") {
		t.Errorf("Expected 'failed to get default org' error, got: %v", err)
	}
}

func TestCLIExecutor_Run_CompilationError(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return JSON with compilation error
		cmd := exec.Command("echo", `{
  "status": 1,
  "result": {
    "success": false,
    "compiled": false,
    "compileProblem": "Unexpected token '}'",
    "exceptionMessage": "",
    "exceptionStackTrace": "",
    "line": 5,
    "column": 10,
    "logs": ""
  }
}`)
		return cmd
	}
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	_, err := executor.Run("String s = 'test';", "test-org")

	if err == nil {
		t.Error("Expected error for compilation failure")
	}

	if !strings.Contains(err.Error(), "Apex compilation failed") {
		t.Errorf("Expected 'Apex compilation failed' error, got: %v", err)
	}
}

func TestCLIExecutor_Run_ExecutionError(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return JSON with execution error
		cmd := exec.Command("echo", `{
  "status": 1,
  "result": {
    "success": false,
    "compiled": true,
    "compileProblem": "",
    "exceptionMessage": "System.NullPointerException: Attempt to de-reference a null object",
    "exceptionStackTrace": "Class.TestClass.method: line 10",
    "line": 10,
    "column": 5,
    "logs": ""
  }
}`)
		return cmd
	}
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	_, err := executor.Run("String s = 'test';", "test-org")

	if err == nil {
		t.Error("Expected error for execution failure")
	}

	if !strings.Contains(err.Error(), "Apex execution failed") {
		t.Errorf("Expected 'Apex execution failed' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "NullPointerException") {
		t.Errorf("Expected error message to contain exception details, got: %v", err)
	}
}

func TestCLIExecutor_Run_InvalidJSON(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return invalid JSON
		cmd := exec.Command("echo", "not valid json")
		return cmd
	}
	defer func() { execCommand = oldExecCommand }()

	executor := NewCLIExecutor()
	_, err := executor.Run("String s = 'test';", "test-org")

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected 'failed to parse' error, got: %v", err)
	}
}

func TestCheckOrgAuth_InvalidJSON(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return invalid JSON
		return exec.Command("echo", "invalid{json")
	}
	defer func() { execCommand = oldExecCommand }()

	err := CheckOrgAuth("test-org")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected 'failed to parse' error, got: %v", err)
	}
}

func TestCheckOrgAuth_NullResult(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return JSON with null result
		return exec.Command("echo", `{"status": 0, "result": null}`)
	}
	defer func() { execCommand = oldExecCommand }()

	err := CheckOrgAuth("test-org")
	if err == nil {
		t.Error("Expected error for null result")
	}

	if !strings.Contains(err.Error(), "not authenticated or not found") {
		t.Errorf("Expected 'not authenticated or not found' error, got: %v", err)
	}
}

func TestCheckOrgAuth_NotConnected(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return JSON with disconnected status
		return exec.Command("echo", `{
  "status": 0,
  "result": {
    "id": "00D000000000000",
    "username": "test@example.com",
    "connectedStatus": "Expired"
  }
}`)
	}
	defer func() { execCommand = oldExecCommand }()

	err := CheckOrgAuth("test-org")
	if err == nil {
		t.Error("Expected error for disconnected org")
	}

	if !strings.Contains(err.Error(), "Expired") {
		t.Errorf("Expected error about 'Expired' status, got: %v", err)
	}
}
