package parser

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var (
	reDeletion        = regexp.MustCompile("\\[\\-([^(?\\-\\])]*)\\-\\]")
	reAddition        = regexp.MustCompile("\\{\\+([^(?\\+\\})]*)\\+\\}")
	reLineDiffDiscard = regexp.MustCompile(",[0-9]+")
)

func removeAndColorize(remove, keep *regexp.Regexp, c *color.Color, s string) string {
	stripped := remove.ReplaceAllString(s, "")
	colored := keep.ReplaceAllStringFunc(stripped, func(m string) string {
		match := keep.FindAllStringSubmatch(m, -1)[0][1]
		return c.Sprintf(match)
	})
	return colored
}

var (
	textGreen = color.New(color.BgGreen, color.FgBlack)
	textRed   = color.New(color.BgRed, color.FgBlack)
	textLine  = color.New(color.BgMagenta, color.FgWhite)
)

func RemoveDeletions(s string) string {
	return removeAndColorize(reDeletion, reAddition, textGreen, s)
}

func RemoveAdditions(s string) string {
	return removeAndColorize(reAddition, reDeletion, textRed, s)
}

type Diff struct {
	Left  string
	Right string
}

func Parse(diffLines, diffWords string) (*Diff, error) {
	// remove header
	wordsDiffByLine := strings.Split(diffWords, "\n")[5:]

	changesMap := make(map[int]int)
	linesDiffByLine := strings.Split(diffLines, "\n")
	for _, line := range linesDiffByLine {
		if !strings.HasPrefix(line, "@@") {
			continue
		}
		changed := strings.Fields(strings.Split(line, "@@")[1])
		for i := range changed {
			changed[i] = reLineDiffDiscard.ReplaceAllString(changed[i], "")
			n, err := strconv.Atoi(changed[i])
			if err != nil {
				return nil, err
			}
			abs := int(math.Abs(float64(n)))
			changesMap[abs] = n
		}
	}

	lineIndent := int(math.Ceil(math.Log10(float64(len(wordsDiffByLine) + 1))))
	format := fmt.Sprintf("%%%dd %%s", lineIndent)
	for i, line := range wordsDiffByLine {
		prefix := " "
		if s, ok := changesMap[i+1]; ok {
			if s < 0 {
				prefix = "-"
			} else {
				prefix = "+"
			}
		}
		line = fmt.Sprintf("%s %s", prefix, line)
		wordsDiffByLine[i] = fmt.Sprintf(format, i+1, line)
	}

	joined := strings.Join(wordsDiffByLine, "\n")

	left := strings.Split(RemoveAdditions(joined), "\n")
	right := strings.Split(RemoveDeletions(joined), "\n")

	magenta := "\x1b[47;30m"
	cyan := "\x1b[47;30m"
	for i := range wordsDiffByLine {
		if _, ok := changesMap[i+1]; !ok {
			continue
		}
		left[i] = magenta + strings.Replace(left[i], "\x1b[0m", magenta, -1) + "\x1b[0m"
		right[i] = cyan + strings.Replace(right[i], "\x1b[0m", cyan, -1) + "\x1b[0m"
	}

	return &Diff{
		Left:  strings.Join(left, "\n"),
		Right: strings.Join(right, "\n"),
	}, nil
}
