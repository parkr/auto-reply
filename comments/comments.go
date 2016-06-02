package comments

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

type CommentHandler func(context *ctx.Context, comment github.IssueCommentEvent) error

type CommentsHandler struct {
	context              *ctx.Context
	issueCommentHandlers []CommentHandler
	pullCommentHandlers  []CommentHandler
}

// NewHandler returns an HTTP handler which deprecates repositories
// by closing new issues with a comment directing attention elsewhere.
func NewHandler(context *ctx.Context, issuesHandlers []CommentHandler, pullRequestsHandlers []CommentHandler) *CommentsHandler {
	return &CommentsHandler{
		context:              context,
		issueCommentHandlers: issuesHandlers,
		pullCommentHandlers:  pullRequestsHandlers,
	}
}

func (h *CommentsHandler) HandlePayload(w http.ResponseWriter, r *http.Request, payload []byte) {
	if eventType := r.Header.Get("X-GitHub-Event"); !isComment(eventType) {
		log.Printf("received invalid event of type X-GitHub-Event: %s", eventType)
		http.Error(w, "not an issue_comment event.", 200)
		return
	}

	var event github.IssueCommentEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		log.Println("error unmarshalling IssueCommentEvent:", err)
		log.Println("payload:", payload)
		http.Error(w, "bad json", 400)
		return
	}

	var handlers []CommentHandler
	if isPullRequest(event) {
		handlers = h.pullCommentHandlers
	} else {
		handlers = h.issueCommentHandlers
	}

	for _, handler := range handlers {
		go handler(h.context, event)
	}

	fmt.Fprintf(w, "fired %d handlers", len(handlers))
}

func isPullRequest(event github.IssueCommentEvent) bool {
	return &event != nil && event.Issue != nil && event.Issue.PullRequestLinks != nil
}

func isComment(eventType string) bool {
	return eventType == "issue_comment" || eventType == "pull_request"
}
