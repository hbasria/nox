package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"nox/internal/config"
	"nox/internal/llm"
	"nox/internal/prompts"
)

// maxDiffChars caps how much of the staged diff gets sent to the model.
const maxDiffChars = 8000

// Commit reads the staged git diff, asks the model for a commit message,
// shows it, and runs `git commit -m` once confirmed.
func Commit(cfg *config.Config, auto, verbose bool) error {
	diff, err := stagedDiff()
	if err != nil {
		return err
	}
	if strings.TrimSpace(diff) == "" {
		fmt.Println("no staged changes (use git add to stage them).")
		return nil
	}
	if len(diff) > maxDiffChars {
		diff = diff[:maxDiffChars] + "\n... (diff truncated)"
	}

	client, err := llm.New(cfg, verbose)
	if err != nil {
		return err
	}

	msg, err := client.Complete(context.Background(), prompts.CommitMsg(), diff)
	if err != nil {
		return err
	}
	msg = strings.Trim(strings.TrimSpace(msg), "`\"")
	if msg == "" {
		return fmt.Errorf("model returned an empty commit message")
	}

	fmt.Println("Proposed commit message:")
	fmt.Println("---")
	fmt.Println(msg)
	fmt.Println("---")

	if !auto {
		fmt.Print("[Enter] commit  [Ctrl+C] cancel ")
		reader := bufio.NewReader(os.Stdin)
		if _, err := reader.ReadString('\n'); err != nil {
			return nil
		}
	}

	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func stagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not run git diff --staged: %w", err)
	}
	return string(out), nil
}
