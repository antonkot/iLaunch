package env

import (
	"strings"
	"testing"
)

func TestParseExample(t *testing.T) {
	input := strings.NewReader(`# comment
PORT=3000
API_URL=https://example.com

`)
	entries, err := ParseExample(input)
	if err != nil {
		t.Fatalf("ParseExample() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "API_URL" || entries[0].Default != "https://example.com" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
}

func TestParseExampleInvalidLine(t *testing.T) {
	_, err := ParseExample(strings.NewReader("INVALID"))
	if err == nil {
		t.Fatal("expected error for invalid line")
	}
}
