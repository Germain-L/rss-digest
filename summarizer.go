package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Summarizer struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewSummarizer(apiKey, model string) *Summarizer {
	if model == "" {
		model = "glm-4-flash"
	}
	return &Summarizer{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (s *Summarizer) Summarize(ctx context.Context, items []FeedItem) (string, error) {
	// Build content from items (limit to last 80 items for better context)
	maxItems := 80
	if len(items) > maxItems {
		items = items[:maxItems]
	}

	var content strings.Builder
	for _, item := range items {
		content.WriteString(fmt.Sprintf("## %s\n", item.Title))
		content.WriteString(fmt.Sprintf("Source: %s\n", item.Source))
		if item.Description != "" {
			desc := item.Description
			if len(desc) > 400 {
				desc = desc[:400] + "..."
			}
			content.WriteString(fmt.Sprintf("%s\n", desc))
		}
		content.WriteString("\n---\n\n")
	}

	prompt := fmt.Sprintf(`You are a news curator. Summarize the following RSS feed items into a concise daily digest.

Rules:
- Group related stories together by topic
- Highlight the most important news
- Be concise but informative
- Skip duplicates or very similar stories
- Format with clear sections using markdown headers (##) and bullet points
- Include source names in parentheses after each item
- Create sections like: Conflict & Geopolitics, Politics & Elections, Business & Tech, Science & Environment, Other News

Here are today's news items:

%s`, content.String())

	reqBody := ChatRequest{
		Model: s.model,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Use GLM API endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", "https://open.bigmodel.cn/api/paas/v4/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parse error: %v, body: %s", err, string(respBody))
	}

	if chatResp.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API: %s", string(respBody))
	}

	return chatResp.Choices[0].Message.Content, nil
}

// Backward compatibility
type GroqSummarizer = Summarizer

func NewGroqSummarizer(apiKey string) *Summarizer {
	return NewSummarizer(apiKey, "glm-4-flash")
}
