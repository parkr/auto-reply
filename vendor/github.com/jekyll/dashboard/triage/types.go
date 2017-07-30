package triage

import (
	"time"

	"github.com/google/go-github/github"
)

type IssueGrouping struct {
	Label  string
	Issues Issues

	lastUpdated time.Time
}

type Issues []github.Issue

func (g Issues) Len() int {
	return len(g)
}
func (g Issues) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}
func (g Issues) Less(i, j int) bool {
	return g[i].GetCreatedAt().Before(g[j].GetCreatedAt())
}
