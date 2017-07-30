package issuecomment

import (
	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/labeler"
)

const pendingFeedbackLabel = "pending-feedback"

func PendingFeedbackUnlabeler(context *ctx.Context, event interface{}) error {
	comment, ok := event.(*github.IssueCommentEvent)
	if !ok {
		return context.NewError("PendingFeedbackUnlabeler: not an issue comment event")
	}

	if senderAndCreatorEqual(comment) && hasLabel(comment.Issue.Labels, pendingFeedbackLabel) {
		owner, name, number := *comment.Repo.Owner.Login, *comment.Repo.Name, *comment.Issue.Number
		err := labeler.RemoveLabelIfExists(context, owner, name, number, pendingFeedbackLabel)
		if err != nil {
			return context.NewError("PendingFeedbackUnlabeler: error removing label on %s/%s#%d: %v", owner, name, number, err)
		}
	}

	return nil
}

func senderAndCreatorEqual(event *github.IssueCommentEvent) bool {
	return event.Sender != nil && event.Issue != nil && event.Issue.User != nil && *event.Sender.ID == *event.Issue.User.ID
}

func hasLabel(labels []github.Label, desiredLabel string) bool {
	for _, label := range labels {
		if *label.Name == desiredLabel {
			return true
		}
	}
	return false
}
