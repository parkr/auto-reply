package labeler

import (
	"fmt"
	"log"

	"github.com/parkr/auto-reply/common"
	"github.com/parkr/auto-reply/ctx"
)

func AddLabels(context *ctx.Context, owner, repo string, number int, labels []string) error {
	_, res, err := context.GitHub.Issues.AddLabelsToIssue(context.Context(), owner, repo, number, labels)
	return common.ErrorFromResponse(res, err)
}

func RemoveLabel(context *ctx.Context, owner, repo string, number int, label string) error {
	res, err := context.GitHub.Issues.RemoveLabelForIssue(context.Context(), owner, repo, number, label)
	return common.ErrorFromResponse(res, err)
}

func RemoveLabels(context *ctx.Context, owner, repo string, number int, labels []string) error {
	var anyError error
	for _, label := range labels {
		err := RemoveLabel(context, owner, repo, number, label)
		if err != nil {
			anyError = err
			log.Printf("couldn't remove label '%s' from %s/%s#%d: %v", label, owner, repo, number, err)
		}
	}
	return anyError
}

func RemoveLabelIfExists(context *ctx.Context, owner, repo string, number int, label string) error {
	if IssueHasLabel(context, owner, repo, number, label) {
		return RemoveLabel(context, owner, repo, number, label)
	}
	return fmt.Errorf("%s/%s#%d doesn't have label: %s", owner, repo, number, label)
}

func IssueHasLabel(context *ctx.Context, owner, repo string, number int, label string) bool {
	labels, res, err := context.GitHub.Issues.ListLabelsByIssue(context.Context(), owner, repo, number, nil)
	if err = common.ErrorFromResponse(res, err); err != nil {
		return false
	}

	for _, issueLabel := range labels {
		if *issueLabel.Name == label {
			return true
		}
	}
	return false
}
