package lgtm

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/auth"
	"github.com/parkr/auto-reply/ctx"
)

const lgtmContext = "jekyll/lgtm"

var (
	lgtmBodyRegexp = regexp.MustCompile(`(?i:\ALGTM\s+|\s+LGTM\.?\z|\ALGTM\.?\z)`)

	statusCache = statusMap{data: make(map[string]*github.RepoStatus)}
)

type statusMap struct {
	sync.Mutex // protects data
	data       map[string]*github.RepoStatus
}

type statusInfo struct {
	lgtmers    []string
	repoStatus *github.RepoStatus
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
	status, err := getStatus(context, ref)
	if err != nil {
		return context.NewError("lgtm.IssueCommentHandler: couldn't get status for %s: %v", ref, err)
	}

	statusInfo, err := parseStatus(status)
	if err != nil {
		return context.NewError("lgtm.IssueCommentHandler: couldn't parse status for %s: %v", ref, err)
	}

	// Already LGTM'd by you? Exit.
	if statusInfo.IsLGTMer(lgtmer) {
		return context.NewError(
			"lgtm.IssueCommentHandler: no duplicate LGTM allowed for @%s on %s", lgtmer, ref)
	}

	newStatus := "add lgtmer and generate a new status with that info."
	if err := setStatus(); err != nil {
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

func setStatus(context *ctx.Context, ref prRef, sha string, status *github.RepoStatus) error {
	newStatus, _, err := context.GitHub.Repositories.CreateStatus(ref.Owner, ref.Name, sha, status)
	if err != nil {
		return err
	}

	statusCache.Lock()
	statusCache.data[ref.String()] = newStatus
	statusCache.Unlock()

	return nil
}

func getStatus(context *ctx.Context, ref prRef) (*github.RepoStatus, error) {
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
	for _, status := range statuses {
		if *status.Context == lgtmContext {
			preExistingStatus = status
			break
		}
	}

	if preExistingStatus == nil {
		preExistingStatus = newEmptyStatus()
		setStatus(context, ref, *pr.Head.SHA, preExistingStatus)
	}

	statusCache.Lock()
	statusCache.data[ref.String()] = preExistingStatus
	statusCache.Unlock()

	return preExistingStatus, nil
}

func parseStatus(repoStatus *github.RepoStatus) (statusInfo, error) {
	return statusInfo{}, fmt.Errorf("lgtm.parseStatus: not implemented yet")
}

func generateDescription(lgtmers []string) string {
	switch len(lgtmers) {
	case 0:
		return "This pull request has not received any LGTM's."
	case 1:
		return fmt.Sprintf("%s has approved this PR.", lgtmers[0])
	case 2:
		return fmt.Sprintf("%s and %s have approved this PR.", lgtmers[0], lgtmers[1])
	default:
		lastIndex := len(lgtmers) - 1
		return fmt.Sprintf("%s, and %s have approved this PR.",
			strings.Join(lgtmers[0:lastIndex], ","), lgtmers[lastIndex])
	}
}

func newEmptyStatus() *github.RepoStatus {
	return &github.RepoStatus{
		Context:     github.String(lgtmContext),
		State:       github.String("failure"),
		Description: github.String("This pull request has not received any LGTM's."),
	}
}
