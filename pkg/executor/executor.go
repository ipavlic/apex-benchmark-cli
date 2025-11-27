package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

// execCommand is a variable that points to exec.Command
// This allows us to mock it in tests
var execCommand = exec.Command

// Executor interface allows for mocking in tests
type Executor interface {
	Run(apexCode string, org string) (string, error)
	ExecuteParallel(apexCode string, runs int, maxConcurrent int, org string) ([]string, error)
}

// CLIExecutor implements Executor using the Salesforce CLI
type CLIExecutor struct{}

// NewCLIExecutor creates a new executor that uses sf CLI
func NewCLIExecutor() *CLIExecutor {
	return &CLIExecutor{}
}

// ApexRunResponse represents the JSON response from `sf apex run --json`
// Reference: https://developer.salesforce.com/docs/atlas.en-us.sfdx_cli_reference.meta/sfdx_cli_reference/cli_reference_apex_commands_unified.htm
//
// Expected JSON structure:
//
//	{
//	  "status": 0,
//	  "result": {
//	    "success": true,
//	    "compiled": true,
//	    "compileProblem": "",
//	    "exceptionMessage": "",
//	    "exceptionStackTrace": "",
//	    "line": -1,
//	    "column": -1,
//	    "logs": "...debug log output..."
//	  }
//	}
//
// On error, compileProblem or exceptionMessage will contain error details.
type ApexRunResponse struct {
	Status int `json:"status"`
	Result struct {
		Success             bool   `json:"success"`
		Compiled            bool   `json:"compiled"`
		CompileProblem      string `json:"compileProblem"`
		ExceptionMessage    string `json:"exceptionMessage"`
		ExceptionStackTrace string `json:"exceptionStackTrace"`
		Line                int    `json:"line"`
		Column              int    `json:"column"`
		Logs                string `json:"logs"`
	} `json:"result"`
}

// Run executes Apex code once and returns the debug log output
func (e *CLIExecutor) Run(apexCode string, org string) (string, error) {
	// Create temp file
	tempFile, err := createTempApexFile(apexCode)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile)

	// Build sf command with --json flag for structured output
	args := []string{"apex", "run", "--file", tempFile, "--json"}
	if org != "" {
		args = append(args, "--target-org", org)
	}

	// Execute command
	cmd := execCommand("sf", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("sf apex run failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON response
	var response ApexRunResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return "", fmt.Errorf("failed to parse sf apex run JSON output: %w\nOutput: %s", err, string(output))
	}

	// Check if execution was successful
	if !response.Result.Success {
		if !response.Result.Compiled {
			return "", fmt.Errorf("Apex compilation failed: %s", response.Result.CompileProblem)
		}
		return "", fmt.Errorf("Apex execution failed: %s", response.Result.ExceptionMessage)
	}

	// Return the logs which contain our BENCH_RESULT output
	return response.Result.Logs, nil
}

// ExecuteParallel runs the same Apex code multiple times in parallel
func (e *CLIExecutor) ExecuteParallel(apexCode string, runs int, maxConcurrent int, org string) ([]string, error) {
	if runs <= 0 {
		return nil, fmt.Errorf("runs must be positive, got %d", runs)
	}
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}

	// Create semaphore for rate limiting
	sem := semaphore.NewWeighted(int64(maxConcurrent))
	ctx := context.Background()

	results := make([]string, runs)
	errors := make([]error, runs)
	var wg sync.WaitGroup

	for i := 0; i < runs; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Acquire semaphore
			if err := sem.Acquire(ctx, 1); err != nil {
				errors[index] = fmt.Errorf("failed to acquire semaphore: %w", err)
				return
			}
			defer sem.Release(1)

			// Execute
			output, err := e.Run(apexCode, org)
			if err != nil {
				errors[index] = err
				return
			}
			results[index] = output
		}(i)
	}

	wg.Wait()

	// Check for errors
	var errorMessages []string
	for i, err := range errors {
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("run %d: %v", i+1, err))
		}
	}
	if len(errorMessages) > 0 {
		return nil, fmt.Errorf("execution errors:\n%s", strings.Join(errorMessages, "\n"))
	}

	return results, nil
}

// createTempApexFile writes Apex code to a temporary file
func createTempApexFile(apexCode string) (string, error) {
	tmpFile, err := os.CreateTemp("", "apex-bench-*.apex")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write([]byte(apexCode)); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// CheckSalesforceCLI verifies that sf CLI is installed
func CheckSalesforceCLI() error {
	cmd := execCommand("sf", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sf CLI not found or not working: %w\nPlease install Salesforce CLI: https://developer.salesforce.com/tools/salesforcecli", err)
	}

	if !strings.Contains(string(output), "@salesforce/cli") {
		return fmt.Errorf("unexpected sf CLI output: %s", string(output))
	}

	return nil
}

// ConfigGetResponse represents the JSON response from `sf config get --json`
// Reference: https://developer.salesforce.com/docs/atlas.en-us.sfdx_setup.meta/sfdx_setup/sfdx_dev_cli_json_support.htm
//
// Expected JSON structure:
//
//	{
//	  "status": 0,
//	  "result": [
//	    {
//	      "name": "target-org",
//	      "value": "my-org-alias",
//	      "location": "Local"
//	    }
//	  ]
//	}
//
// If no config is set, result array will be empty or value will be empty/null.
type ConfigGetResponse struct {
	Status int `json:"status"`
	Result []struct {
		Name     string `json:"name"`
		Value    string `json:"value"`
		Location string `json:"location,omitempty"`
	} `json:"result"`
}

// GetDefaultOrg returns the default Salesforce org alias/username
func GetDefaultOrg() (string, error) {
	cmd := execCommand("sf", "config", "get", "target-org", "--json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get default org: %w", err)
	}

	var response ConfigGetResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return "", fmt.Errorf("failed to parse config output: %w", err)
	}

	if len(response.Result) == 0 || response.Result[0].Value == "" || response.Result[0].Value == "null" {
		return "", fmt.Errorf("no default org configured. Run: sf org login web")
	}

	return response.Result[0].Value, nil
}

// GetOrg returns the specified org or the default org if none specified
func GetOrg(specified string) (string, error) {
	if specified != "" {
		return specified, nil
	}

	org, err := GetDefaultOrg()
	if err != nil {
		return "", fmt.Errorf("no org specified and could not get default org: %w", err)
	}

	return org, nil
}
