package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

type LLMClient struct {
	client  *openai.Client
	model   string
	baseURL string
	apiKey  string
	mu      sync.Mutex
	once    sync.Once
}

func NewLLMClient(baseURL, modelName, apiKey string) *LLMClient {
	return &LLMClient{
		baseURL: baseURL,
		model:   modelName,
		apiKey:  apiKey,
	}
}

func (c *LLMClient) getClient() *openai.Client {
	c.once.Do(func() {
		cfg := openai.DefaultConfig(c.apiKey)
		cfg.BaseURL = c.baseURL
		c.client = openai.NewClientWithConfig(cfg)
	})
	return c.client
}

func (c *LLMClient) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	return c.chatCompletionWithRetry(ctx, prompt, 2)
}

func (c *LLMClient) chatCompletionWithRetry(ctx context.Context, prompt string, retries int) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}

		resp, err := c.getClient().CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			MaxTokens:   2048,
			Temperature: 0.7,
		})
		if err != nil {
			lastErr = err
			continue
		}
		if len(resp.Choices) == 0 {
			lastErr = fmt.Errorf("empty response from LLM")
			continue
		}
		return resp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("LLM completion failed after %d retries: %w", retries, lastErr)
}
