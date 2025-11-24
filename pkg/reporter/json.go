package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// PrintJSON outputs the result as formatted JSON
func PrintJSON(result interface{}, writer io.Writer) error {
	if writer == nil {
		writer = os.Stdout
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
