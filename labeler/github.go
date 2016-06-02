package labeler

import (
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
