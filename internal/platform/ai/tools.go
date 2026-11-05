package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}
type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type toolChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Tools       []Tool        `json:"tools"`
	ToolChoice  string        `json:"tool_choice"`
	Temperature float64       `json:"temperature"`
}

type toolChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func (c *Client) ChatWithTools(ctx context.Context, messages []ChatMessage, tools []Tool) (ChatMessage, error) {
	reqBody := toolChatRequest{
		Model:       c.model,
		Messages:    messages,
		Tools:       tools,
		ToolChoice:  "auto",
		Temperature: 0.3,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return ChatMessage{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, groqURL, bytes.NewReader(body))
	if err != nil {
		return ChatMessage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return ChatMessage{}, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return ChatMessage{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return ChatMessage{}, fmt.Errorf("groq returned status %d: %s", resp.StatusCode, string(data))
	}

	var result toolChatResponse

	err = json.Unmarshal(data, &result)
	if err != nil {
		return ChatMessage{}, err
	}

	if len(result.Choices) == 0 {
		return ChatMessage{}, fmt.Errorf("groq returned no choices")
	}
	return result.Choices[0].Message, nil
}
