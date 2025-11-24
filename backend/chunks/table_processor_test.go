package chunks

import (
	"strings"
	"testing"
)

func TestTableProcessor(t *testing.T) {
	tp := NewTableProcessor()

	text := `
Here is some intro text.

# Comparison of Wi-Fi Standards 5 6 7
Feature:
Wi-Fi 5
Wi-Fi 6
Wi-Fi 7
Speed
3.5 Gbps
9.6 Gbps
46 Gbps
Latency
High
Low
Very Low

And some text after.
`

	segments := tp.Process(text)

	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(segments))
	}

	// Segment 0: Intro text
	if segments[0].IsTable {
		t.Errorf("Expected segment 0 to be text")
	}
	if !strings.Contains(segments[0].Text, "intro text") {
		t.Errorf("Segment 0 content mismatch")
	}

	// Segment 1: Table
	if !segments[1].IsTable {
		t.Errorf("Expected segment 1 to be table")
	}
	if !strings.Contains(segments[1].Text, "COMPARISON TABLE: Comparison of Wi-Fi Standards") {
		t.Errorf("Table title not found in processed text. Got:\n%s", segments[1].Text)
	}
	if !strings.Contains(segments[1].Text, "Wi-Fi 5: 3.5 Gbps") {
		t.Errorf("Table content mismatch")
	}

	// Segment 2: Outro text
	if segments[2].IsTable {
		t.Errorf("Expected segment 2 to be text")
	}
	if !strings.Contains(segments[2].Text, "text after") {
		t.Errorf("Segment 2 content mismatch")
	}
}
