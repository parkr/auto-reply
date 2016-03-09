package autopull

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/common"
)

var (
	baseBranch *string = github.String("master")
)

type AutoPullHandler struct {
	client *github.Client
	repos  map[string]bool
}

func shortMessage(message string) string {
	return strings.SplitN(message, "\n", 1)[0]
}

// branchFromRef takes "refs/heads/pull/my-pull" and returns "pull/my-pull"
func branchFromRef(ref string) string {
	return strings.Replace(ref, "refs/heads/", "", 1)
}

func prBodyForPush(push github.PushEvent) string {
	var mention string
	if author := push.Commits[0].Author; author != nil {
		if author.Login != nil {
			mention = *author.Login
		} else {
			mention = *author.Name
		}
	} else {
		mention = "unknown"
	}
	return fmt.Sprintf(
		"PR automatically created for @%s.\n\n```text\n%s\n```",
		mention,
		*push.Commits[0].Message,
	)
}

func newPRForPush(push github.PushEvent) *github.NewPullRequest {
	if push.Commits == nil || len(push.Commits) == 0 {
		return nil
	}
	return &github.NewPullRequest{
		Title: github.String(shortMessage(*push.Commits[0].Message)),
		Head:  github.String(branchFromRef(*push.Ref)),
		Base:  github.String("master"),
		Body:  github.String(prBodyForPush(push)),
	}
}

func (h *AutoPullHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-GitHub-Event") != "push" {
		http.Error(w, "pong. ignored this one.", 200)
		return
	}

	var push github.PushEvent
	err := json.NewDecoder(r.Body).Decode(&push)
	if err != nil {
		log.Println("error unmarshalling issue stuffs:", err)
		body, err := ioutil.ReadAll(r.Body)
		log.Println("ioutil.ReadAll:", body, err)
		http.Error(w, "bad json", 400)
		return
	}

	if os.Getenv("AUTO_REPLY_DEBUG") == "true" {
		log.Println("received push:", push)
	}
	if _, ok := h.repos[*push.Repo.FullName]; ok && strings.HasPrefix(*push.Ref, "refs/heads/pull/") {
		pr := newPRForPush(push)
		if pr == nil {
			http.Error(w, "ignoring", 200)
			return
		}

		pull, _, err := h.client.PullRequests.Create(*push.Repo.Owner.Name, *push.Repo.Name, pr)
		if err != nil {
			log.Printf("error creating pull request for %s/%s: %v", *push.Repo.Owner.Name, *push.Repo.Name, err)
			http.Error(w, "pr could not be created", 500)
			return
		}

		http.Error(w, "pr created: "+*pull.HTMLURL, 201)
	} else {
		http.Error(w, "ignoring due to bad ref or repo", 200)
	}
}

func NewHandler(client *github.Client, handledRepos []string) *AutoPullHandler {
	return &AutoPullHandler{
		client: client,
		repos:  common.SliceLookup(handledRepos),
	}
}
