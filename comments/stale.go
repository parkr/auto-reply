package comments

import (
	"log"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/labeler"
)

var StaleUnlabeler = func(context *ctx.Context, event github.IssueCommentEvent) error {
	if *event.Action != "created" {
		return nil
	}

	owner, repo, num := *event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number
	err := labeler.RemoveLabelIfExists(context.GitHub, owner, repo, num, "stale")
	if err != nil {
		log.Printf("error removing the pending-rebase label: %v", err)
	}

	return err
}
