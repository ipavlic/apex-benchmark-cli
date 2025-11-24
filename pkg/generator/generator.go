package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

// Generate creates Apex code from a CodeSpec using the template
func Generate(spec types.CodeSpec) (string, error) {
	// Validate input
	if err := validateSpec(spec); err != nil {
		return "", err
	}

	// Parse template
	tmpl, err := template.New("apex").Parse(apexTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, spec); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// validateSpec ensures the CodeSpec has valid values
func validateSpec(spec types.CodeSpec) error {
	if strings.TrimSpace(spec.UserCode) == "" {
		return fmt.Errorf("user code cannot be empty")
	}

	if spec.Iterations <= 0 {
		return fmt.Errorf("iterations must be positive, got %d", spec.Iterations)
	}

	if spec.Warmup < 0 {
		return fmt.Errorf("warmup cannot be negative, got %d", spec.Warmup)
	}

	if strings.TrimSpace(spec.Name) == "" {
		return fmt.Errorf("benchmark name cannot be empty")
	}

	return nil
}
