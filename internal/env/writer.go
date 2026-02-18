package env

import (
	"fmt"
	"os"
	"sort"
)

func WriteFile(path string, values map[string]string) error {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]byte, 0, len(keys)*16)
	for _, k := range keys {
		if values[k] == "" {
			return fmt.Errorf("empty value for key %s", k)
		}
		lines = append(lines, []byte(fmt.Sprintf("%s=%s\n", k, values[k]))...)
	}
	if err := os.WriteFile(path, lines, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
