package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/fatih/color"
)

var (
	reDeletion        = regexp2.MustCompile("\\[\\-(((?!\\-\\]).)+)\\-\\]", 0)
	reAddition        = regexp2.MustCompile("\\{\\+(((?!\\+\\}).)+)\\+\\}", 0)
	reLineDiffDiscard = regexp2.MustCompile(",[0-9]+", 0)
)

func removeAndColorize(remove, keep *regexp2.Regexp, c *color.Color, s string) string {
	// TODO these panics are dangerous, they should be fixed
	stripped, err := remove.Replace(s, "", 0, -1)
	if err != nil {
		panic(err)
	}
	colored, err := keep.ReplaceFunc(stripped, func(m regexp2.Match) string {
		match, err := keep.FindStringMatch(m.String())
		if err != nil {
			panic(err)
		}
		return c.Sprintf(match.Groups()[1].Captures[0].String())
	}, 0, -1)
	if err != nil {
		panic(err)
	}
	return colored
}

var (
	textGreen = color.New(color.BgGreen, color.FgBlack)
	textRed   = color.New(color.BgRed, color.FgBlack)
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

const (
	leftLineHL  = "\x1b[0m"
	rightLineHL = "\x1b[0m"
)

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
			var err error
			changed[i], err = reLineDiffDiscard.Replace(changed[i], "", 0, -1)
			if err != nil {
				return nil, err
			}
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
		//prefix := " "
		//if s, ok := changesMap[i+1]; ok {
		//if s < 0 {
		//prefix = "-"
		//} else {
		//prefix = "+"
		//}
		//}
		line = strings.Replace(line, "%", "%%", -1)
		//line = fmt.Sprintf("%s %s", prefix, line)
		wordsDiffByLine[i] = fmt.Sprintf(format, i+1, line)
	}

	joined := strings.Join(wordsDiffByLine, "\n")

	left := strings.Split(RemoveAdditions(joined), "\n")
	right := strings.Split(RemoveDeletions(joined), "\n")

	for i := range wordsDiffByLine {
		if _, ok := changesMap[i+1]; !ok {
			continue
		}
		left[i] = leftLineHL + strings.Replace(left[i], "\x1b[0m", leftLineHL, -1) + "\x1b[0m"
		right[i] = rightLineHL + strings.Replace(right[i], "\x1b[0m", rightLineHL, -1) + "\x1b[0m"
	}

	return &Diff{
		Left:  strings.Join(left, "\n"),
		Right: strings.Join(right, "\n"),
	}, nil
}
