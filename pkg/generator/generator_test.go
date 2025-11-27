package generator

import (
	"strings"
	"testing"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

func TestGenerate_BasicCode(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "TestBenchmark",
		UserCode:   "String s = 'hello';",
		Iterations: 100,
		Warmup:     10,
		TrackHeap:  false,
		TrackDB:    false,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify the generated code contains expected elements
	expectations := []string{
		"TestBenchmark",
		"String s = 'hello';",
		"Integer warmupIterations = 10;",
		"Integer measurementIterations = 100;",
		"BENCH_RESULT:",
		"< warmupIterations;",  // Loop uses UUID-based variable
		"< measurementIterations;", // Loop uses UUID-based variable
		"Long wallStart = System.now().getTime();",
		"Integer cpuStart = Limits.getCpuTime();",
	}

	for _, expected := range expectations {
		if !strings.Contains(result, expected) {
			t.Errorf("Generated code missing expected string: %q", expected)
		}
	}
}

func TestGenerate_WithHeapTracking(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "HeapTest",
		UserCode:   "List<String> lst = new List<String>();",
		Iterations: 50,
		Warmup:     5,
		TrackHeap:  true,
		TrackDB:    false,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify heap tracking code is present
	heapExpectations := []string{
		"Long totalHeapUsed = 0;",
		"Long heapBefore = Limits.getHeapSize();",
		"Long heapAfter = Limits.getHeapSize();",
		"avgHeapKb",
		"minHeapKb",
		"maxHeapKb",
	}

	for _, expected := range heapExpectations {
		if !strings.Contains(result, expected) {
			t.Errorf("Generated code missing heap tracking: %q", expected)
		}
	}
}

func TestGenerate_WithDBTracking(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "DBTest",
		UserCode:   "Account acc = [SELECT Id FROM Account LIMIT 1];",
		Iterations: 20,
		Warmup:     2,
		TrackHeap:  false,
		TrackDB:    true,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify DB tracking code is present
	dbExpectations := []string{
		"Integer dmlStatementsBefore = Limits.getDmlStatements();",
		"Integer soqlQueriesBefore = Limits.getQueries();",
		"dmlStatements",
		"soqlQueries",
	}

	for _, expected := range dbExpectations {
		if !strings.Contains(result, expected) {
			t.Errorf("Generated code missing DB tracking: %q", expected)
		}
	}
}

func TestGenerate_WithSetupAndTeardown(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "SetupTeardownTest",
		UserCode:   "Integer x = 1 + 1;",
		Setup:      "System.debug('Setting up');",
		Teardown:   "System.debug('Tearing down');",
		Iterations: 10,
		Warmup:     1,
		TrackHeap:  false,
		TrackDB:    false,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify setup and teardown are present
	if !strings.Contains(result, "System.debug('Setting up');") {
		t.Error("Generated code missing setup code")
	}

	if !strings.Contains(result, "System.debug('Tearing down');") {
		t.Error("Generated code missing teardown code")
	}
}

func TestValidateSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    types.CodeSpec
		wantErr bool
	}{
		{
			name: "empty code",
			spec: types.CodeSpec{
				Name:       "Test",
				UserCode:   "",
				Iterations: 10,
				Warmup:     1,
			},
			wantErr: true,
		},
		{
			name: "whitespace only code",
			spec: types.CodeSpec{
				Name:       "Test",
				UserCode:   "   \t\n  ",
				Iterations: 10,
				Warmup:     1,
			},
			wantErr: true,
		},
		{
			name: "zero iterations",
			spec: types.CodeSpec{
				Name:       "Test",
				UserCode:   "String s = 'test';",
				Iterations: 0,
				Warmup:     1,
			},
			wantErr: true,
		},
		{
			name: "negative iterations",
			spec: types.CodeSpec{
				Name:       "Test",
				UserCode:   "String s = 'test';",
				Iterations: -10,
				Warmup:     1,
			},
			wantErr: true,
		},
		{
			name: "negative warmup",
			spec: types.CodeSpec{
				Name:       "Test",
				UserCode:   "String s = 'test';",
				Iterations: 10,
				Warmup:     -1,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			spec: types.CodeSpec{
				Name:       "",
				UserCode:   "String s = 'test';",
				Iterations: 10,
				Warmup:     1,
			},
			wantErr: true,
		},
		{
			name: "whitespace only name",
			spec: types.CodeSpec{
				Name:       "   ",
				UserCode:   "String s = 'test';",
				Iterations: 10,
				Warmup:     1,
			},
			wantErr: true,
		},
		{
			name: "valid spec",
			spec: types.CodeSpec{
				Name:       "ValidTest",
				UserCode:   "String s = 'test';",
				Iterations: 100,
				Warmup:     10,
			},
			wantErr: false,
		},
		{
			name: "valid spec with zero warmup",
			spec: types.CodeSpec{
				Name:       "NoWarmup",
				UserCode:   "Integer x = 1;",
				Iterations: 50,
				Warmup:     0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerate_WithBothHeapAndDB(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "BothTracking",
		UserCode:   "Integer x = 1;",
		Iterations: 10,
		Warmup:     1,
		TrackHeap:  true,
		TrackDB:    true,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Both tracking features should be present
	if !strings.Contains(result, "heapBefore") {
		t.Error("Missing heap tracking")
	}
	if !strings.Contains(result, "dmlStatementsBefore") {
		t.Error("Missing DML tracking")
	}
	if !strings.Contains(result, "soqlQueriesBefore") {
		t.Error("Missing SOQL tracking")
	}
}

func TestGenerate_LongUserCode(t *testing.T) {
	// Test with longer, multi-line user code
	spec := types.CodeSpec{
		Name: "LongCode",
		UserCode: `List<String> items = new List<String>();
for (Integer i = 0; i < 10; i++) {
    items.add('Item ' + i);
}
String result = String.join(items, ', ');`,
		Iterations: 50,
		Warmup:     5,
		TrackHeap:  false,
		TrackDB:    false,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// All lines of user code should be present
	if !strings.Contains(result, "List<String> items") {
		t.Error("Missing first line of user code")
	}
	if !strings.Contains(result, "items.add('Item ' + i);") {
		t.Error("Missing middle of user code")
	}
	if !strings.Contains(result, "String.join(items, ', ');") {
		t.Error("Missing last line of user code")
	}
}

func TestGenerate_SpecialCharacters(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "SpecialChars",
		UserCode:   "String s = 'test\\nwith\\ttabs';",
		Iterations: 10,
		Warmup:     1,
		TrackHeap:  false,
		TrackDB:    false,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(result, "test\\nwith\\ttabs") {
		t.Error("Special characters not preserved in user code")
	}
}

func TestGenerate_EmptySetup(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "EmptySetup",
		UserCode:   "Integer x = 1;",
		Setup:      "",
		Teardown:   "System.debug('cleanup');",
		Iterations: 10,
		Warmup:     1,
		TrackHeap:  false,
		TrackDB:    false,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should have teardown but not setup
	if !strings.Contains(result, "System.debug('cleanup');") {
		t.Error("Missing teardown code")
	}
}

func TestGenerate_MaxIterations(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "ManyIterations",
		UserCode:   "Integer x = 1;",
		Iterations: 10000,
		Warmup:     1000,
		TrackHeap:  false,
		TrackDB:    false,
	}

	result, err := Generate(spec)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(result, "Integer measurementIterations = 10000;") {
		t.Error("Large iteration count not set correctly")
	}
	if !strings.Contains(result, "Integer warmupIterations = 1000;") {
		t.Error("Large warmup count not set correctly")
	}
}

func TestGenerate_ValidationError(t *testing.T) {
	spec := types.CodeSpec{
		Name:       "", // Invalid: empty name
		UserCode:   "String s = 'test';",
		Iterations: 10,
		Warmup:     1,
		TrackHeap:  false,
		TrackDB:    false,
	}

	_, err := Generate(spec)
	if err == nil {
		t.Error("Expected error for invalid spec")
	}

	if !strings.Contains(err.Error(), "name") {
		t.Errorf("Expected error about name, got: %v", err)
	}
}
