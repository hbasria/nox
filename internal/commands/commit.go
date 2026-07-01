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
func Commit(cfg *config.Config, auto bool) error {
	diff, err := stagedDiff()
	if err != nil {
		return err
	}
	if strings.TrimSpace(diff) == "" {
		fmt.Println("staged edilmiş değişiklik yok (git add ile stage edin).")
		return nil
	}
	if len(diff) > maxDiffChars {
		diff = diff[:maxDiffChars] + "\n... (diff kısaltıldı)"
	}

	client, err := llm.New(cfg)
	if err != nil {
		return err
	}

	msg, err := client.Complete(context.Background(), prompts.CommitMsg(), diff)
	if err != nil {
		return err
	}
	msg = strings.Trim(strings.TrimSpace(msg), "`\"")
	if msg == "" {
		return fmt.Errorf("model boş commit mesajı döndürdü")
	}

	fmt.Println("Önerilen commit mesajı:")
	fmt.Println("---")
	fmt.Println(msg)
	fmt.Println("---")

	if !auto {
		fmt.Print("[Enter] commit et  [Ctrl+C] iptal ")
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
		return "", fmt.Errorf("git diff --staged çalıştırılamadı: %w", err)
	}
	return string(out), nil
}
