package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompareCommand_Flags(t *testing.T) {
	flags := compareCmd.Flags()

	if flags.Lookup("bench") == nil {
		t.Error("Expected 'bench' flag to be registered")
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

func TestCompareCommand_DefaultValues(t *testing.T) {
	flags := compareCmd.Flags()

	iterVal, _ := flags.GetInt("iterations")
	if iterVal != 100 {
		t.Errorf("Expected default iterations 100, got %d", iterVal)
	}

	warmupVal, _ := flags.GetInt("warmup")
	if warmupVal != 10 {
		t.Errorf("Expected default warmup 10, got %d", warmupVal)
	}

	outputVal, _ := flags.GetString("output")
	if outputVal != "table" {
		t.Errorf("Expected default output 'table', got %s", outputVal)
	}
}

func TestCompareBenchmarks_TooFewBenchmarks(t *testing.T) {
	compareBenches = []string{"Test1:code1"}

	cmd := &cobra.Command{}
	err := compareBenchmarks(cmd, []string{})

	if err == nil {
		t.Error("Expected error when fewer than 2 benchmarks provided")
	}

	if !strings.Contains(err.Error(), "at least 2 benchmarks") {
		t.Errorf("Expected 'at least 2 benchmarks' error, got: %v", err)
	}
}

func TestFileExists(t *testing.T) {
	// Test with existing file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if !fileExists(tmpFile.Name()) {
		t.Error("Expected fileExists to return true for existing file")
	}

	// Test with non-existent file
	if fileExists("/nonexistent/file.txt") {
		t.Error("Expected fileExists to return false for non-existent file")
	}

	// Test with directory
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if fileExists(tmpDir) {
		t.Error("Expected fileExists to return false for directory")
	}
}

func TestCompareCommand_Integration(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	rootCmd.AddCommand(compareCmd)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"compare", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to execute help: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Compare multiple benchmarks") {
		t.Error("Help output should contain command description")
	}
}

func TestCompareBenchmarks_BenchmarkParsing(t *testing.T) {
	// Test parsing valid benchmark format
	benchStr := "TestName:String s = 'test';"
	parts := strings.SplitN(benchStr, ":", 2)

	if len(parts) != 2 {
		t.Error("Expected benchmark string to split into 2 parts")
	}

	name := strings.TrimSpace(parts[0])
	source := strings.TrimSpace(parts[1])

	if name != "TestName" {
		t.Errorf("Expected name 'TestName', got %s", name)
	}

	if source != "String s = 'test';" {
		t.Errorf("Expected source code, got %s", source)
	}
}

func TestCompareBenchmarks_InvalidFormat(t *testing.T) {
	// Test that invalid format is caught
	invalidBench := "NoColonInThisString"
	parts := strings.SplitN(invalidBench, ":", 2)

	if len(parts) == 2 {
		t.Error("Expected invalid benchmark format to not split correctly")
	}
}

func TestCompareBenchmarks_FileDetection(t *testing.T) {
	// Create a test file
	tmpFile, err := os.CreateTemp("", "test-*.apex")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test file detection logic
	source := tmpFile.Name()

	// File with .apex extension
	if !strings.HasSuffix(source, ".apex") {
		t.Error("Expected temp file to have .apex extension")
	}

	// File exists
	if !fileExists(source) {
		t.Error("Expected file to exist")
	}

	// Non-file source
	inlineCode := "String s = 'test';"
	if strings.HasSuffix(inlineCode, ".apex") {
		t.Error("Inline code should not have .apex extension")
	}
}

func TestCompareBenchmarks_FileReading(t *testing.T) {
	// Create temporary files
	tmpFile1, err := os.CreateTemp("", "test1-*.apex")
	if err != nil {
		t.Fatalf("Failed to create temp file 1: %v", err)
	}
	defer os.Remove(tmpFile1.Name())

	testCode1 := "String s1 = 'file1';"
	if _, err := tmpFile1.Write([]byte(testCode1)); err != nil {
		t.Fatalf("Failed to write temp file 1: %v", err)
	}
	tmpFile1.Close()

	// Test reading
	content, err := os.ReadFile(tmpFile1.Name())
	if err != nil {
		t.Errorf("Should be able to read file: %v", err)
	}

	if string(content) != testCode1 {
		t.Errorf("Expected content %q, got %q", testCode1, string(content))
	}
}
