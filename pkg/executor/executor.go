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

// Run executes Apex code once and returns the debug log output
func (e *CLIExecutor) Run(apexCode string, org string) (string, error) {
	// Create temp file
	tempFile, err := createTempApexFile(apexCode)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile)

	// Build sf command
	args := []string{"apex", "run", "--file", tempFile}
	if org != "" {
		args = append(args, "--target-org", org)
	}

	// Execute command
	cmd := exec.Command("sf", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("sf apex run failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
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
	cmd := exec.Command("sf", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sf CLI not found or not working: %w\nPlease install Salesforce CLI: https://developer.salesforce.com/tools/salesforcecli", err)
	}

	if !strings.Contains(string(output), "@salesforce/cli") {
		return fmt.Errorf("unexpected sf CLI output: %s", string(output))
	}

	return nil
}

// GetDefaultOrg returns the default Salesforce org alias/username
func GetDefaultOrg() (string, error) {
	cmd := exec.Command("sf", "config", "get", "target-org", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get default org: %w", err)
	}

	var result struct {
		Status int `json:"status"`
		Result []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"result"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse config output: %w", err)
	}

	if len(result.Result) == 0 || result.Result[0].Value == "" || result.Result[0].Value == "null" {
		return "", fmt.Errorf("no default org configured. Run: sf org login web")
	}

	return result.Result[0].Value, nil
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

// CheckOrgAuth verifies that an org is authenticated
func CheckOrgAuth(org string) error {
	args := []string{"org", "display"}
	if org != "" {
		args = append(args, "--target-org", org)
	}

	cmd := exec.Command("sf", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("org %q is not authenticated: %w\nOutput: %s", org, err, string(output))
	}

	return nil
}
