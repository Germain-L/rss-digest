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

type GroqSummarizer struct {
	apiKey     string
	httpClient *http.Client
}

type GroqRequest struct {
	Model    string          `json:"model"`
	Messages []GroqMessage   `json:"messages"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewGroqSummarizer(apiKey string) *GroqSummarizer {
	return &GroqSummarizer{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (g *GroqSummarizer) Summarize(ctx context.Context, items []FeedItem) (string, error) {
	// Build content from items (limit to last 50 items, most recent first)
	maxItems := 50
	if len(items) > maxItems {
		items = items[:maxItems]
	}

	var content strings.Builder
	for _, item := range items {
		content.WriteString(fmt.Sprintf("## %s\n", item.Title))
		content.WriteString(fmt.Sprintf("Source: %s\n", item.Source))
		if item.Description != "" {
			desc := item.Description
			if len(desc) > 300 {
				desc = desc[:300] + "..."
			}
			content.WriteString(fmt.Sprintf("%s\n", desc))
		}
		content.WriteString("\n---\n\n")
	}

	prompt := fmt.Sprintf(`You are a tech news curator. Summarize the following RSS feed items into a concise daily digest.

Rules:
- Group related stories together
- Highlight the most important news
- Be concise but informative
- Skip duplicates or very similar stories
- Format with clear sections and bullet points
- Include source names in parentheses

Here are today's news items:

%s`, content.String())

	reqBody := GroqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []GroqMessage{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var groqResp GroqResponse
	if err := json.Unmarshal(respBody, &groqResp); err != nil {
		return "", err
	}

	if groqResp.Error.Message != "" {
		return "", fmt.Errorf("Groq API error: %s", groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return groqResp.Choices[0].Message.Content, nil
}
