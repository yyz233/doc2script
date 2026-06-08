package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentReaderEngine(t *testing.T) {
	engine := NewDocumentReaderEngine()
	exts := engine.GetSupportedExtensions()
	if len(exts) == 0 {
		t.Fatal("expected at least one supported extension")
	}
	t.Logf("supported extensions: %v", exts)

	// Test unsupported extension
	_, err := engine.ReadDocument("test.xyz")
	if err == nil {
		t.Fatal("expected error for unsupported file type")
	}

	// Test non-existent file
	_, err = engine.ReadDocument("nonexistent.docx")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestDocxReaderIntegration(t *testing.T) {
	// Skip if no test file available
	testFile := filepath.Join("..", "test.docx")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test.docx not found, skipping integration test")
	}
	engine := NewDocumentReaderEngine()
	content, err := engine.ReadDocument(testFile)
	if err != nil {
		t.Fatalf("failed to read document: %v", err)
	}
	if content == "" {
		t.Fatal("expected non-empty content")
	}
	t.Logf("read %d characters", len(content))
}
