package reporter

import (
	"fmt"
	"io"
	"os"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
	"github.com/olekukonko/tablewriter"
)

// PrintTable outputs a single result as a formatted table
func PrintTable(result types.AggregatedResult, writer io.Writer) error {
	if writer == nil {
		writer = os.Stdout
	}

	table := tablewriter.NewWriter(writer)
	table.Header("Name", "Avg CPU", "Min CPU", "Max CPU", "Std Dev")

	err := table.Append([]string{
		result.Name,
		fmt.Sprintf("%.3f ms", result.AvgCpuMs),
		fmt.Sprintf("%.3f ms", result.MinCpuMs),
		fmt.Sprintf("%.3f ms", result.MaxCpuMs),
		fmt.Sprintf("%.3f ms", result.StdDevCpuMs),
	})
	if err != nil {
		return fmt.Errorf("failed to append row: %w", err)
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

// PrintComparison outputs multiple results as a comparison table
func PrintComparison(results []types.AggregatedResult, writer io.Writer) error {
	if writer == nil {
		writer = os.Stdout
	}

	if len(results) == 0 {
		return fmt.Errorf("no results to display")
	}

	// Find the fastest (lowest avg CPU time)
	fastestIdx := 0
	fastestCpu := results[0].AvgCpuMs
	for i, r := range results {
		if r.AvgCpuMs < fastestCpu {
			fastestCpu = r.AvgCpuMs
			fastestIdx = i
		}
	}

	table := tablewriter.NewWriter(writer)
	table.Header("Name", "Avg CPU", "Min CPU", "Max CPU", "Relative")

	for i, result := range results {
		relative := result.AvgCpuMs / fastestCpu
		relativeStr := fmt.Sprintf("%.2fx", relative)

		if i == fastestIdx {
			relativeStr = "1.00x ⭐"
		}

		err := table.Append([]string{
			result.Name,
			fmt.Sprintf("%.3f ms", result.AvgCpuMs),
			fmt.Sprintf("%.3f ms", result.MinCpuMs),
			fmt.Sprintf("%.3f ms", result.MaxCpuMs),
			relativeStr,
		})
		if err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	// Print fastest
	fmt.Fprintf(writer, "\nFastest: %s\n", results[fastestIdx].Name)

	return nil
}
