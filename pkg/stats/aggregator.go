package stats

import (
	"fmt"
	"math"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

// Aggregate combines multiple Results and calculates statistics
func Aggregate(results []types.Result) (types.AggregatedResult, error) {
	if len(results) == 0 {
		return types.AggregatedResult{}, fmt.Errorf("cannot aggregate empty results")
	}

	// Use first result for metadata
	first := results[0]
	agg := types.AggregatedResult{
		Name:       first.Name,
		Runs:       len(results),
		Iterations: first.Iterations,
		Warmup:     0, // Warmup not tracked in Result, would need to pass separately
		RawResults: results,
	}

	// Aggregate CPU time
	cpuTimes := make([]float64, len(results))
	minCpu := results[0].MinCpuMs
	maxCpu := results[0].MaxCpuMs
	for i, r := range results {
		cpuTimes[i] = r.AvgCpuMs
		if r.MinCpuMs < minCpu {
			minCpu = r.MinCpuMs
		}
		if r.MaxCpuMs > maxCpu {
			maxCpu = r.MaxCpuMs
		}
	}
	agg.AvgCpuMs = mean(cpuTimes)
	agg.StdDevCpuMs = stdDev(cpuTimes)
	agg.MinCpuMs = minCpu
	agg.MaxCpuMs = maxCpu

	// Aggregate wall time
	wallTimes := make([]float64, len(results))
	minWall := results[0].MinWallMs
	maxWall := results[0].MaxWallMs
	for i, r := range results {
		wallTimes[i] = r.AvgWallMs
		if r.MinWallMs < minWall {
			minWall = r.MinWallMs
		}
		if r.MaxWallMs > maxWall {
			maxWall = r.MaxWallMs
		}
	}
	agg.AvgWallMs = mean(wallTimes)
	agg.StdDevWallMs = stdDev(wallTimes)
	agg.MinWallMs = minWall
	agg.MaxWallMs = maxWall

	return agg, nil
}

// mean calculates the arithmetic mean of a slice of float64
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// stdDev calculates the standard deviation of a slice of float64
func stdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	avg := mean(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - avg
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}
