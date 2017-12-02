// autopull provides a webhook which will automatically create pull requests for a push to a special branch name prefix.
package autopull

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

type Handler struct {
	repos          []string
	acceptAllRepos bool
}

func (h *Handler) handlesRepo(repo string) bool {
	for _, handled := range h.repos {
		if handled == repo {
			return true
		}
	}
	return false
}

func (h *Handler) AddRepo(owner, name string) {
	h.repos = append(h.repos, owner+"/"+name)
}

func (h *Handler) AcceptAllRepos(newValue bool) {
	h.acceptAllRepos = newValue
}

func (h *Handler) CreatePullRequestFromPush(context *ctx.Context, event interface{}) error {
	push, ok := event.(*github.PushEvent)
	if !ok {
		return context.NewError("AutoPull: not an push event")
	}

	if strings.HasPrefix(*push.Ref, "refs/heads/pull/") && (h.acceptAllRepos || h.handlesRepo(*push.Repo.FullName)) {
		pr := newPRForPush(push)
		if pr == nil {
			return context.NewError("AutoPull: no commits for %s on %s/%s", *push.Ref, *push.Repo.Owner.Name, *push.Repo.Name)
		}

		pull, _, err := context.GitHub.PullRequests.Create(context.Context(), *push.Repo.Owner.Name, *push.Repo.Name, pr)
		if err != nil {
			return context.NewError(
				"AutoPull: error creating pull request for %s on %s/%s: %v",
				*push.Ref, *push.Repo.Owner.Name, *push.Repo.Name, err,
			)
		}
		log.Printf("created pull request: %s#%d", *push.Repo.FullName, *pull.Number)
	}

	return nil
}

func shortMessage(message string) string {
	return strings.SplitN(message, "\n", 1)[0]
}

// branchFromRef takes "refs/heads/pull/my-pull" and returns "pull/my-pull"
func branchFromRef(ref string) string {
	return strings.Replace(ref, "refs/heads/", "", 1)
}

func prBodyForPush(push *github.PushEvent) string {
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

func newPRForPush(push *github.PushEvent) *github.NewPullRequest {
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
