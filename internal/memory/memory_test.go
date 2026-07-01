package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendCreatesAndAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.md")
	t.Setenv("NOX_MEMORY", path)

	if err := Append("ss is not available on macOS"); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := Append("user prefers verbose output"); err != nil {
		t.Fatalf("Append: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "# nox memory") {
		t.Errorf("expected starter header, got:\n%s", content)
	}
	if !strings.Contains(content, "- ss is not available on macOS") {
		t.Errorf("missing first note, got:\n%s", content)
	}
	if !strings.Contains(content, "- user prefers verbose output") {
		t.Errorf("missing second note, got:\n%s", content)
	}
}
