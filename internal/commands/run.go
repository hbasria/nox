package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"nox/internal/config"
	"nox/internal/llm"
	"nox/internal/memory"
	"nox/internal/prompts"
)

// RunNL turns a natural language request into a shell command, shows it,
// and runs it once confirmed. Before running, it checks that the command's
// binaries actually exist on this machine and, if not, offers to install
// them first.
func RunNL(cfg *config.Config, request string, auto, verbose bool) error {
	if strings.TrimSpace(request) == "" {
		return fmt.Errorf("empty request")
	}

	memCtx, err := memory.Load()
	if err != nil {
		return err
	}

	client, err := llm.New(cfg, verbose)
	if err != nil {
		return err
	}

	ctx := context.Background()

	raw, err := client.Complete(ctx, prompts.CommandGen(memCtx), request)
	if err != nil {
		return err
	}

	cmdStr, notes := extractMemoryNotes(raw)
	for _, note := range notes {
		if err := memory.Append(note); err != nil {
			fmt.Fprintf(os.Stderr, "nox: could not save memory note: %v\n", err)
			continue
		}
		fmt.Fprintf(os.Stderr, "nox: remembered: %s\n", note)
	}

	cmdStr = stripCodeFence(cmdStr)
	if cmdStr == "" {
		return fmt.Errorf("model returned an empty command")
	}

	for _, bin := range missingBinaries(cmdStr) {
		installed, err := offerInstall(ctx, client, memCtx, bin, auto)
		if err != nil {
			return err
		}
		if !installed {
			fmt.Printf("%q is still missing, not running the command.\n", bin)
			return nil
		}
	}

	return confirmAndRun(cmdStr, auto)
}

// offerInstall asks the model for a shell command that installs bin, shows
// it, and runs it once confirmed. It reports whether bin is available
// afterwards.
func offerInstall(ctx context.Context, client *llm.Client, memCtx, bin string, auto bool) (bool, error) {
	installCmd, err := client.Complete(ctx, prompts.InstallCmd(memCtx), fmt.Sprintf("Missing command-line tool: %s", bin))
	if err != nil {
		return false, err
	}
	installCmd = stripCodeFence(installCmd)
	if installCmd == "" {
		return false, fmt.Errorf("model returned no install command for %q", bin)
	}

	fmt.Printf("%q was not found on this system. Suggested install:\n", bin)
	if err := confirmAndRun(installCmd, auto); err != nil {
		return false, err
	}
	return binaryExists(bin), nil
}

// extractMemoryNotes pulls out any "MEMORY: <fact>" lines from the model's
// raw reply, returning the remaining text and the extracted notes in order.
func extractMemoryNotes(raw string) (string, []string) {
	lines := strings.Split(raw, "\n")
	kept := lines[:0:0]
	var notes []string
	for _, line := range lines {
		if note, ok := strings.CutPrefix(strings.TrimSpace(line), "MEMORY:"); ok {
			if note = strings.TrimSpace(note); note != "" {
				notes = append(notes, note)
			}
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n")), notes
}

// stripCodeFence removes a leading/trailing ``` fence in case the model
// ignores the "no markdown" instruction.
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) >= 2 {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
