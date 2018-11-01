package dependencies

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/search"
	"github.com/parkr/githubapi/githubsearch"
)

func containsSubstring(s, substring string) bool {
	return strings.Index(strings.ToLower(s), strings.ToLower(substring)) >= 0
}

func GitHubUpdateIssueForDependency(context *ctx.Context, repoOwner, repoName string, dependency Dependency) *github.Issue {
	query := githubsearch.IssueSearchParameters{
		Repository: &githubsearch.RepositoryName{
			Owner: repoOwner,
			Name:  repoName,
		},
		State: githubsearch.Open,
		Scope: githubsearch.TitleScope,
		Query: fmt.Sprintf("update %s v%s", dependency.GetName(), dependency.GetLatestVersion(context)),
	}

	issues, err := search.GitHubIssues(context, query)
	if err != nil {
		context.Log("dependencies: couldn't search github: %+v", err)
		return nil
	}

	for _, issue := range issues {
		if containsSubstring(*issue.Title, "update") && containsSubstring(*issue.Title, dependency.GetName()) {
			return &issue
		}
	}

	return nil
}

func FileGitHubIssueForDependency(context *ctx.Context, repoOwner, repoName string, dependency Dependency) (*github.Issue, error) {
	issue, _, err := context.GitHub.Issues.Create(context.Context(), repoOwner, repoName, &github.IssueRequest{
		Title: github.String(fmt.Sprintf(
			"Update dependency constraint to allow for %s v%s",
			dependency.GetName(), dependency.GetLatestVersion(context),
		)),
		Body: github.String(fmt.Sprintf(
			"Hey there! :wave:\n\nI noticed that the constraint you have for %s doesn't allow for the latest version to be used.\n\nThe constraint I found was `%s`, and the latest version available is `%s`.\n\nCan you look into updating that constraint so our users can use the latest and greatest version? Thanks! :revolving_hearts:",
			dependency.GetName(), dependency.GetConstraint(), dependency.GetLatestVersion(context),
		)),
		Labels: &[]string{"help-wanted", "dependency"},
	})
	return issue, err
}
