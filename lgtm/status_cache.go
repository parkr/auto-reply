package lgtm

import (
	"fmt"
	"sync"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var statusCache = statusMap{data: make(map[string]*statusInfo)}

type statusMap struct {
	sync.Mutex // protects 'data'
	data       map[string]*statusInfo
}

func lgtmContext(owner string) string {
	return owner + "/lgtm"
}

func setStatus(context *ctx.Context, ref prRef, sha string, status *statusInfo) error {
	_, _, err := context.GitHub.Repositories.CreateStatus(
		context.Context(), ref.Repo.Owner, ref.Repo.Name, sha, status.NewRepoStatus(ref.Repo.Owner))
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

	pr, _, err := context.GitHub.PullRequests.Get(context.Context(), ref.Repo.Owner, ref.Repo.Name, ref.Number)
	if err != nil {
		return nil, err
	}

	statuses, _, err := context.GitHub.Repositories.ListStatuses(context.Context(), ref.Repo.Owner, ref.Repo.Name, *pr.Head.SHA, nil)
	if err != nil {
		return nil, err
	}

	var preExistingStatus *github.RepoStatus
	var info *statusInfo
	// Find the status matching context.
	neededContext := lgtmContext(ref.Repo.Owner)
	for _, status := range statuses {
		if *status.Context == neededContext {
			preExistingStatus = status
			info = parseStatus(*pr.Head.SHA, status)
			break
		}
	}

	// None of the contexts matched.
	if preExistingStatus == nil {
		preExistingStatus = newEmptyStatus(ref.Repo.Owner, ref.Repo.Quorum)
		info = parseStatus(*pr.Head.SHA, preExistingStatus)
		err := setStatus(context, ref, *pr.Head.SHA, info)
		if err != nil {
			fmt.Printf("getStatus: couldn't save new empty status to %s for %s: %v\n", ref, *pr.Head.SHA, err)
		}
	}

	if ref.Repo.Quorum != 0 {
		info.quorum = ref.Repo.Quorum
	}

	statusCache.Lock()
	statusCache.data[ref.String()] = info
	statusCache.Unlock()

	return info, nil
}

func newEmptyStatus(owner string, quorum int) *github.RepoStatus {
	return &github.RepoStatus{
		Context:     github.String(lgtmContext(owner)),
		State:       github.String("pending"),
		Description: github.String(statusInfo{quorum: quorum}.newDescription()),
	}
}
