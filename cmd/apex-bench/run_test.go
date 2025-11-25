package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunCommand_Flags(t *testing.T) {
	// Test that all flags are registered
	flags := runCmd.Flags()

	if flags.Lookup("code") == nil {
		t.Error("Expected 'code' flag to be registered")
	}
	if flags.Lookup("file") == nil {
		t.Error("Expected 'file' flag to be registered")
	}
	if flags.Lookup("name") == nil {
		t.Error("Expected 'name' flag to be registered")
	}
	if flags.Lookup("iterations") == nil {
		t.Error("Expected 'iterations' flag to be registered")
	}
	if flags.Lookup("warmup") == nil {
		t.Error("Expected 'warmup' flag to be registered")
	}
	if flags.Lookup("runs") == nil {
		t.Error("Expected 'runs' flag to be registered")
	}
	if flags.Lookup("parallel") == nil {
		t.Error("Expected 'parallel' flag to be registered")
	}
	if flags.Lookup("track-heap") == nil {
		t.Error("Expected 'track-heap' flag to be registered")
	}
	if flags.Lookup("track-db") == nil {
		t.Error("Expected 'track-db' flag to be registered")
	}
	if flags.Lookup("org") == nil {
		t.Error("Expected 'org' flag to be registered")
	}
	if flags.Lookup("output") == nil {
		t.Error("Expected 'output' flag to be registered")
	}
}

func TestRunCommand_DefaultValues(t *testing.T) {
	// Test default flag values
	flags := runCmd.Flags()

	iterVal, _ := flags.GetInt("iterations")
	if iterVal != 100 {
		t.Errorf("Expected default iterations 100, got %d", iterVal)
	}

	warmupVal, _ := flags.GetInt("warmup")
	if warmupVal != 10 {
		t.Errorf("Expected default warmup 10, got %d", warmupVal)
	}

	runsVal, _ := flags.GetInt("runs")
	if runsVal != 1 {
		t.Errorf("Expected default runs 1, got %d", runsVal)
	}

	parallelVal, _ := flags.GetInt("parallel")
	if parallelVal != 1 {
		t.Errorf("Expected default parallel 1, got %d", parallelVal)
	}

	outputVal, _ := flags.GetString("output")
	if outputVal != "json" {
		t.Errorf("Expected default output 'json', got %s", outputVal)
	}
}

func TestRunBenchmark_NoCodeOrFile(t *testing.T) {
	// Reset flags
	runCode = ""
	runFile = ""

	cmd := &cobra.Command{}
	err := runBenchmark(cmd, []string{})

	if err == nil {
		t.Error("Expected error when no code or file provided")
	}

	if !strings.Contains(err.Error(), "must provide either --code or --file") {
		t.Errorf("Expected 'must provide either --code or --file' error, got: %v", err)
	}
}

func TestRunBenchmark_BothCodeAndFile(t *testing.T) {
	runCode = "String s = 'test';"
	runFile = "test.apex"

	cmd := &cobra.Command{}
	err := runBenchmark(cmd, []string{})

	if err == nil {
		t.Error("Expected error when both code and file provided")
	}

	if !strings.Contains(err.Error(), "cannot provide both --code and --file") {
		t.Errorf("Expected 'cannot provide both' error, got: %v", err)
	}
}

func TestRunCommand_Integration(t *testing.T) {
	// Test the cobra command setup
	rootCmd := &cobra.Command{Use: "test"}
	rootCmd.AddCommand(runCmd)

	// Test help
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"run", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to execute help: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Run a benchmark") {
		t.Error("Help output should contain command description")
	}
}

func TestRunCommand_FileReadScenario(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.apex")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testCode := "String s = 'from file';"
	if _, err := tmpFile.Write([]byte(testCode)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Verify file exists and can be read (simulating what runBenchmark does)
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Errorf("Should be able to read test file: %v", err)
	}

	if string(content) != testCode {
		t.Errorf("Expected file content %q, got %q", testCode, string(content))
	}
}

func TestRunCommand_FileNotFoundScenario(t *testing.T) {
	// Test that reading a non-existent file returns error
	_, err := os.ReadFile("/nonexistent/file.apex")
	if err == nil {
		t.Error("Expected error when reading non-existent file")
	}
}
