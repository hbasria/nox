package commands

import (
	"reflect"
	"testing"
)

func TestExtractMemoryNotes(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantCmd   string
		wantNotes []string
	}{
		{
			name:      "no memory line",
			raw:       "lsof -i -P | grep LISTEN",
			wantCmd:   "lsof -i -P | grep LISTEN",
			wantNotes: nil,
		},
		{
			name:      "single memory line",
			raw:       "lsof -i -P | grep LISTEN\nMEMORY: user asked to list open ports",
			wantCmd:   "lsof -i -P | grep LISTEN",
			wantNotes: []string{"user asked to list open ports"},
		},
		{
			name:      "multiple memory lines",
			raw:       "ss -tuln\nMEMORY: ss is not available on macOS\nMEMORY: user prefers verbose output",
			wantCmd:   "ss -tuln",
			wantNotes: []string{"ss is not available on macOS", "user prefers verbose output"},
		},
		{
			name:      "empty memory line ignored",
			raw:       "ls -la\nMEMORY:   ",
			wantCmd:   "ls -la",
			wantNotes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotNotes := extractMemoryNotes(tt.raw)
			if gotCmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotNotes, tt.wantNotes) {
				t.Errorf("notes = %v, want %v", gotNotes, tt.wantNotes)
			}
		})
	}
}
