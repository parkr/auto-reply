package labeler

import (
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/common"
)

func AddLabels(client *github.Client, owner, repo string, number int, labels []string) error {
	_, res, err := client.Issues.AddLabelsToIssue(owner, repo, number, labels)
	return common.ErrorFromResponse(res, err)
}

func RemoveLabel(client *github.Client, owner, repo string, number int, label string) error {
	res, err := client.Issues.RemoveLabelForIssue(owner, repo, number, label)
	return common.ErrorFromResponse(res, err)
}

func RemoveLabels(client *github.Client, owner, repo string, number int, labels []string) error {
	var anyError error
	for _, label := range labels {
		err := RemoveLabel(client, owner, repo, number, label)
		if err != nil {
			anyError = err
			log.Printf("couldn't remove label '%s' from %s/%s#%d: %v", label, owner, repo, number, err)
		}
	}
	return anyError
}

func RemoveLabelIfExists(client *github.Client, owner, repo string, number int, label string) error {
	if IssueHasLabel(client, owner, repo, number, label) {
		return RemoveLabel(client, owner, repo, number, label)
	}
	return fmt.Errorf("%s/%s#%d doesn't have label: %s", owner, repo, number, label)
}

func IssueHasLabel(client *github.Client, owner, repo string, number int, label string) bool {
	labels, res, err := client.Issues.ListLabelsByIssue(owner, repo, number, nil)
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
