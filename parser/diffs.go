package parser

import (
	"fmt"
	"strings"

	"github.com/ruggi/kit/runner"
)

const (
	gitFolder = "." // TODO make this selectable
)

func LoadDiffs() ([]*Diff, error) {
	b := runner.NewBuffered()
	err := b.Git(gitFolder, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	status := strings.Split(strings.TrimSpace(b.Buffer.String()), "\n")

	diffs := make([]*Diff, 0, len(status))
	for i := range status {
		fields := strings.Fields(status[i])
		if len(fields) != 2 {
			return nil, fmt.Errorf("diff is empty") // TODO better error management here
		}
		mode, file := fields[0], fields[1]
		if mode != "M" {
			continue
		}

		b.Buffer.Reset()
		err = b.Git(gitFolder, "diff", "-U0")
		if err != nil {
			return nil, err
		}
		diffLines := strings.TrimSpace(b.Buffer.String())

		b.Buffer.Reset()
		err = b.Git(gitFolder, "diff", "--word-diff=plain", "-U1000", file) // TODO this should account for the actual file lines instad of 1k
		if err != nil {
			return nil, err
		}
		diffWords := strings.TrimSpace(b.Buffer.String())

		diff, err := Parse(diffLines, diffWords)
		if err != nil {
			return nil, err
		}

		diffs = append(diffs, diff)
	}
	return diffs, nil
}
