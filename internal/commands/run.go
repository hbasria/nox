package commands

import (
	"context"
	"fmt"
	"strings"

	"nox/internal/config"
	"nox/internal/llm"
	"nox/internal/prompts"
)

// RunNL turns a natural language request into a shell command, shows it,
// and runs it once confirmed.
func RunNL(cfg *config.Config, request string, auto bool) error {
	if strings.TrimSpace(request) == "" {
		return fmt.Errorf("boş istek")
	}

	client, err := llm.New(cfg)
	if err != nil {
		return err
	}

	cmdStr, err := client.Complete(context.Background(), prompts.CommandGen(), request)
	if err != nil {
		return err
	}
	cmdStr = stripCodeFence(cmdStr)
	if cmdStr == "" {
		return fmt.Errorf("model boş komut döndürdü")
	}

	return confirmAndRun(cmdStr, auto)
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
