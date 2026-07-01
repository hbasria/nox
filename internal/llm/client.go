// Package llm implements a minimal OpenAI-compatible chat completions client.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"nox/internal/config"
)

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type request struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	// Think disables reasoning/"thinking" output on models that support the
	// toggle (e.g. served via Ollama). Providers that don't recognize this
	// field simply ignore it.
	Think bool `json:"think"`
}

type response struct {
	Choices []struct {
		Message      message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Client talks to a single OpenAI-compatible provider endpoint.
type Client struct {
	baseURL     string
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
	verbose     bool
	httpClient  *http.Client
}

// New builds a Client from the resolved config. When verbose is true,
// Complete logs the outgoing request and raw response to stderr.
func New(cfg *config.Config, verbose bool) (*Client, error) {
	p, err := cfg.ActiveProvider()
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:     strings.TrimRight(p.BaseURL, "/"),
		apiKey:      p.ResolveAPIKey(),
		model:       cfg.Default.Model,
		temperature: cfg.Default.Temperature,
		maxTokens:   cfg.Default.MaxTokens,
		verbose:     verbose,
		httpClient:  &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// Complete sends a system+user prompt pair and returns the model's raw text reply.
func (c *Client) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := request{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
		Think:       false,
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("could not build request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	if c.verbose {
		fmt.Fprintf(os.Stderr, "nox: --> POST %s\n%s\n", url, buf)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("could not connect to %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	respBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "nox: <-- HTTP %d\n%s\n", resp.StatusCode, respBuf)
	}

	var parsed response
	if err := json.Unmarshal(respBuf, &parsed); err != nil {
		return "", fmt.Errorf("could not parse response (HTTP %d): %s", resp.StatusCode, string(respBuf))
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("provider error (HTTP %d): %s", resp.StatusCode, parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("provider returned HTTP %d: %s", resp.StatusCode, string(respBuf))
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("provider returned no choices (HTTP %d)", resp.StatusCode)
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		if parsed.Choices[0].FinishReason == "length" {
			return "", fmt.Errorf("response got cut off before any content was written (hit max_tokens=%d) — this model may spend tokens on internal reasoning first; try raising default.max_tokens in your config", c.maxTokens)
		}
		return "", fmt.Errorf("provider returned an empty message (finish_reason=%q); rerun with --verbose to see the raw response", parsed.Choices[0].FinishReason)
	}
	return content, nil
}
