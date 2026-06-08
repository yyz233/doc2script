package pkg

import (
	"os"
	"strconv"
)

type Config struct {
	Port             string
	TempDir          string
	SaveDir          string
	MaxUploadSize    int64
	LLMBaseURL       string
	LLMModelNameHC   string
	LLMModelNameLC   string
	LLMAPIKey        string
	MaxConcurrency   int
	RequestTimeoutMs int
}

func LoadConfig() *Config {
	return &Config{
		Port:             envOrDefault("PORT", "8081"),
		TempDir:          envOrDefault("TEMP_DIR", "./temp"),
		SaveDir:          envOrDefault("SAVE_DIR", "./scripts"),
		MaxUploadSize:    envInt64OrDefault("MAX_UPLOAD_SIZE", 100<<20),
		LLMBaseURL:       envOrDefault("LLM_BASE_URL", ""),
		LLMModelNameHC:   envOrDefault("LLM_MODEL_NAME_HC", ""),
		LLMModelNameLC:   envOrDefault("LLM_MODEL_NAME_LC", ""),
		LLMAPIKey:        envOrDefault("LLM_API_KEY", ""),
		MaxConcurrency:   envIntOrDefault("MAX_CONCURRENCY", 5),
		RequestTimeoutMs: envIntOrDefault("REQUEST_TIMEOUT_MS", 300000),
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envIntOrDefault(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func envInt64OrDefault(key string, defaultVal int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return defaultVal
}
