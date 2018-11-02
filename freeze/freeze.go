package freeze

import (
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var (
	TooOld    = time.Now().Add(-365 * 24 * time.Hour).Format("2006-01-02")
	LabelName = "frozen-due-to-age"
)

func AllTooOldIssues(context *ctx.Context, owner, repo string) ([]github.Issue, error) {
	issues := []github.Issue{}
	query := fmt.Sprintf("repo:%s/%s is:closed -label:%v updated:<=%s", owner, repo, LabelName, TooOld)
	opts := &github.SearchOptions{
		Sort:  "created",
		Order: "asc",
		ListOptions: github.ListOptions{
			PerPage: 500,
		},
	}

	for {
		result, resp, err := context.GitHub.Search.Issues(context.Context(), query, opts)
		if err != nil {
			return nil, err
		}

		if *result.Total == 0 {
			return issues, nil
		}

		issues = append(issues, result.Issues...)

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return issues, nil
}

func Freeze(context *ctx.Context, owner, repo string, issueNum int) error {
	_, err := context.GitHub.Issues.Lock(context.Context(), owner, repo, issueNum, nil)
	if err != nil {
		return err
	}
	_, _, err = context.GitHub.Issues.AddLabelsToIssue(context.Context(), owner, repo, issueNum, []string{LabelName})
	return err
}
