package pkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"
)

type ModelType string

const (
	ModelHighCost ModelType = "hc"
	ModelLowCost  ModelType = "lc"
)

func (m ModelType) IsValid() bool {
	return m == ModelHighCost || m == ModelLowCost
}

type GenType string

const (
	GenTypeDetailed GenType = "d"
	GenTypeRough    GenType = "r"
)

func (g GenType) IsValid() bool {
	return g == GenTypeDetailed || g == GenTypeRough
}

type EngineConfig struct {
	Model     ModelType
	GenType   GenType
	NumScript int
	Pattern   string
	FilePath  string
	SaveDir   string
}

type Engine struct {
	config     *EngineConfig
	llmClient  *LLMClient
	taskStore  *TaskStore
	maxWorkers int
}

func NewEngineConfig(model, gentype, pattern, filePath, saveDir string, numScript int) (*EngineConfig, error) {
	m := ModelType(model)
	if !m.IsValid() {
		return nil, errors.New("invalid model type: " + model)
	}
	g := GenType(gentype)
	if !g.IsValid() {
		return nil, errors.New("invalid generation type: " + gentype)
	}
	// Validate file path is within allowed directories
	cleaned := filepath.Clean(filePath)
	if strings.Contains(cleaned, "..") {
		return nil, errors.New("invalid file path")
	}
	return &EngineConfig{
		Model:     m,
		GenType:   g,
		NumScript: numScript,
		Pattern:   pattern,
		FilePath:  filePath,
		SaveDir:   saveDir,
	}, nil
}

func NewEngine(config *EngineConfig, llmClient *LLMClient, taskStore *TaskStore, maxWorkers int) (*Engine, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if config.FilePath == "" {
		return nil, errors.New("file path cannot be empty")
	}
	if llmClient == nil {
		return nil, errors.New("LLM client cannot be nil")
	}
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", config.FilePath)
	}
	return &Engine{
		config:     config,
		llmClient:  llmClient,
		taskStore:  taskStore,
		maxWorkers: maxWorkers,
	}, nil
}

func genSingle(ctx context.Context, client *LLMClient, episodeNum int, chapterContent string, gentype GenType) (string, error) {
	var maxTokens int
	if gentype == GenTypeDetailed {
		maxTokens = 1000
	} else {
		maxTokens = 500
	}
	prompt := fmt.Sprintf(`请将以下小说片段转换为%d字以内的电视剧剧本，需要详细还原原小说中的细节。要求：

1.格式为%d--1、大场景/小场景 时间 场景描述\n人物表。当前为第%d集。
2.用△表示一个镜头，后面紧接镜头拍摄的内容。
3.人物表仅指示当前幕的人物，本幕未出现的人物不必列出。
4.突出小说中的精彩情节，使得整体剧本紧凑且富有戏剧性。
5.对话的内容需要有之前内容的铺垫，使得情节清晰严谨。
6.V.O.表示人物说话，O.S.表示人物内心独白。
7.无需添加markdown格式，直接输出纯文本。
以下是一个简单的格式示例：

1--1、某酒店/宴会厅 晚上 宾客众多，觥筹交错
人物：
AA
BB

△AA和BB在宴会上相遇，互相交换名片。
AA(V.O./微笑)：你好，我是AA。
BB(V.O./热情)：你好，AA，我是BB。
...
1--2、监狱/囚室中 深夜 寂静无声
人物：
CC

△CC(19岁，囚服脏旧)猛地抽搐，从水泥地上惊醒，剧痛袭遍全身。
CC(V.O./痛哼)：呃…痛！好痛…这是…哪里？
CC(O.S.)：我这是怎么了？
...

原文
%s

/no_think`, maxTokens, episodeNum, episodeNum, chapterContent)

	return client.ChatCompletion(ctx, prompt)
}

func saveScript(savePath, content string, index int) error {
	thinkRegex := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleanedContent := thinkRegex.ReplaceAllString(content, "")
	cleanedContent = strings.ReplaceAll(cleanedContent, "*", "")
	cleanedContent = strings.ReplaceAll(cleanedContent, "#", "")
	cleanedContent = fmt.Sprintf("第%d集\n\n%s", index, cleanedContent)
	digitDashRegex := regexp.MustCompile(`\d+--\d+`)
	counter := 1
	result := digitDashRegex.ReplaceAllStringFunc(cleanedContent, func(match string) string {
		replacement := fmt.Sprintf("%d--%d", index, counter)
		counter++
		return replacement
	})
	if err := os.WriteFile(savePath, []byte(result), 0644); err != nil {
		return fmt.Errorf("failed to save script: %w", err)
	}
	return nil
}

// GenerateScriptAsync starts script generation in the background and returns a task ID.
func (e *Engine) GenerateScriptAsync(ctx context.Context, taskID string) {
	total := e.config.NumScript
	e.taskStore.UpdateStatus(taskID, TaskStatusRunning, 0, "")

	err := e.generateScriptInternal(ctx, taskID, total)
	if err != nil {
		e.taskStore.UpdateStatus(taskID, TaskStatusFailed, 0, err.Error())
		return
	}
	e.taskStore.UpdateStatus(taskID, TaskStatusCompleted, total, "")
	e.taskStore.SetSaveDir(taskID, e.config.SaveDir)
}

func (e *Engine) generateScriptInternal(ctx context.Context, taskID string, total int) error {
	readerEngine := NewDocumentReaderEngine()
	content, err := readerEngine.ReadDocument(e.config.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read document: %w", err)
	}

	chapters, err := SplitText(content, e.config.Pattern)
	if err != nil {
		return fmt.Errorf("failed to split text: %w", err)
	}

	bins, err := CalBins(len(chapters), e.config.NumScript)
	if err != nil {
		return fmt.Errorf("failed to calculate bins: %w", err)
	}

	mergedChapters := MergeChapter(bins, chapters)
	if len(mergedChapters) != e.config.NumScript {
		return fmt.Errorf("merged chapters count mismatch: %d != %d", len(mergedChapters), e.config.NumScript)
	}

	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, e.maxWorkers)
	results := make([]string, len(mergedChapters))

	for i, merged := range mergedChapters {
		i, merged := i, merged
		g.Go(func() error {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
			defer func() { <-sem }()

			result, err := genSingle(ctx, e.llmClient, i+1, merged, e.config.GenType)
			if err != nil {
				return fmt.Errorf("episode %d: %w", i+1, err)
			}
			results[i] = result
			e.taskStore.UpdateStatus(taskID, TaskStatusRunning, i+1, "")
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if err := os.MkdirAll(e.config.SaveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	for i, result := range results {
		fullPath := filepath.Join(e.config.SaveDir, fmt.Sprintf("script_%02d.txt", i+1))
		if err := saveScript(fullPath, result, i+1); err != nil {
			return fmt.Errorf("failed to save episode %d: %w", i+1, err)
		}
	}

	return nil
}
