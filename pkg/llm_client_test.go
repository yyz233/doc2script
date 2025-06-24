package pkg
import (
	"testing"
)

func TestCompletion(t *testing.T) {
	config := CompletionConfig{
		BaseURL: "http://192.168.11.218:8192/v1",
		ModelName: "novel",
		APIKey: "None",
		Prompt: "Hello",
	}
	result, err := ChatCompletion(config)
	t.Log(result)
}

