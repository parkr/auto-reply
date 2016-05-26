package labeler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/github"
)

type PushHandler func(client *github.Client, event github.PushEvent) error
type PullRequestHandler func(client *github.Client, event github.PullRequestEvent) error

type LabelerHandler struct {
	client              *github.Client
	pushHandlers        []PushHandler
	pullRequestHandlers []PullRequestHandler
}

// NewHandler returns an HTTP handler which deprecates repositories
// by closing new issues with a comment directing attention elsewhere.
func NewHandler(client *github.Client, pushHandlers []PushHandler, pullRequestHandlers []PullRequestHandler) *LabelerHandler {
	return &LabelerHandler{
		client:              client,
		pushHandlers:        pushHandlers,
		pullRequestHandlers: pullRequestHandlers,
	}
}

func (h *LabelerHandler) HandlePayload(w http.ResponseWriter, r *http.Request, payload []byte) {
	eventType := r.Header.Get("X-GitHub-Event")

	switch eventType {
	case "pull_request":
		var event github.PullRequestEvent
		err := json.Unmarshal(payload, &event)
		if err != nil {
			log.Println("error unmarshalling pull request event:", err)
			http.Error(w, "bad json", 400)
			return
		}
		for _, handler := range h.pullRequestHandlers {
			go handler(h.client, event)
		}
		fmt.Fprintf(w, "fired %d handlers", len(h.pullRequestHandlers))

	case "push":
		var event github.PushEvent
		err := json.Unmarshal(payload, &event)
		if err != nil {
			log.Println("error unmarshalling pull request event:", err)
			http.Error(w, "bad json", 400)
			return
		}
		for _, handler := range h.pushHandlers {
			go handler(h.client, event)
		}
		fmt.Fprintf(w, "fired %d handlers", len(h.pushHandlers))

	default:
		log.Printf("received invalid event of type X-GitHub-Event: %s", eventType)
		http.Error(w, "not an issue_comment event.", 200)
	}
}
