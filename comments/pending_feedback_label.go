package comments

import (
	"log"

	"github.com/google/go-github/github"
)

var (
	pendingFeedbackLabel = "pending-feedback"

	HandlerPendingFeedbackLabel = func(client *github.Client, event github.IssueCommentEvent) error {
		// if the comment is from the issue author & issue has the "pending-feedback", remove the label

		if os.Getenv("AUTO_REPLY_DEBUG") == "true" {
			log.Println("received event:", event)
		}

		if hasLabel(event.Issue.Labels, pendingFeedbackLabel) && event.Sender.ID == event.Issue.User.ID {
			owner, name, number := *event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number
			_, err := client.Issues.RemoveLabelForIssue(owner, name, number, pendingFeedbackLabel)
			if err != nil {
				log.Printf("[pending_feedback_label]: error removing label (%s/%s#%d): %v", owner, name, number, err)
				return err
			}
			return nil
		}
		return nil
	}
)

func hasLabel(labels []github.Label, desiredLabel string) bool {
	for _, label := range labels {
		if *label.Name == desiredLabel {
			return true
		}
	}
	return false
}
