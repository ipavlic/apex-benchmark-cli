package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "apex-bench",
	Short: "Apex Benchmark CLI - Benchmark Salesforce Apex code snippets",
	Long: `apex-bench is a CLI tool for benchmarking Salesforce Apex code snippets
without deployment. It wraps your code in measurement logic and executes
it via the Salesforce CLI.`,
	Version: version,
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(compareCmd)
}
