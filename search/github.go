package search

import (
	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/githubapi/githubsearch"
)

func GitHubIssues(context *ctx.Context, query githubsearch.IssueSearchParameters) ([]github.Issue, error) {
	issues := []github.Issue{}
	opts := &github.SearchOptions{Sort: "created", Order: "desc", ListOptions: github.ListOptions{Page: 0, PerPage: 100}}
	queryStr := query.String()
	for {
		result, resp, err := context.GitHub.Search.Issues(context.Context(), queryStr, opts)
		if err != nil {
			return nil, context.NewError("search: error running GitHub issues search query: '%s': %v", queryStr, err)
		}

		for _, issue := range result.Issues {
			issues = append(issues, issue)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return issues, nil
}
