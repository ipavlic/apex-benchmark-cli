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
	// Flags for run command
	runCode       string
	runFile       string
	runName       string
	runIterations int
	runWarmup     int
	runRuns       int
	runParallel   int
	runTrackHeap  bool
	runTrackDB    bool
	runOrg        string
	runOutput     string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a single benchmark",
	Long: `Run a benchmark on a single Apex code snippet.
You must provide either --code for inline code or --file for a code file.`,
	RunE: runBenchmark,
}

func init() {
	runCmd.Flags().StringVar(&runCode, "code", "", "Inline Apex code to benchmark")
	runCmd.Flags().StringVar(&runFile, "file", "", "Path to Apex code file")
	runCmd.Flags().StringVar(&runName, "name", "Benchmark", "Benchmark name")
	runCmd.Flags().IntVar(&runIterations, "iterations", 100, "Number of measurement iterations")
	runCmd.Flags().IntVar(&runWarmup, "warmup", 10, "Number of warmup iterations")
	runCmd.Flags().IntVar(&runRuns, "runs", 1, "Number of complete runs for aggregation")
	runCmd.Flags().IntVar(&runParallel, "parallel", 1, "Maximum concurrent executions")
	runCmd.Flags().BoolVar(&runTrackHeap, "track-heap", false, "Enable heap usage tracking")
	runCmd.Flags().BoolVar(&runTrackDB, "track-db", false, "Enable DML/SOQL tracking")
	runCmd.Flags().StringVar(&runOrg, "org", "", "Target Salesforce org (uses default if not specified)")
	runCmd.Flags().StringVar(&runOutput, "output", "json", "Output format: json, table")
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	// Validate flags
	if runCode == "" && runFile == "" {
		return fmt.Errorf("must provide either --code or --file")
	}
	if runCode != "" && runFile != "" {
		return fmt.Errorf("cannot provide both --code and --file")
	}

	// Check Salesforce CLI
	if err := executor.CheckSalesforceCLI(); err != nil {
		return err
	}

	// Get org
	org, err := executor.GetOrg(runOrg)
	if err != nil {
		return err
	}
	if runOrg == "" {
		fmt.Fprintf(os.Stderr, "Using default org: %s\n", org)
	}

	// Check org authentication
	if err := executor.CheckOrgAuth(org); err != nil {
		return err
	}

	// Read code from file if needed
	userCode := runCode
	if runFile != "" {
		content, err := os.ReadFile(runFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", runFile, err)
		}
		userCode = string(content)
	}

	// Build CodeSpec
	spec := types.CodeSpec{
		Name:       runName,
		UserCode:   strings.TrimSpace(userCode),
		Iterations: runIterations,
		Warmup:     runWarmup,
		TrackHeap:  runTrackHeap,
		TrackDB:    runTrackDB,
	}

	// Generate Apex code
	fmt.Fprintf(os.Stderr, "Generating benchmark code...\n")
	apexCode, err := generator.Generate(spec)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Execute
	exec := executor.NewCLIExecutor()
	var outputs []string

	if runRuns == 1 {
		fmt.Fprintf(os.Stderr, "Executing benchmark (1 run)...\n")
		output, err := exec.Run(apexCode, org)
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
		outputs = []string{output}
	} else {
		fmt.Fprintf(os.Stderr, "Executing benchmark (%d runs, %d parallel)...\n", runRuns, runParallel)
		var err error
		outputs, err = exec.ExecuteParallel(apexCode, runRuns, runParallel, org)
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
	}

	// Parse results
	fmt.Fprintf(os.Stderr, "Parsing results...\n")
	results, err := parser.ParseMultipleResults(outputs)
	if err != nil {
		return fmt.Errorf("failed to parse results: %w", err)
	}

	// Aggregate
	fmt.Fprintf(os.Stderr, "Aggregating results...\n")
	aggregated, err := stats.Aggregate(results)
	if err != nil {
		return fmt.Errorf("failed to aggregate results: %w", err)
	}
	aggregated.Warmup = runWarmup

	// Output
	fmt.Fprintf(os.Stderr, "\n")
	switch runOutput {
	case "json":
		return reporter.PrintJSON(aggregated, os.Stdout)
	case "table":
		return reporter.PrintTable(aggregated, os.Stdout)
	default:
		return fmt.Errorf("unknown output format: %s", runOutput)
	}
}
