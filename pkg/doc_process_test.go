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
	chapters, err := SplitText(text, `第[一二三四五六七八九十\d]+章`)
	if err != nil {
		t.Errorf("expected nil error, got:%v", err)
	}
	t.Log(chapters)
}