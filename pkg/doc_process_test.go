package pkg

import (
	"testing"
)

func TestSplitText(t *testing.T) {
	text := `
这是前言内容，不会被包含在拆分结果中。

第一章 故事的开始
这是第一章的内容...
很多很多文字内容...

第二章 情节的发展
这是第二章的内容...
故事继续展开...

第三章 精彩的结局
这是第三章的内容...
故事达到高潮并结束...`
	chapters, err := SplitText(text, `(?m)^第[一二三四五六七八九十\d]+章`)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(chapters) != 3 {
		t.Fatalf("expected 3 chapters, got %d", len(chapters))
	}
	t.Logf("found %d chapters", len(chapters))
}

func TestCalBins(t *testing.T) {
	result, err := CalBins(10, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum := 0
	for _, v := range result {
		sum += v
	}
	if sum != 10 {
		t.Fatalf("expected sum 10, got %d", sum)
	}
	t.Logf("bins: %v", result)
}

func TestMergeChapter(t *testing.T) {
	chapters := []string{"aaa", "bbb", "ccc", "ddd", "eee"}
	bins := []int{2, 3}
	merged := MergeChapter(bins, chapters)
	if len(merged) != 2 {
		t.Fatalf("expected 2 merged groups, got %d", len(merged))
	}
	if merged[0] != "aaabbb" {
		t.Fatalf("unexpected merged[0]: %s", merged[0])
	}
	if merged[1] != "cccdddeee" {
		t.Fatalf("unexpected merged[1]: %s", merged[1])
	}
}

func TestCalBinsError(t *testing.T) {
	_, err := CalBins(2, 5)
	if err == nil {
		t.Fatal("expected error when total < bins")
	}
}
