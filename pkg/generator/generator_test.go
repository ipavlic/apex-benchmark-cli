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
		"for (Integer i = 0; i < warmupIterations; i++)",
		"for (Integer i = 0; i < measurementIterations; i++)",
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
