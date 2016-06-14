package hooks

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

// HookHandler describes the interface for any type which can handle a webhook payload.
type HookHandler interface {
	HandlePayload(w http.ResponseWriter, r *http.Request, payload []byte)
}

// EventHandler is An event handler takes in a given event and operates on it.
type EventHandler func(context *ctx.Context, event interface{}) error

// structFromPayload creates the proper webhook event and unmarshals the payload data into it.
func structFromPayload(eventType string, payload []byte) (event interface{}) {
	switch eventType {
	case "CommitCommentEvent", "commit_comment":
		event = &github.CommitCommentEvent{}
	case "CreateEvent", "create":
		event = &github.CreateEvent{}
	case "DeleteEvent", "delete":
		event = &github.DeleteEvent{}
	case "DeploymentEvent", "deployment":
		event = &github.DeploymentEvent{}
	case "DeploymentStatusEvent", "deployment_status":
		event = &github.DeploymentStatusEvent{}
	case "ForkEvent", "fork":
		event = &github.ForkEvent{}
	case "GollumEvent", "gollum":
		event = &github.GollumEvent{}
	case "IssueCommentEvent", "issue_comment":
		event = &github.IssueCommentEvent{}
	case "IssuesEvent", "issues":
		event = &github.IssuesEvent{}
	case "MemberEvent", "member":
		event = &github.MemberEvent{}
	case "MembershipEvent", "membership":
		event = &github.MembershipEvent{}
	case "PageBuildEvent", "page_build":
		event = &github.PageBuildEvent{}
	case "PublicEvent", "public":
		event = &github.PublicEvent{}
	case "PullRequestEvent", "pull_request":
		event = &github.PullRequestEvent{}
	case "PullRequestReviewCommentEvent", "pull_request_review_comment":
		event = &github.PullRequestReviewCommentEvent{}
	case "PushEvent", "push":
		event = &github.PushEvent{}
	case "ReleaseEvent", "release":
		event = &github.ReleaseEvent{}
	case "RepositoryEvent", "repository":
		event = &github.RepositoryEvent{}
	case "StatusEvent", "status":
		event = &github.StatusEvent{}
	case "TeamAddEvent", "team_add":
		event = &github.TeamAddEvent{}
	case "WatchEvent", "watch":
		event = &github.WatchEvent{}
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Println("error unmarshalling %s event: %+v", eventType, err)
		panic(err.Error())
	}
	return payload
}
