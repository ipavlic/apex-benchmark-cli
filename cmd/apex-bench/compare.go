package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ipavlic/apex-benchmark-cli/pkg/executor"
	"github.com/ipavlic/apex-benchmark-cli/pkg/generator"
	"github.com/ipavlic/apex-benchmark-cli/pkg/parser"
	"github.com/ipavlic/apex-benchmark-cli/pkg/reporter"
	"github.com/ipavlic/apex-benchmark-cli/pkg/stats"
	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	// Flags for compare command
	compareBenches []string
	compareIterations int
	compareWarmup     int
	compareRuns       int
	compareParallel   int
	compareTrackHeap  bool
	compareTrackDB    bool
	compareOrg        string
	compareOutput     string
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare multiple benchmarks",
	Long: `Compare multiple benchmarks side-by-side.
Use --bench flag multiple times to specify benchmarks.
Format: --bench "Name:code" or --bench "Name:path/to/file.apex"`,
	RunE: compareBenchmarks,
}

func init() {
	compareCmd.Flags().StringArrayVar(&compareBenches, "bench", []string{}, "Benchmark to compare (repeatable)")
	compareCmd.Flags().IntVar(&compareIterations, "iterations", 100, "Number of measurement iterations")
	compareCmd.Flags().IntVar(&compareWarmup, "warmup", 10, "Number of warmup iterations")
	compareCmd.Flags().IntVar(&compareRuns, "runs", 1, "Number of complete runs for aggregation")
	compareCmd.Flags().IntVar(&compareParallel, "parallel", 1, "Maximum concurrent executions")
	compareCmd.Flags().BoolVar(&compareTrackHeap, "track-heap", false, "Enable heap usage tracking")
	compareCmd.Flags().BoolVar(&compareTrackDB, "track-db", false, "Enable DML/SOQL tracking")
	compareCmd.Flags().StringVar(&compareOrg, "org", "", "Target Salesforce org (uses default if not specified)")
	compareCmd.Flags().StringVar(&compareOutput, "output", "table", "Output format: json, table")

	compareCmd.MarkFlagRequired("bench")
}

func compareBenchmarks(cmd *cobra.Command, args []string) error {
	// Validate flags
	if len(compareBenches) < 2 {
		return fmt.Errorf("must provide at least 2 benchmarks to compare")
	}

	// Check Salesforce CLI
	if err := executor.CheckSalesforceCLI(); err != nil {
		return err
	}

	// Get org
	org, err := executor.GetOrg(compareOrg)
	if err != nil {
		return err
	}
	if compareOrg == "" {
		fmt.Fprintf(os.Stderr, "Using default org: %s\n", org)
	}

	// Check org authentication
	if err := executor.CheckOrgAuth(org); err != nil {
		return err
	}

	// Parse benchmark specifications
	benchSpecs := make([]types.BenchmarkSpec, 0, len(compareBenches))
	for _, bench := range compareBenches {
		parts := strings.SplitN(bench, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid benchmark format %q, expected 'Name:code' or 'Name:file'", bench)
		}

		name := strings.TrimSpace(parts[0])
		source := strings.TrimSpace(parts[1])

		spec := types.BenchmarkSpec{
			Name: name,
		}

		// Check if source is a file (ends with .apex or exists as a file)
		if strings.HasSuffix(source, ".apex") || fileExists(source) {
			spec.File = source
		} else {
			spec.Code = source
		}

		benchSpecs = append(benchSpecs, spec)
	}

	// Create executor and run
	exec := executor.NewCLIExecutor()
	return compareBenchmarksWithExecutor(exec, org, benchSpecs, compareIterations, compareWarmup, compareRuns, compareParallel, compareTrackHeap, compareTrackDB, compareOutput)
}

// compareBenchmarksWithExecutor is the testable core logic
func compareBenchmarksWithExecutor(exec executor.Executor, org string, benchSpecs []types.BenchmarkSpec, iterations int, warmup int, runs int, parallel int, trackHeap bool, trackDB bool, outputFormat string) error {
	aggregatedResults := make([]types.AggregatedResult, 0, len(benchSpecs))

	for i, benchSpec := range benchSpecs {
		fmt.Fprintf(os.Stderr, "\n[%d/%d] Running benchmark: %s\n", i+1, len(benchSpecs), benchSpec.Name)

		// Read code
		userCode := benchSpec.Code
		if benchSpec.File != "" {
			content, err := os.ReadFile(benchSpec.File)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", benchSpec.File, err)
			}
			userCode = string(content)
		}

		// Build CodeSpec
		spec := types.CodeSpec{
			Name:       benchSpec.Name,
			UserCode:   strings.TrimSpace(userCode),
			Iterations: iterations,
			Warmup:     warmup,
			TrackHeap:  trackHeap,
			TrackDB:    trackDB,
		}

		// Generate
		apexCode, err := generator.Generate(spec)
		if err != nil {
			return fmt.Errorf("failed to generate code for %s: %w", benchSpec.Name, err)
		}

		// Execute
		var outputs []string
		if runs == 1 {
			output, err := exec.Run(apexCode, org)
			if err != nil {
				return fmt.Errorf("execution failed for %s: %w", benchSpec.Name, err)
			}
			outputs = []string{output}
		} else {
			var err error
			outputs, err = exec.ExecuteParallel(apexCode, runs, parallel, org)
			if err != nil {
				return fmt.Errorf("execution failed for %s: %w", benchSpec.Name, err)
			}
		}

		// Parse
		results, err := parser.ParseMultipleResults(outputs)
		if err != nil {
			return fmt.Errorf("failed to parse results for %s: %w", benchSpec.Name, err)
		}

		// Aggregate
		aggregated, err := stats.Aggregate(results)
		if err != nil {
			return fmt.Errorf("failed to aggregate results for %s: %w", benchSpec.Name, err)
		}
		aggregated.Warmup = warmup

		aggregatedResults = append(aggregatedResults, aggregated)
		fmt.Fprintf(os.Stderr, "  Completed: avg CPU %.3f ms\n", aggregated.AvgCpuMs)
	}

	// Output
	fmt.Fprintf(os.Stderr, "\n")
	switch outputFormat {
	case "json":
		return reporter.PrintJSON(aggregatedResults, os.Stdout)
	case "table":
		return reporter.PrintComparison(aggregatedResults, os.Stdout)
	default:
		return fmt.Errorf("unknown output format: %s", outputFormat)
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
