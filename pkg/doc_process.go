package pkg

import (
	"errors"
	"regexp"
	"log"
	"strings"
)

var ErrNoMatchFound = errors.New("未找到匹配的章节分割标记")
func SplitText(text string, pattern string) ([]string, error){
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	log.Printf("开始拆分文本，总长度： %d 字符",len(text))

	matches := regex.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, ErrNoMatchFound
	}
	chapters := make([]string, len(matches))
	for i, match := range matches {
		start := match[0]
		var end int 
		if i + 1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(text)
		}
		chapters[i] = text[start:end]
	}
	return chapters, nil
}

func CalBins(total, bins int) ([]int, error) {
	if total < bins {
		return nil, errors.New("总数不能小于分组数")
	}
	result := make([]int, bins)
	baseAmount := total / bins
	for i := 0; i < bins; i++ {
		result[i] = baseAmount
	}
	
	remainder := total % bins
	for i := 0; i < bins; i++ {
		result[i] = baseAmount
	}
	for i := 0; i < remainder; i++ {
		result[i]++
	}
	
	return result, nil
}

func MergeChapter(binNum []int, chapters []string) ([]string){
	var result []string
	chapterIndex := 0
	for i := 0; i < len(binNum); i++ {
		var mergedChapter strings.Builder
		for j := 0; j < binNum[i] && chapterIndex < len(chapters); j++ {
			mergedChapter.WriteString(chapters[chapterIndex])
			chapterIndex++
		}
		
		result = append(result, mergedChapter.String())
	}
	return result
}