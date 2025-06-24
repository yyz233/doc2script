package pkg

import (
	"testing"
)

func TestDocumentReaderEngine (t *testing.T) {
	engine := NewDocumentReaderEngine()
	_, err := engine.ReadDocument("D:\\work\\project\\doc2script\\test.docx")
	if err != nil {
		t.Fatalf("读取失败 %v", err)
	}
}