package commands

import (
	"reflect"
	"testing"
)

func TestParseCleanupResponse(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		wantCmd       string
		wantExplain   []string
		wantNoCleanup bool
		wantErr       bool
	}{
		{
			name:        "command with explanations",
			raw:         "rm -rf __pycache__ node_modules .DS_Store\n\n- __pycache__: Python compiled bytecode files.\n- node_modules: NodeJS dependencies.\n- .DS_Store: macOS metadata file.",
			wantCmd:     "rm -rf __pycache__ node_modules .DS_Store",
			wantExplain: []string{"- __pycache__: Python compiled bytecode files.", "- node_modules: NodeJS dependencies.", "- .DS_Store: macOS metadata file."},
		},
		{
			name:          "no cleanup needed",
			raw:           "echo 'No cleanup needed'",
			wantCmd:       "echo 'No cleanup needed'",
			wantNoCleanup: true,
		},
		{
			name:        "code fence is stripped",
			raw:         "```\nrm -rf dist\n- dist: build output\n```",
			wantCmd:     "rm -rf dist",
			wantExplain: []string{"- dist: build output"},
		},
		{
			name:    "empty response is an error",
			raw:     "",
			wantErr: true,
		},
		{
			name:    "whitespace-only response is an error",
			raw:     "   \n\n  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, explain, noCleanup, err := parseCleanupResponse(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if noCleanup != tt.wantNoCleanup {
				t.Errorf("noCleanup = %v, want %v", noCleanup, tt.wantNoCleanup)
			}
			if !reflect.DeepEqual(explain, tt.wantExplain) {
				t.Errorf("explanations = %v, want %v", explain, tt.wantExplain)
			}
		})
	}
}
