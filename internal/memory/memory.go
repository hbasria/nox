// Package memory manages nox's persistent, human-editable memory file
// (~/.nox/memory.md). It is plain text: nox only ever reads it whole and
// feeds it to the model as context, so there is no schema to keep in sync.
// Users are free to hand-edit it and add their own notes.
package memory

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"nox/internal/config"
)

// Path returns the memory file path, honoring NOX_MEMORY for overrides.
func Path() (string, error) {
	if p := os.Getenv("NOX_MEMORY"); p != "" {
		return p, nil
	}
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "memory.md"), nil
}

// Load returns the memory file's contents, creating it with detected
// system facts on first run.
func Load() (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", fmt.Errorf("could not create memory directory: %w", err)
		}
		if err := os.WriteFile(path, []byte(detect()), 0o644); err != nil {
			return "", fmt.Errorf("could not write memory file: %w", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read memory file: %w", err)
	}
	return string(data), nil
}

// Append adds a new fact/note as a bullet line, creating the file with its
// usual starter content first if it doesn't exist yet. It reports whether
// the note was actually written (false if an identical note is already
// present, in which case it's silently skipped).
func Append(note string) (bool, error) {
	note = strings.TrimSpace(note)

	existing, err := Load()
	if err != nil {
		return false, err
	}
	for line := range strings.SplitSeq(existing, "\n") {
		if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-")) == note {
			return false, nil
		}
	}

	path, err := Path()
	if err != nil {
		return false, err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return false, fmt.Errorf("could not open memory file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("- " + note + "\n"); err != nil {
		return false, fmt.Errorf("could not write to memory file: %w", err)
	}
	return true, nil
}

// detect builds the starter memory content from the current system.
func detect() string {
	var b strings.Builder
	b.WriteString("# nox memory\n\n")
	b.WriteString("Facts nox knows about this machine. Edit freely; nox reads this file as-is.\n\n")
	fmt.Fprintf(&b, "- OS: %s (%s)\n", osName(), runtime.GOOS)
	fmt.Fprintf(&b, "- Architecture: %s\n", runtime.GOARCH)
	fmt.Fprintf(&b, "- Shell: %s\n", shellName())
	if pm := packageManager(); pm != "" {
		fmt.Fprintf(&b, "- Package manager: %s\n", pm)
	}
	return b.String()
}

func osName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}

func shellName() string {
	if s := os.Getenv("SHELL"); s != "" {
		return filepath.Base(s)
	}
	return "unknown"
}

// packageManager returns the first known package manager found on PATH.
func packageManager() string {
	candidates := []string{"brew", "apt-get", "dnf", "yum", "pacman", "apk", "zypper"}
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}
	return ""
}
