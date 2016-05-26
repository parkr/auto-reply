package labeler

import (
	"log"

	"github.com/google/go-github/github"
)

var PendingRebasePRLabeler = func(client *github.Client, event github.PullRequestEvent) error {
	if *event.Action != "synchronize" {
		return nil
	}

	owner, repo := *event.Repo.Owner.Login, *event.Repo.Name
	err := RemoveLabel(client, owner, repo, *event.Number, "pending-rebase")
	if err != nil {
		log.Println("error removing the pending-rebase label: %v", err)
	}
	return err
}
