package pkg

import (
	"context"
	"testing"
	"time"
)

func TestChatCompletion(t *testing.T) {
	client := NewLLMClient("http://192.168.11.218:8192/v1", "novel", "None")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := client.ChatCompletion(ctx, "Hello")
	if err != nil {
		t.Logf("LLM call failed (expected if LLM is unreachable): %v", err)
		return
	}
	if result == "" {
		t.Error("expected non-empty response")
	}
	trunc := len(result)
	if trunc > 100 {
		trunc = 100
	}
	t.Logf("response: %s", result[:trunc])
}
