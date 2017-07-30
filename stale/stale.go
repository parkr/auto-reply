package stale

import (
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/labeler"
)

var (
	defaultNonStaleableLabels = []string{
		"has-pull-request",
		"pinned",
		"security",
	}
)

type Configuration struct {
	// Whether to actuall perform the action. If false, just outputs what *would* happen.
	Perform bool

	// Labels, in addition to "stale", which should be mark an issue as already stale.
	StaleLabels []string

	// If an issue has this label, it is not stale.
	ExemptLabels []string

	// After this duration, the next action is performed (either mark or close).
	DormantDuration time.Duration

	// Comment to leave on a stale issue if being marked.
	// No comment is left if this is stale.
	NotificationComment *github.IssueComment
}

func MarkAndCloseForRepo(context *ctx.Context, config Configuration) error {
	if context.Repo.IsEmpty() {
		return context.NewError("stale: no repository present in context")
	}

	owner, name, nonStaleIssues, failedIssues := context.Repo.Owner, context.Repo.Name, 0, 0

	staleIssuesListOptions := &github.IssueListByRepoOptions{
		State:       "open",
		Sort:        "updated",
		Direction:   "asc",
		ListOptions: github.ListOptions{Page: 0, PerPage: 200},
	}

	allIssues := []*github.Issue{}

	for {
		issues, resp, err := context.GitHub.Issues.ListByRepo(context.Context(), owner, name, staleIssuesListOptions)
		if err != nil {
			return context.NewError("could not list issues for %s/%s: %v", owner, name, err)
		}

		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 {
			break
		}
		staleIssuesListOptions.ListOptions.Page = resp.NextPage
	}

	if len(allIssues) == 0 {
		context.Log("no issues for %s/%s", owner, name)
		return nil
	}

	for _, issue := range allIssues {
		if !IsStale(issue, config) {
			nonStaleIssues += 1
			continue
		}

		err := MarkOrCloseIssue(context, issue, config)
		if err != nil {
			context.Log("ERR %s !! failed marking or closing issue %d: %+v", context.Repo, *issue.Number, err)
			failedIssues += 1
		}
	}

	context.Log("INF %s -- ignored non-stale issues: %d", context.Repo, nonStaleIssues)
	if failedIssues > 0 {
		return context.NewError("ERR %s !! failed issues: %d", context.Repo, failedIssues)
	}

	return nil
}

func MarkOrCloseIssue(context *ctx.Context, issue *github.Issue, config Configuration) error {
	if context.Repo.IsEmpty() {
		return context.NewError("stale: no repository present in context")
	}

	if !IsStale(issue, config) {
		return context.NewError("stale: issue %s#%d is not stale", context.Repo, *issue.Number)
	}

	if hasStaleLabel(issue, config) {
		// Close!
		if config.Perform {
			context.Log("https://github.com/%s/issues/%d is being closed.", context.Repo, *issue.Number)
			return closeIssue(context, issue)
		} else {
			context.Log("https://github.com/%s/issues/%d would have been closed (dry-run).", context.Repo, *issue.Number)
		}
	} else {
		// Mark!
		if config.Perform {
			context.Log("https://github.com/%s/issues/%d is being marked.", context.Repo, *issue.Number)
			return markIssue(context, issue, config.NotificationComment)
		} else {
			context.Log("https://github.com/%s/issues/%d would have been marked (dry-run).", context.Repo, *issue.Number)
		}
	}

	return nil
}

func closeIssue(context *ctx.Context, issue *github.Issue) error {
	_, _, err := context.GitHub.Issues.Edit(
		context.Context(),
		context.Repo.Owner,
		context.Repo.Name,
		*issue.Number,
		&github.IssueRequest{State: github.String("closed")},
	)
	return err
}

func markIssue(context *ctx.Context, issue *github.Issue, comment *github.IssueComment) error {
	// Mark with "stale" label.
	err := labeler.AddLabels(context, context.Repo.Owner, context.Repo.Name, *issue.Number, []string{"stale"})
	if err != nil {
		return context.NewError("stale: couldn't mark issue as stale %s#%d: %+v", context.Repo, *issue.Number, err)
	}

	if comment != nil {
		// Leave comment.
		_, _, err := context.GitHub.Issues.CreateComment(
			context.Context(),
			context.Repo.Owner, context.Repo.Name, *issue.Number, comment)
		if err != nil {
			return context.NewError("stale: couldn't leave comment on %s#%d: %+v", context.Repo, *issue.Number, err)
		}
	}

	return nil
}

func IsStale(issue *github.Issue, config Configuration) bool {
	return issue.PullRequestLinks == nil &&
		!isUpdatedWithinDuration(issue, config) &&
		excludesNonStaleableLabels(issue, config)
}

func isUpdatedWithinDuration(issue *github.Issue, config Configuration) bool {
	return (*issue.UpdatedAt).Unix() >= time.Now().Add(-config.DormantDuration).Unix()
}

// Returns true if none of the exempt labels are present, false if at least one exempt label is present.
func excludesNonStaleableLabels(issue *github.Issue, config Configuration) bool {
	if len(issue.Labels) == 0 {
		return true
	}

	for _, exemptLabel := range config.ExemptLabels {
		for _, issueLabel := range issue.Labels {
			if *issueLabel.Name == exemptLabel {
				return false
			}
		}
	}

	return true
}

func hasStaleLabel(issue *github.Issue, config Configuration) bool {
	if issue.Labels == nil {
		return false
	}

	for _, issueLabel := range issue.Labels {
		if *issueLabel.Name == "stale" {
			return true
		}
		for _, staleLabel := range config.StaleLabels {
			if *issueLabel.Name == staleLabel {
				return true
			}
		}
	}

	return false
}
