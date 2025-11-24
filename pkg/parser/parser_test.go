package parser

import (
	"strings"
	"testing"
)

func TestParseResult_ValidJSON(t *testing.T) {
	output := `Execute Anonymous: System.debug('BENCH_RESULT:' + resultJson);
13:45:23.123 (123456)|USER_DEBUG|[1]|DEBUG|BENCH_RESULT:{"name":"TestBench","iterations":100,"avgWallMs":1.5,"avgCpuMs":1.2,"minWallMs":1.0,"maxWallMs":2.0,"minCpuMs":1.0,"maxCpuMs":1.5}
13:45:23.456 (456789)|CUMULATIVE_LIMIT_USAGE`

	result, err := ParseResult(output)
	if err != nil {
		t.Fatalf("ParseResult failed: %v", err)
	}

	if result.Name != "TestBench" {
		t.Errorf("Expected name 'TestBench', got %q", result.Name)
	}

	if result.Iterations != 100 {
		t.Errorf("Expected iterations 100, got %d", result.Iterations)
	}

	if result.AvgWallMs != 1.5 {
		t.Errorf("Expected avgWallMs 1.5, got %f", result.AvgWallMs)
	}

	if result.AvgCpuMs != 1.2 {
		t.Errorf("Expected avgCpuMs 1.2, got %f", result.AvgCpuMs)
	}
}

func TestParseResult_WithHeapData(t *testing.T) {
	output := `USER_DEBUG|BENCH_RESULT:{"name":"HeapTest","iterations":50,"avgWallMs":2.0,"avgCpuMs":1.8,"minWallMs":1.5,"maxWallMs":2.5,"minCpuMs":1.5,"maxCpuMs":2.0,"avgHeapKb":125.5,"minHeapKb":100.0,"maxHeapKb":150.0}`

	result, err := ParseResult(output)
	if err != nil {
		t.Fatalf("ParseResult failed: %v", err)
	}

	if result.AvgHeapKb == nil {
		t.Fatal("Expected heap data to be present")
	}

	if *result.AvgHeapKb != 125.5 {
		t.Errorf("Expected avgHeapKb 125.5, got %f", *result.AvgHeapKb)
	}
}

func TestParseResult_WithDBData(t *testing.T) {
	output := `USER_DEBUG|BENCH_RESULT:{"name":"DBTest","iterations":20,"avgWallMs":5.0,"avgCpuMs":4.5,"minWallMs":4.0,"maxWallMs":6.0,"minCpuMs":4.0,"maxCpuMs":5.0,"dmlStatements":2,"soqlQueries":5}`

	result, err := ParseResult(output)
	if err != nil {
		t.Fatalf("ParseResult failed: %v", err)
	}

	if result.DmlStatements == nil {
		t.Fatal("Expected DML statements to be present")
	}

	if *result.DmlStatements != 2 {
		t.Errorf("Expected dmlStatements 2, got %d", *result.DmlStatements)
	}

	if result.SoqlQueries == nil {
		t.Fatal("Expected SOQL queries to be present")
	}

	if *result.SoqlQueries != 5 {
		t.Errorf("Expected soqlQueries 5, got %d", *result.SoqlQueries)
	}
}

func TestParseResult_NoMarker(t *testing.T) {
	output := `Some debug output without the marker
USER_DEBUG|Something else
More output`

	_, err := ParseResult(output)
	if err == nil {
		t.Error("Expected error when BENCH_RESULT marker not found")
	}

	if !strings.Contains(err.Error(), "could not find") {
		t.Errorf("Expected 'could not find' error, got: %v", err)
	}
}

func TestParseResult_InvalidJSON(t *testing.T) {
	// JSON with syntax error (trailing comma, missing quotes, etc)
	output := `USER_DEBUG|BENCH_RESULT:{"name":"Bad",invalid:,}`

	_, err := ParseResult(output)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if err != nil && !strings.Contains(err.Error(), "could not find valid BENCH_RESULT") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestParseMultipleResults_Valid(t *testing.T) {
	outputs := []string{
		`USER_DEBUG|BENCH_RESULT:{"name":"Test1","iterations":10,"avgWallMs":1.0,"avgCpuMs":0.9,"minWallMs":0.8,"maxWallMs":1.2,"minCpuMs":0.8,"maxCpuMs":1.0}`,
		`USER_DEBUG|BENCH_RESULT:{"name":"Test2","iterations":10,"avgWallMs":1.1,"avgCpuMs":1.0,"minWallMs":0.9,"maxWallMs":1.3,"minCpuMs":0.9,"maxCpuMs":1.1}`,
		`USER_DEBUG|BENCH_RESULT:{"name":"Test3","iterations":10,"avgWallMs":1.2,"avgCpuMs":1.1,"minWallMs":1.0,"maxWallMs":1.4,"minCpuMs":1.0,"maxCpuMs":1.2}`,
	}

	results, err := ParseMultipleResults(outputs)
	if err != nil {
		t.Fatalf("ParseMultipleResults failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if results[0].Name != "Test1" {
		t.Errorf("Expected first result name 'Test1', got %q", results[0].Name)
	}

	if results[1].Name != "Test2" {
		t.Errorf("Expected second result name 'Test2', got %q", results[1].Name)
	}

	if results[2].Name != "Test3" {
		t.Errorf("Expected third result name 'Test3', got %q", results[2].Name)
	}
}

func TestParseMultipleResults_WithErrors(t *testing.T) {
	outputs := []string{
		`USER_DEBUG|BENCH_RESULT:{"name":"Test1","iterations":10,"avgWallMs":1.0,"avgCpuMs":0.9,"minWallMs":0.8,"maxWallMs":1.2,"minCpuMs":0.8,"maxCpuMs":1.0}`,
		`Invalid output without marker`,
		`USER_DEBUG|BENCH_RESULT:{"name":"Test3","iterations":10,"avgWallMs":1.2,"avgCpuMs":1.1,"minWallMs":1.0,"maxWallMs":1.4,"minCpuMs":1.0,"maxCpuMs":1.2}`,
	}

	_, err := ParseMultipleResults(outputs)
	if err == nil {
		t.Error("Expected error when one output is invalid")
	}

	if !strings.Contains(err.Error(), "failed to parse some results") {
		t.Errorf("Expected 'failed to parse' error, got: %v", err)
	}
}

func TestExtractDebugLines(t *testing.T) {
	output := `Execute Anonymous: some code
13:45:23.123 (123456)|EXECUTION_STARTED
13:45:23.234 (234567)|USER_DEBUG|[1]|DEBUG|BENCH_RESULT:{"name":"Test"}
13:45:23.345 (345678)|USER_DEBUG|[2]|DEBUG|Some other debug
13:45:23.456 (456789)|CUMULATIVE_LIMIT_USAGE`

	debugLines := ExtractDebugLines(output)

	if len(debugLines) != 2 {
		t.Errorf("Expected 2 debug lines, got %d", len(debugLines))
	}

	foundBenchResult := false
	for _, line := range debugLines {
		if strings.Contains(line, "BENCH_RESULT") {
			foundBenchResult = true
			break
		}
	}

	if !foundBenchResult {
		t.Error("Expected to find BENCH_RESULT in debug lines")
	}
}

func TestExtractDebugLines_Empty(t *testing.T) {
	output := `Execute Anonymous: some code
13:45:23.123 (123456)|EXECUTION_STARTED
13:45:23.456 (456789)|CUMULATIVE_LIMIT_USAGE`

	debugLines := ExtractDebugLines(output)

	if len(debugLines) != 0 {
		t.Errorf("Expected 0 debug lines, got %d", len(debugLines))
	}
}
