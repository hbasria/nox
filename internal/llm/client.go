// Package llm implements a minimal OpenAI-compatible chat completions client.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

type response struct {
	Choices []struct {
		Message message `json:"message"`
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
	httpClient  *http.Client
}

// New builds a Client from the resolved config.
func New(cfg *config.Config) (*Client, error) {
	p, err := cfg.ActiveProvider()
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:     strings.TrimRight(p.BaseURL, "/"),
		apiKey:      p.APIKey(),
		model:       cfg.Default.Model,
		temperature: cfg.Default.Temperature,
		maxTokens:   cfg.Default.MaxTokens,
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
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("istek hazırlanamadı: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("istek oluşturulamadı: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("%s adresine bağlanılamadı: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	respBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("yanıt okunamadı: %w", err)
	}

	var parsed response
	if err := json.Unmarshal(respBuf, &parsed); err != nil {
		return "", fmt.Errorf("yanıt parse edilemedi (HTTP %d): %s", resp.StatusCode, string(respBuf))
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("provider hatası (HTTP %d): %s", resp.StatusCode, parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("provider HTTP %d döndürdü: %s", resp.StatusCode, string(respBuf))
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("provider boş yanıt döndürdü")
	}

	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}
