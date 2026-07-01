package commands

import (
	"context"
	"fmt"
	"strings"

	"nox/internal/config"
	"nox/internal/llm"
	"nox/internal/prompts"
)

// Explain asks the model to break down a shell command in plain language,
// without running it.
func Explain(cfg *config.Config, cmdStr string, verbose bool) error {
	if strings.TrimSpace(cmdStr) == "" {
		return fmt.Errorf("empty command")
	}

	client, err := llm.New(cfg, verbose)
	if err != nil {
		return err
	}

	explanation, err := client.Complete(context.Background(), prompts.Explain(), cmdStr)
	if err != nil {
		return err
	}

	fmt.Println(explanation)
	return nil
}
