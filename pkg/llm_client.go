package pkg

import (
	"context"
	"errors"
	"github.com/sashabaranov/go-openai"
)

type CompletionConfig struct {
	BaseURL string 
	ModelName string 
	APIKey string
	Prompt string
}

func ChatCompletion(cc CompletionConfig) (string, error) {
	config := openai.DefaultConfig(cc.APIKey)
	config.BaseURL = cc.BaseURL
	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: cc.ModelName,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: cc.Prompt,
				},
			},
			MaxTokens:   2048,
			Temperature: 0.7,
		},
	)
	if err != nil {
		return "", errors.New("LLM completion error :" + err.Error())
	}
	return resp.Choices[0].Message.Content, nil
}