package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/nguyenthenguyen/docx"
)

type DocumentReader interface {
	ReadDocument(filePath string) (string, error)
	SupportedExtensions() []string
}

type DocxReader struct{}

// DOCX Reader
func (d *DocxReader) ReadDocument(filePath string) (string, error) {
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开DOCX文件: %v", err)
	}
	defer r.Close()
	docx := r.Editable()
	content := docx.GetContent()
	
	return content, nil
}

func (d *DocxReader) SupportedExtensions() []string {
	return []string{".docx"}
}

// DocumentReaderEngine 
type DocumentReaderEngine struct {
	readers map[string]DocumentReader
}

func (de *DocumentReaderEngine) RegisterReader(reader DocumentReader) {
	for _, ext := range reader.SupportedExtensions() {
		de.readers[strings.ToLower(ext)] = reader
	}
}

func (de *DocumentReaderEngine) GetSupportedExtensions() []string {
	var extensions []string
	for ext := range de.readers {
		extensions = append(extensions, ext)
	}
	return extensions
}

func NewDocumentReaderEngine() *DocumentReaderEngine {
	de := &DocumentReaderEngine{
		readers: make(map[string]DocumentReader),
	}
	de.RegisterReader(&DocxReader{})
	return de
}

func (de *DocumentReaderEngine) ReadDocument(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if reader, exists := de.readers[ext]; exists {
		return reader.ReadDocument(filePath)
	}
	return "", fmt.Errorf("不支持的文件类型: %s", ext)
}