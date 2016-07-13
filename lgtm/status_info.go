package lgtm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
)

var lgtmerExtractor = regexp.MustCompile("@[a-zA-Z0-9_-]+")

type statusInfo struct {
	lgtmers    []string
	sha        string
	repoStatus *github.RepoStatus
}

func parseStatus(sha string, repoStatus *github.RepoStatus) *statusInfo {
	status := &statusInfo{sha: sha, repoStatus: repoStatus}

	if repoStatus.Description != nil {
		lgtmersExtracted := lgtmerExtractor.FindAllStringSubmatch(*repoStatus.Description, -1)
		if len(lgtmersExtracted) > 0 {
			for _, lgtmerWrapping := range lgtmersExtracted {
				for _, lgtmer := range lgtmerWrapping {
					status.lgtmers = append(status.lgtmers, lgtmer)
				}
			}
		}
	}

	return status
}

func (s statusInfo) IsLGTMer(username string) bool {
	for _, lgtmer := range s.lgtmers {
		if lgtmer == username || lgtmer == "@"+username {
			return true
		}
	}
	return false
}

func newState(lgtmers []string) string {
	if len(lgtmers) >= 2 {
		return "success"
	}
	return "failure"
}

func newDescription(lgtmers []string) string {
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

func statusStateAndDescription(lgtmers []string) (state string, description string) {
	return newState(lgtmers), newDescription(lgtmers)
}

func (s statusInfo) NewStatus() *github.RepoStatus {
	state, description := statusStateAndDescription(s.lgtmers)
	return &github.RepoStatus{
		Context:     github.String(lgtmContext),
		State:       github.String(state),
		Description: github.String(description),
	}
}
