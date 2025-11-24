package stats

import (
	"math"
	"testing"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

func TestAggregate_SingleResult(t *testing.T) {
	results := []types.Result{
		{
			Name:       "Test",
			Iterations: 100,
			AvgWallMs:  1.5,
			AvgCpuMs:   1.2,
			MinWallMs:  1.0,
			MaxWallMs:  2.0,
			MinCpuMs:   1.0,
			MaxCpuMs:   1.5,
		},
	}

	agg, err := Aggregate(results)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if agg.Name != "Test" {
		t.Errorf("Expected name 'Test', got %q", agg.Name)
	}

	if agg.Runs != 1 {
		t.Errorf("Expected 1 run, got %d", agg.Runs)
	}

	if agg.AvgCpuMs != 1.2 {
		t.Errorf("Expected avgCpuMs 1.2, got %f", agg.AvgCpuMs)
	}

	if agg.StdDevCpuMs != 0 {
		t.Errorf("Expected stdDevCpuMs 0 for single result, got %f", agg.StdDevCpuMs)
	}
}

func TestAggregate_MultipleResults(t *testing.T) {
	results := []types.Result{
		{
			Name:       "Test",
			Iterations: 100,
			AvgWallMs:  1.0,
			AvgCpuMs:   0.9,
			MinWallMs:  0.8,
			MaxWallMs:  1.2,
			MinCpuMs:   0.7,
			MaxCpuMs:   1.1,
		},
		{
			Name:       "Test",
			Iterations: 100,
			AvgWallMs:  1.2,
			AvgCpuMs:   1.1,
			MinWallMs:  0.9,
			MaxWallMs:  1.5,
			MinCpuMs:   0.9,
			MaxCpuMs:   1.3,
		},
		{
			Name:       "Test",
			Iterations: 100,
			AvgWallMs:  1.1,
			AvgCpuMs:   1.0,
			MinWallMs:  0.85,
			MaxWallMs:  1.3,
			MinCpuMs:   0.8,
			MaxCpuMs:   1.2,
		},
	}

	agg, err := Aggregate(results)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if agg.Runs != 3 {
		t.Errorf("Expected 3 runs, got %d", agg.Runs)
	}

	// Check averages (1.0 + 1.2 + 1.1) / 3 = 1.1
	expectedAvgWall := 1.1
	if math.Abs(agg.AvgWallMs-expectedAvgWall) > 0.01 {
		t.Errorf("Expected avgWallMs ~%f, got %f", expectedAvgWall, agg.AvgWallMs)
	}

	// Check averages (0.9 + 1.1 + 1.0) / 3 = 1.0
	expectedAvgCpu := 1.0
	if math.Abs(agg.AvgCpuMs-expectedAvgCpu) > 0.01 {
		t.Errorf("Expected avgCpuMs ~%f, got %f", expectedAvgCpu, agg.AvgCpuMs)
	}

	// Check min/max
	if agg.MinCpuMs != 0.7 {
		t.Errorf("Expected minCpuMs 0.7, got %f", agg.MinCpuMs)
	}

	if agg.MaxCpuMs != 1.3 {
		t.Errorf("Expected maxCpuMs 1.3, got %f", agg.MaxCpuMs)
	}

	if agg.MinWallMs != 0.8 {
		t.Errorf("Expected minWallMs 0.8, got %f", agg.MinWallMs)
	}

	if agg.MaxWallMs != 1.5 {
		t.Errorf("Expected maxWallMs 1.5, got %f", agg.MaxWallMs)
	}

	// Check standard deviation is non-zero
	if agg.StdDevCpuMs == 0 {
		t.Error("Expected non-zero standard deviation for CPU time")
	}

	if agg.StdDevWallMs == 0 {
		t.Error("Expected non-zero standard deviation for wall time")
	}
}

func TestAggregate_EmptyResults(t *testing.T) {
	results := []types.Result{}

	_, err := Aggregate(results)
	if err == nil {
		t.Error("Expected error for empty results")
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"single value", []float64{5.0}, 5.0},
		{"multiple values", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, 3.0},
		{"decimals", []float64{1.5, 2.5, 3.5}, 2.5},
		{"empty", []float64{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mean(tt.values)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestStdDev(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"single value", []float64{5.0}, 0.0},
		{"same values", []float64{3.0, 3.0, 3.0}, 0.0},
		{"simple set", []float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0}, 2.0},
		{"empty", []float64{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stdDev(tt.values)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

