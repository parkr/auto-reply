package labeler

import (
	"log"
	"regexp"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var fixesIssueMatcher = regexp.MustCompile(`Fixes #(\d+)`)

var IssueHasPullRequestLabeler = func(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("IssueHasPullRequestLabeler: not a pull request event")
	}

	if *event.Action != "opened" {
		return nil
	}

        description := *event.PullRequest.Body
	issueNum := issueFixed(description)

	var err error
	if issueNum != "" {
		log.Printf("detected a pull request that fixes issue %v", issueNum)
	}

	return err
}

func issueFixed(description string) string {
	issueSubmatches := fixesIssueMatcher.FindAllStringSubmatch(description, -1)
	if len(issueSubmatches) == 0 || len(issueSubmatches[0]) < 2 {
		return ""
	}

	return issueSubmatches[0][1]
}
