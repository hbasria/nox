package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nox/internal/config"
	"nox/internal/llm"
	"nox/internal/memory"
	"nox/internal/prompts"
)

// Cleanup walks the current directory to format the file/folder structure,
// asks the model to identify cleanup candidates and generate a shell command,
// prints the explanations, and runs the command after confirmation.
func Cleanup(cfg *config.Config, auto, verbose bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}

	structure, err := scanProjectStructure(cwd)
	if err != nil {
		return fmt.Errorf("could not scan project directory: %w", err)
	}

	memCtx, err := memory.Load()
	if err != nil {
		return err
	}

	client, err := llm.New(cfg, verbose)
	if err != nil {
		return err
	}

	raw, err := client.Complete(context.Background(), prompts.Cleanup(structure, memCtx), "Clean up this project")
	if err != nil {
		return err
	}

	cmdStr, explanations, noCleanup, err := parseCleanupResponse(raw)
	if err != nil {
		return err
	}
	if noCleanup {
		fmt.Println("No cleanup candidates identified.")
		return nil
	}

	if len(explanations) > 0 {
		fmt.Println("Proposed cleanup:")
		for _, exp := range explanations {
			fmt.Println(exp)
		}
		fmt.Println()
	}

	return confirmAndRun(cmdStr, auto, nil)
}

// parseCleanupResponse splits the model's raw reply into the cleanup
// command (first line) and its explanation lines (the rest). noCleanup
// reports whether the model decided nothing needs cleaning up.
func parseCleanupResponse(raw string) (cmdStr string, explanations []string, noCleanup bool, err error) {
	raw = stripCodeFence(raw)
	lines := strings.Split(raw, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return "", nil, false, fmt.Errorf("model returned an empty command")
	}

	cmdStr = strings.TrimSpace(lines[0])
	if strings.Contains(cmdStr, "No cleanup needed") {
		return cmdStr, nil, true, nil
	}

	for _, line := range lines[1:] {
		if line = strings.TrimSpace(line); line != "" {
			explanations = append(explanations, line)
		}
	}
	return cmdStr, explanations, false, nil
}

// scanProjectStructure walks the directory up to depth 2, returning a tree-like string representation.
func scanProjectStructure(root string) (string, error) {
	var sb strings.Builder

	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		name := entry.Name()
		// Skip version control and common heavy/hidden editor directories
		if name == ".git" || name == ".hg" || name == ".svn" || name == ".idea" || name == ".vscode" || name == ".github" {
			continue
		}

		if entry.IsDir() {
			sb.WriteString(name + "/\n")
			subEntries, err := os.ReadDir(filepath.Join(root, name))
			if err == nil {
				// Cap sub-entries per directory to avoid bloating the prompt.
				limit := min(len(subEntries), 15)
				for _, subEntry := range subEntries[:limit] {
					subName := subEntry.Name()
					if subName == ".git" || subName == ".hg" || subName == ".svn" {
						continue
					}
					if subEntry.IsDir() {
						sb.WriteString("  " + subName + "/\n")
					} else {
						sb.WriteString("  " + subName + "\n")
					}
				}
				if len(subEntries) > 15 {
					sb.WriteString("  ...\n")
				}
			}
		} else {
			sb.WriteString(name + "\n")
		}
	}

	return sb.String(), nil
}
