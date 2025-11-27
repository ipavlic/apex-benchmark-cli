package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/ipavlic/apex-benchmark-cli/pkg/types"
)

// templateData extends CodeSpec with additional template variables
type templateData struct {
	types.CodeSpec
	LoopVar string
}

// Generate creates Apex code from a CodeSpec using the template
func Generate(spec types.CodeSpec) (string, error) {
	// Validate input
	if err := validateSpec(spec); err != nil {
		return "", err
	}

	// Generate unique loop variable name to avoid conflicts with user code
	loopVar := "i_" + strings.ReplaceAll(uuid.New().String(), "-", "_")

	// Parse template
	tmpl, err := template.New("apex").Parse(apexTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	data := templateData{
		CodeSpec: spec,
		LoopVar:  loopVar,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
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
