package commands

import (
	"reflect"
	"testing"
)

func TestMissingBinaries(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want []string
	}{
		{"present binary", "lsof -i -P", nil},
		{"missing binary", "ss -tuln", []string{"ss"}},
		{"chained segments", "lsof -i -P && ss -tuln", []string{"ss"}},
		{"env assignment skipped", "FOO=bar lsof -i", nil},
		{"sudo checks real command", "sudo ss -tuln", []string{"ss"}},
		{"dedup across segments", "ss -tuln; ss -a", []string{"ss"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := missingBinaries(tt.cmd)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("missingBinaries(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}
