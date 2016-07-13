package lgtm

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/auth"
	"github.com/parkr/auto-reply/ctx"
)

const lgtmContext = "jekyll/lgtm"

var (
	lgtmBodyRegexp = regexp.MustCompile(`(?i:\ALGTM\s+|\s+LGTM\.?\z|\ALGTM\.?\z)`)

	statusCache = statusMap{data: make(map[string]*statusInfo)}
)

type statusMap struct {
	sync.Mutex // protects data
	data       map[string]*statusInfo
}

type prRef struct {
	Owner, Name string
	Number      int
}

func (r prRef) String() string {
	return fmt.Sprintf("%s/%s#%d", r.Owner, r.Name, r.Number)
}

func IssueCommentHandler(context *ctx.Context, payload interface{}) error {
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

	ref := prRef{*comment.Comment.User.Login, *comment.Repo.FullName, *comment.Issue.Number}
	lgtmer := *comment.Comment.User.Login

	// Does the user have merge/label abilities?
	if !auth.CommenterHasPushAccess(context, *comment) {
		return context.NewError(
			"%s isn't authenticated to merge anything on %s/%s",
			*comment.Comment.User.Login, ref.Owner, ref.Name)
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

func PullRequestHandler(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("lgtm.PullRequestHandler: not a pull request event")
	}

	owner, name, number := *event.Repo.Owner.Login, *event.Repo.Name, *event.Number

	if *event.Action == "opened" {
		_, _, err := context.GitHub.Repositories.CreateStatus(
			owner, name, *event.PullRequest.Head.SHA,
			newEmptyStatus(),
		)
		if err != nil {
			return context.NewError(
				"lgtm.PullRequestHandler: could not create status on %s/%s#%d: %v",
				owner, name, number, err,
			)
		}
	}

	return nil
}

func setStatus(context *ctx.Context, ref prRef, sha string, status *statusInfo) error {
	_, _, err := context.GitHub.Repositories.CreateStatus(ref.Owner, ref.Name, sha, status.NewStatus())
	if err != nil {
		return err
	}

	statusCache.Lock()
	statusCache.data[ref.String()] = status
	statusCache.Unlock()

	return nil
}

func getStatus(context *ctx.Context, ref prRef) (*statusInfo, error) {
	statusCache.Lock()
	cachedStatus, ok := statusCache.data[ref.String()]
	statusCache.Unlock()
	if ok && cachedStatus != nil {
		return cachedStatus, nil
	}

	pr, _, err := context.GitHub.PullRequests.Get(ref.Owner, ref.Name, ref.Number)
	if err != nil {
		return nil, err
	}

	statuses, _, err := context.GitHub.Repositories.ListStatuses(ref.Owner, ref.Name, *pr.Head.SHA, nil)
	if err != nil {
		return nil, err
	}

	var preExistingStatus *github.RepoStatus
	var info *statusInfo
	for _, status := range statuses {
		if *status.Context == lgtmContext {
			preExistingStatus = status
			info = parseStatus(*pr.Head.SHA, status)
			break
		}
	}

	if preExistingStatus == nil {
		preExistingStatus = newEmptyStatus()
		info = parseStatus(*pr.Head.SHA, preExistingStatus)
		setStatus(context, ref, *pr.Head.SHA, info)
	}

	statusCache.Lock()
	statusCache.data[ref.String()] = info
	statusCache.Unlock()

	return info, nil
}

func newEmptyStatus() *github.RepoStatus {
	return &github.RepoStatus{
		Context:     github.String(lgtmContext),
		State:       github.String("failure"),
		Description: github.String("This pull request has not received any LGTM's."),
	}
}
