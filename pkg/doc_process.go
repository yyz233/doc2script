package pkg

import (
	"errors"
	"regexp"
	"strings"
)

var ErrNoMatchFound = errors.New("no matching chapter markers found")
var ErrTotalLessThanBins = errors.New("total chapters cannot be less than bins")

func SplitText(text string, pattern string) ([]string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	matches := regex.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, ErrNoMatchFound
	}
	chapters := make([]string, len(matches))
	for i, match := range matches {
		start := match[0]
		var end int
		if i+1 < len(matches) {
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
		return nil, ErrTotalLessThanBins
	}
	result := make([]int, bins)
	baseAmount := total / bins
	remainder := total % bins
	for i := 0; i < bins; i++ {
		result[i] = baseAmount
		if i < remainder {
			result[i]++
		}
	}
	return result, nil
}

func MergeChapter(binNum []int, chapters []string) []string {
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
