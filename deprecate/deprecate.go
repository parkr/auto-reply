package deprecate

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/parkr/auto-reply/Godeps/_workspace/src/github.com/google/go-github/github"
)

const (
	closedState = `closed`
)

type RepoDeprecation struct {
	// Name with organization, e.g. "jekyll/jekyll-help"
	Nwo string

	// Comment to send when closing the issue.
	Message string
}

type DeprecateHandler struct {
	client   *github.Client
	repos    []RepoDeprecation
	messages map[string]string
}

func deprecationsToMap(deprecations []RepoDeprecation) map[string]string {
	deps := map[string]string{}
	for _, dep := range deprecations {
		deps[dep.Nwo] = dep.Message
	}
	return deps
}

// NewHandler returns an HTTP handler which deprecates repositories
// by closing new issues with a comment directing attention elsewhere.
func NewHandler(client *github.Client, deprecations []RepoDeprecation) *DeprecateHandler {
	return &DeprecateHandler{
		client:   client,
		repos:    deprecations,
		messages: deprecationsToMap(deprecations),
	}
}

func (dh *DeprecateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var issue github.IssueActivityEvent
	err := json.NewDecoder(r.Body).Decode(&issue)
	if err != nil {
		log.Println("error unmarshalling issue stuffs:", err)
		http.Error(w, "bad json", 400)
		return
	}

	if *issue.Action != "opened" {
		http.Error(w, "ignored", 200)
		return
	}

	if msg, ok := dh.messages[*issue.Repo.FullName]; ok {
		err = dh.leaveComment(issue, msg)
		if err != nil {
			log.Println("error leaving comment:", err)
			http.Error(w, "couldnt leave comment", 500)
			return
		}
		err = dh.closeIssue(issue)
		if err != nil {
			log.Println("error closing comment:", err)
			http.Error(w, "couldnt close comment", 500)
			return
		}
	} else {
		log.Printf("looks like '%s' repo isn't deprecated", *issue.Repo.FullName)
		http.Error(w, "non-deprecated repo", 404)
		return
	}

	w.Write([]byte(`sorry ur deprecated`))
}

func (dh *DeprecateHandler) leaveComment(issue github.IssueActivityEvent, msg string) error {
	_, _, err := dh.client.Issues.CreateComment(
		*issue.Repo.Owner.Login,
		*issue.Repo.Name,
		*issue.Issue.Number,
		&github.IssueComment{Body: github.String(msg)},
	)
	return err
}

func (dh *DeprecateHandler) closeIssue(issue github.IssueActivityEvent) error {
	_, _, err := dh.client.Issues.Edit(
		*issue.Repo.Owner.Login,
		*issue.Repo.Name,
		*issue.Issue.Number,
		&github.IssueRequest{State: github.String(closedState)},
	)
	return err
}
