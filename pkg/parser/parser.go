package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

// ParseResult extracts the benchmark result from sf apex run output
func ParseResult(debugOutput string) (types.Result, error) {
	// Look for the BENCH_RESULT marker in the output
	// The generated Apex code outputs: System.debug('BENCH_RESULT:' + resultJson);
	// sf apex run output includes this as: USER_DEBUG|...|BENCH_RESULT:{json}

	// Find all occurrences of BENCH_RESULT: and try to parse JSON from each
	marker := "BENCH_RESULT:"
	searchPos := 0

	for {
		markerIdx := strings.Index(debugOutput[searchPos:], marker)
		if markerIdx == -1 {
			break
		}

		markerIdx += searchPos
		jsonStart := markerIdx + len(marker)
		remaining := debugOutput[jsonStart:]

		// Find matching brace using a brace counter for robust parsing
		braceCount := 0
		jsonEnd := -1
		for i, ch := range remaining {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					jsonEnd = i + 1
					break
				}
			}
		}

		if jsonEnd != -1 {
			jsonStr := remaining[:jsonEnd]

			// Try to parse JSON
			var result types.Result
			if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
				// Successfully parsed!
				return result, nil
			}
		}

		// Move to next occurrence
		searchPos = markerIdx + len(marker)
	}

	return types.Result{}, fmt.Errorf("could not find valid BENCH_RESULT JSON in output.\n\nOutput:\n%s", debugOutput)
}

// ParseMultipleResults parses results from multiple executions
func ParseMultipleResults(outputs []string) ([]types.Result, error) {
	results := make([]types.Result, len(outputs))
	var errors []string

	for i, output := range outputs {
		result, err := ParseResult(output)
		if err != nil {
			errors = append(errors, fmt.Sprintf("output %d: %v", i+1, err))
			continue
		}
		results[i] = result
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to parse some results:\n%s", strings.Join(errors, "\n"))
	}

	return results, nil
}

// ExtractDebugLines extracts just the debug log lines from sf output (utility function)
func ExtractDebugLines(output string) []string {
	var debugLines []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// sf apex run typically prefixes debug lines with timestamps and log levels
		// Look for lines that contain USER_DEBUG or similar
		if strings.Contains(line, "USER_DEBUG") || strings.Contains(line, "BENCH_RESULT") {
			debugLines = append(debugLines, line)
		}
	}

	return debugLines
}
