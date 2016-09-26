// lgtm contains the functionality to handle approval from maintainers.
package lgtm

import (
	"fmt"
	"regexp"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/auth"
	"github.com/parkr/auto-reply/ctx"
)

var lgtmBodyRegexp = regexp.MustCompile(`(?i:\ALGTM[!.,]\s+|\s+LGTM[.!,]*\z|\ALGTM[.!,]*\z)`)

type prRef struct {
	Repo   Repo
	Number int
}

func (r prRef) String() string {
	return fmt.Sprintf("%s/%s#%d", r.Repo.Owner, r.Repo.Name, r.Number)
}

type Repo struct {
	Owner, Name string
	// The number of LGTM's a PR must get before going state: "success"
	Quorum int
}

type Handler struct {
	repos []Repo
}

func (h *Handler) AddRepo(owner, name string, quorum int) {
	if repo := h.findRepo(owner, name); repo != nil {
		repo.Quorum = quorum
	} else {
		h.repos = append(h.repos, Repo{
			Owner:  owner,
			Name:   name,
			Quorum: quorum,
		})
	}
}

func (h *Handler) findRepo(owner, name string) *Repo {
	for _, repo := range h.repos {
		if repo.Owner == owner && repo.Name == name {
			return &repo
		}
	}

	return nil
}

func (h *Handler) isEnabledFor(owner, name string) bool {
	return h.findRepo(owner, name) != nil
}

func (h *Handler) newPRRef(owner, name string, number int) prRef {
	repo := h.findRepo(owner, name)
	if repo != nil {
		return prRef{
			Repo:   *repo,
			Number: number,
		}
	}
	return prRef{
		Repo:   Repo{Owner: owner, Name: name, Quorum: 0},
		Number: number,
	}
}

func (h *Handler) IssueCommentHandler(context *ctx.Context, payload interface{}) error {
	comment, ok := payload.(*github.IssueCommentEvent)
	if !ok {
		return context.NewError("lgtm.IssueCommentHandler: not an issue comment event")
	}

	// LGTM comment?
	if !lgtmBodyRegexp.MatchString(*comment.Comment.Body) {
		return context.NewError("lgtm.IssueCommentHandler: not a LGTM comment")
	}

	// Is this a pull request?
	if comment.Issue == nil || comment.Issue.PullRequestLinks == nil {
		return context.NewError("lgtm.IssueCommentHandler: not a pull request")
	}

	ref := h.newPRRef(*comment.Repo.Owner.Login, *comment.Repo.Name, *comment.Issue.Number)
	lgtmer := *comment.Comment.User.Login

	if !h.isEnabledFor(ref.Repo.Owner, ref.Repo.Name) {
		return context.NewError("lgtm.IssueCommentHandler: not enabled for %s/%s", ref.Repo.Owner, ref.Repo.Name)
	}

	// Does the user have merge/label abilities?
	if !auth.CommenterHasPushAccess(context, *comment) {
		return context.NewError(
			"%s isn't authenticated to merge anything on %s/%s",
			*comment.Comment.User.Login, ref.Repo.Owner, ref.Repo.Name)
	}

	// Get status
	info, err := getStatus(context, ref)
	if err != nil {
		return context.NewError("lgtm.IssueCommentHandler: couldn't get status for %s: %v", ref, err)
	}

	// Already LGTM'd by you? Exit.
	if info.IsLGTMer(lgtmer) {
		return context.NewError(
			"lgtm.IssueCommentHandler: no duplicate LGTM allowed for @%s on %s", lgtmer, ref)
	}

	info.lgtmers = append(info.lgtmers, "@"+lgtmer)
	if err := setStatus(context, ref, info.sha, info); err != nil {
		return context.NewError(
			"lgtm.IssueCommentHandler: had trouble adding lgtmer '%s' on %s: %v",
			lgtmer, ref, err)
	}
	return nil
}

func (h *Handler) PullRequestHandler(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("lgtm.PullRequestHandler: not a pull request event")
	}

	ref := h.newPRRef(*event.Repo.Owner.Login, *event.Repo.Name, *event.Number)

	if !h.isEnabledFor(ref.Repo.Owner, ref.Repo.Name) {
		return context.NewError("lgtm.PullRequestHandler: not enabled for %s", ref)
	}

	if *event.Action == "opened" || *event.Action == "synchronize" {
		err := setStatus(context, ref, *event.PullRequest.Head.SHA, &statusInfo{
			lgtmers: []string{},
			quorum:  ref.Repo.Quorum,
			sha:     *event.PullRequest.Head.SHA,
		})
		if err != nil {
			return context.NewError(
				"lgtm.PullRequestHandler: could not create status on %s: %v",
				ref, err,
			)
		}
	}

	return nil
}

func (h *Handler) PullRequestReviewHandler(context *ctx.Context, payload interface{}) error {
	return context.NewError("lgtm.PullRequestReviewHandler: pull request review webhooks aren't implemented yet")

	//event, ok := payload.(*github.PullRequestReviewEvent)
	//if !ok {
	//	return context.NewError("lgtm.PullRequestReviewHandler: not a pull request review event")
	//}

	//ref := h.newPRRef(*event.Repo.Owner.Login, *event.Repo.Name, *event.Number)

	//if !h.isEnabledFor(ref.Repo.Owner, ref.Repo.Name) {
	//	return context.NewError("lgtm.PullRequestReviewHandler: not enabled for %s", ref)
	//}
}
