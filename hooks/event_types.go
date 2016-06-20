package hooks

import (
	"encoding/json"
	"log"

	"github.com/google/go-github/github"
)

type EventType string

var (
	CommitCommentEvent            EventType = "commit_comment"
	CreateEvent                   EventType = "create"
	DeleteEvent                   EventType = "delete"
	DeploymentEvent               EventType = "deployment"
	DeploymentStatusEvent         EventType = "deployment_status"
	ForkEvent                     EventType = "fork"
	GollumEvent                   EventType = "gollum"
	IssueCommentEvent             EventType = "issue_comment"
	IssuesEvent                   EventType = "issues"
	MemberEvent                   EventType = "member"
	MembershipEvent               EventType = "membership"
	PageBuildEvent                EventType = "page_build"
	PublicEvent                   EventType = "public"
	PullRequestEvent              EventType = "pull_request"
	PullRequestReviewCommentEvent EventType = "pull_request_review_comment"
	PushEvent                     EventType = "push"
	ReleaseEvent                  EventType = "release"
	RepositoryEvent               EventType = "repository"
	StatusEvent                   EventType = "status"
	TeamAddEvent                  EventType = "team_add"
	WatchEvent                    EventType = "watch"

	pingEvent EventType = "ping"
)

func (e EventType) String() string {
	return string(e)
}

type pingEventPayload struct {
	Zen string `json:"zen"`
}

// structFromPayload creates the proper webhook event and unmarshals the payload data into it.
func structFromPayload(eventType string, payload []byte) (event interface{}) {
	switch EventType(eventType) {
	case CommitCommentEvent:
		event = &github.CommitCommentEvent{}
	case CreateEvent:
		event = &github.CreateEvent{}
	case DeleteEvent:
		event = &github.DeleteEvent{}
	case DeploymentEvent:
		event = &github.DeploymentEvent{}
	case DeploymentStatusEvent:
		event = &github.DeploymentStatusEvent{}
	case ForkEvent:
		event = &github.ForkEvent{}
	case GollumEvent:
		event = &github.GollumEvent{}
	case IssueCommentEvent:
		event = &github.IssueCommentEvent{}
	case IssuesEvent:
		event = &github.IssuesEvent{}
	case MemberEvent:
		event = &github.MemberEvent{}
	case MembershipEvent:
		event = &github.MembershipEvent{}
	case PageBuildEvent:
		event = &github.PageBuildEvent{}
	case pingEvent:
		event = &pingEventPayload{}
	case PublicEvent:
		event = &github.PublicEvent{}
	case PullRequestEvent:
		event = &github.PullRequestEvent{}
	case PullRequestReviewCommentEvent:
		event = &github.PullRequestReviewCommentEvent{}
	case PushEvent:
		event = &github.PushEvent{}
	case ReleaseEvent:
		event = &github.ReleaseEvent{}
	case RepositoryEvent:
		event = &github.RepositoryEvent{}
	case StatusEvent:
		event = &github.StatusEvent{}
	case TeamAddEvent:
		event = &github.TeamAddEvent{}
	case WatchEvent:
		event = &github.WatchEvent{}
	}
	if err := json.Unmarshal(payload, event); err != nil {
		log.Println("error unmarshalling %s event: %+v", eventType, err)
		panic(err.Error())
	}
	return event
}
