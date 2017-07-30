package labeler

import (
	"regexp"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var fixesIssueMatcher = regexp.MustCompile(`(?i)(?:Close|Closes|Closed|Fix|Fixes|Fixed|Resolve|Resolves|Resolved)\s+#(\d+)`)

func IssueHasPullRequestLabeler(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("IssueHasPullRequestLabeler: not a pull request event")
	}

	if *event.Action != "opened" {
		return nil
	}

	owner, repo, description := *event.Repo.Owner.Login, *event.Repo.Name, *event.PullRequest.Body

	issueNums := linkedIssues(description)
	if issueNums == nil {
		return nil
	}

	var err error
	for _, issueNum := range issueNums {
		err := AddLabels(context, owner, repo, issueNum, []string{"has-pull-request"})
		if err != nil {
			context.Log("error adding the has-pull-request label to %s/%s#%d: %v", owner, repo, issueNum, err)
		}
	}

	return err
}

func linkedIssues(description string) []int {
	issueSubmatches := fixesIssueMatcher.FindAllStringSubmatch(description, -1)
	if len(issueSubmatches) == 0 || len(issueSubmatches[0]) < 2 {
		return nil
	}

	issueNums := []int{}
	for _, match := range issueSubmatches {
		if len(match) < 2 {
			continue
		}

		if issueNum, err := strconv.Atoi(match[1]); err == nil {
			issueNums = append(issueNums, issueNum)
		}
	}

	return issueNums
}
