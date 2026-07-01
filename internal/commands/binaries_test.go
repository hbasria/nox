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
		{
			"awk script with internal semicolons and pipes is not split",
			`ps -eo rss,comm | sort -k1,1nr | head -n 10 | awk '{printf "%8.1f KB  ", $2; $2=""; print substr($0,3)}'`,
			nil,
		},
		{
			"double-quoted semicolon is not split",
			`echo "a; b" | grep a`,
			nil,
		},
		{
			"garbage token from a mis-split is never treated as a binary",
			`awk 'BEGIN { x="a" ; y="b" }'`,
			nil,
		},
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
