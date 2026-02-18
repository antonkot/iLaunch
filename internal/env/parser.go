package env

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Entry struct {
	Key     string
	Default string
}

func ParseExample(r io.Reader) ([]Entry, error) {
	scanner := bufio.NewScanner(r)
	entries := make([]Entry, 0)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid line %d: expected KEY=value", lineNum)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("invalid line %d: empty key", lineNum)
		}
		entries = append(entries, Entry{Key: key, Default: value})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan .env.example: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	return entries, nil
}
