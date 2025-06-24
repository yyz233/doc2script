package pkg
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ModelType string  

const (
	ModelHighCost ModelType = "hc"
	ModelLowCost ModelType = "lc"
)

func (m ModelType) IsValid() bool {
	return m == ModelHighCost || m == ModelLowCost
}

type GenType string

const (
	GenTypeDetailed GenType = "d"
	GenTypeRough GenType = "r"
)

func (g GenType) IsValid() bool {
	return g == GenTypeDetailed || g == GenTypeRough
}

type EngineConfig struct {
	model ModelType //选择何种模型生成方案(High Cost/Low cost)
	gentype GenType //是否生成详细剧本
	num int //生成的剧本数量
	pattern string //正则表达式，用于标记章节位置，例如 `^第[一二三四五六七八九十百千]+章`
	path string //小说内容存储的位置
	savePath string //生成剧本的存储位置
}

type Engine struct {
	config *EngineConfig
}

func NewEngineConfig(model, gentype, pattern, path, savePath string, numScript int) (*EngineConfig, error) {
	m := ModelType(model)
	if !m.IsValid() {
		return nil, errors.New("Invalid model type: " + model)
	}
	g := GenType(gentype)
	if !g.IsValid() {
		return nil, errors.New("Invalid generation type: " + gentype)
	}
	ec := &EngineConfig{
		model: m,
		pattern: pattern,
		gentype: g,
		num: numScript,
		path: path,
		savePath: savePath,
	}
	return ec, nil
} 

func NewEngine(config *EngineConfig) (*Engine, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if config.path == "" {
		return nil, errors.New("path cannot be empty")
	}
	engine := &Engine{
		config: config,
	}
	return engine, nil
}

func GenSingle(chapterNum int, chapterContent string, gentype GenType, baseUrl, modelName, apiKey string) (string, error) {
	var s int
	if gentype == GenTypeDetailed {
		s = 1000
	} else {
		s = 500
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

/no_think`, s, chapterNum, chapterNum, chapterContent)
	config := CompletionConfig{
		BaseURL: baseUrl,
		ModelName: modelName,
		APIKey: apiKey,
		Prompt: prompt,
	}
	result, err := ChatCompletion(config)
	if err != nil {
		return "", err
	}
	return result, nil
}

func SaveScript(savePath, content string, index int) error {
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
		return fmt.Errorf("保存文件失败: %w", err)
	}
	return nil
}

func (e *Engine) GenerateScript() error {
	readerEngine := NewDocumentReaderEngine()
	content, err := readerEngine.ReadDocument(e.config.path)
	if err != nil {
		return fmt.Errorf("failed to read document: %v", err)
	}
	chapters, err := SplitText(content, e.config.pattern)
	if err != nil {
		return fmt.Errorf("failed to split text: %v", err)
	}
	bins, err := CalBins(len(chapters), e.config.num)
	if err != nil {
		return fmt.Errorf("failed to calculate bins: %v", err)
	}
	var baseUrl, modelName, apiKey string
	if e.config.model == ModelLowCost {
		baseUrl = "http://192.168.11.218:8192/v1"
		modelName = "novel"
		apiKey = "None"
		// only for test
	} else {
		//to be implemented when high cost model is available
	}
	now := 0
	for i, bin := range bins {
		var result string
		result = ""
		for j := 0; j < bin; j++ {
			response, err := GenSingle(i+1, chapters[now], e.config.gentype, baseUrl, modelName, apiKey)
			if err != nil {
				return fmt.Errorf("failed to generate script for chapter %d: %w", now+1, err)
			}
			result += response + "\n"
			now += 1
		}
		fullPath := filepath.Join(e.config.savePath, fmt.Sprintf("script_%02d.txt", i+1))
		err := SaveScript(fullPath, result, i+1)
		if err != nil {
			return fmt.Errorf("failed to save script %d: %w", i+1, err)
		}
	}
	return nil
}