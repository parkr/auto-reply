package labeler

import (
	"log"
	"regexp"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var fixesIssueMatcher = regexp.MustCompile(`(?i)(?:Close|Closes|Closed|Fix|Fixes|Fixed|Resolve|Resolves|Resolved) #(\d+)`)

var IssueHasPullRequestLabeler = func(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("IssueHasPullRequestLabeler: not a pull request event")
	}

	if *event.Action != "opened" {
		return nil
	}

	owner, repo, description := *event.Repo.Owner.Login, *event.Repo.Name, *event.PullRequest.Body

	issueNum := issueFixed(description)

	var err error
	if issueNum != -1 {
		err := AddLabels(context.GitHub, owner, repo, issueNum, []string{"has-pull-request"})
		if err != nil {
			log.Printf("error adding the has-pull-request label: %v", err)
		}
	}

	return err
}

func issueFixed(description string) int {
	issueSubmatches := fixesIssueMatcher.FindAllStringSubmatch(description, -1)
	if len(issueSubmatches) == 0 || len(issueSubmatches[0]) < 2 {
		return -1
	}

	issueNum, _ := strconv.Atoi(issueSubmatches[0][1])
	return issueNum
}
