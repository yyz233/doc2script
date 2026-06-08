package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"doc2script/pkg"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GenerateRequest struct {
	Model     string `json:"model" binding:"required"`
	GenType   string `json:"gentype" binding:"required"`
	Pattern   string `json:"pattern" binding:"required"`
	FilePath  string `json:"file_path" binding:"required"`
	NumScript int    `json:"num_script" binding:"required"`
}

type TaskSubmitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TaskID  string `json:"task_id,omitempty"`
}

type TaskStatusResponse struct {
	Success   bool            `json:"success"`
	Task      *pkg.Task       `json:"task,omitempty"`
	Message   string          `json:"message,omitempty"`
}

type Service struct {
	uploader  *pkg.Uploader
	cfg       *pkg.Config
	llmClient *pkg.LLMClient
	taskStore *pkg.TaskStore
}

func NewService(cfg *pkg.Config) *Service {
	os.MkdirAll(cfg.TempDir, 0755)
	os.MkdirAll(cfg.SaveDir, 0755)

	var llmClient *pkg.LLMClient
	if cfg.LLMBaseURL != "" {
		llmClient = pkg.NewLLMClient(cfg.LLMBaseURL, cfg.LLMModelNameLC, cfg.LLMAPIKey)
	}

	return &Service{
		uploader:  pkg.NewUploader(cfg.TempDir),
		cfg:       cfg,
		llmClient: llmClient,
		taskStore: pkg.NewTaskStore(),
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func (s *Service) submitGenerate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TaskSubmitResponse{
			Success: false,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if s.llmClient == nil {
		c.JSON(http.StatusInternalServerError, TaskSubmitResponse{
			Success: false,
			Message: "LLM service not configured",
		})
		return
	}

	config, err := pkg.NewEngineConfig(
		req.Model,
		req.GenType,
		req.Pattern,
		req.FilePath,
		s.cfg.SaveDir,
		req.NumScript,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, TaskSubmitResponse{
			Success: false,
			Message: "config error: " + err.Error(),
		})
		return
	}

	engine, err := pkg.NewEngine(config, s.llmClient, s.taskStore, s.cfg.MaxConcurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, TaskSubmitResponse{
			Success: false,
			Message: "failed to create engine: " + err.Error(),
		})
		return
	}

	taskID := uuid.New().String()
	s.taskStore.Create(taskID, req.NumScript)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.RequestTimeoutMs)*time.Millisecond)
	// Run generation in background goroutine; the context with timeout ensures it won't run forever.
	go func() {
		defer cancel()
		engine.GenerateScriptAsync(ctx, taskID)
	}()

	c.JSON(http.StatusAccepted, TaskSubmitResponse{
		Success: true,
		Message: "task submitted",
		TaskID:  taskID,
	})
}

func (s *Service) getTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	task := s.taskStore.Get(taskID)
	if task == nil {
		c.JSON(http.StatusNotFound, TaskStatusResponse{
			Success: false,
			Message: "task not found",
		})
		return
	}

	c.JSON(http.StatusOK, TaskStatusResponse{
		Success: true,
		Task:    task,
	})
}

func (s *Service) setupRoutes() *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = s.cfg.MaxUploadSize
	r.Use(requestIDMiddleware())

	api := r.Group("/api/v1")
	{
		api.POST("/upload", s.uploader.UploadSingleFile)
		api.POST("/generate", s.submitGenerate)
		api.GET("/task/:id", s.getTaskStatus)
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})

	return r
}

func main() {
	cfg := pkg.LoadConfig()

	srv := NewService(cfg)
	r := srv.setupRoutes()

	// Periodic temp file cleanup
	cleanupDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleanTempFiles(cfg.TempDir, 24*time.Hour)
			case <-cleanupDone:
				return
			}
		}
	}()

	s := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("server starting on port %s", cfg.Port)
		log.Printf("upload dir: %s, save dir: %s", cfg.TempDir, cfg.SaveDir)
		if cfg.LLMBaseURL != "" {
			log.Printf("LLM backend: %s", cfg.LLMBaseURL)
		} else {
			log.Printf("WARNING: LLM_BASE_URL not set, generation endpoint will not work")
		}
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server startup failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully...")
	close(cleanupDone)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func cleanTempFiles(tempDir string, maxAge time.Duration) {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(tempDir, entry.Name()))
		}
	}
}
