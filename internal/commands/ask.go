package commands

import (
	"context"
	"fmt"
	"strings"

	"nox/internal/config"
	"nox/internal/llm"
	"nox/internal/prompts"
)

// Ask answers a plain question, optionally using piped input as context.
// Unlike RunNL, this never generates or runs a command.
func Ask(cfg *config.Config, question string, verbose bool) error {
	if strings.TrimSpace(question) == "" {
		return fmt.Errorf("empty question")
	}

	pipedInput, piped := readPipedInput()

	client, err := llm.New(cfg, verbose)
	if err != nil {
		return err
	}

	userPrompt := question
	if piped && strings.TrimSpace(pipedInput) != "" {
		userPrompt = fmt.Sprintf("Piped input (from a previous command):\n%s\n\nQuestion: %s", pipedInput, question)
	}

	answer, err := client.Complete(context.Background(), prompts.Ask(), userPrompt)
	if err != nil {
		return err
	}

	fmt.Println(answer)
	return nil
}
