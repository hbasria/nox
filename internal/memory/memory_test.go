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

	if written, err := Append("ss is not available on macOS"); err != nil || !written {
		t.Fatalf("Append: written=%v err=%v", written, err)
	}
	if written, err := Append("user prefers verbose output"); err != nil || !written {
		t.Fatalf("Append: written=%v err=%v", written, err)
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

func TestAppendSkipsDuplicates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.md")
	t.Setenv("NOX_MEMORY", path)

	if written, err := Append("ss is not available on macOS"); err != nil || !written {
		t.Fatalf("first Append: written=%v err=%v", written, err)
	}
	written, err := Append("ss is not available on macOS")
	if err != nil {
		t.Fatalf("second Append: %v", err)
	}
	if written {
		t.Fatal("expected duplicate note to be skipped, but it was written")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if n := strings.Count(string(data), "ss is not available on macOS"); n != 1 {
		t.Errorf("expected note to appear exactly once, appeared %d times:\n%s", n, data)
	}
}
