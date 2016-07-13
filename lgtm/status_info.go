package lgtm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
)

const descriptionNoLGTMers = "This pull request has received no LGTM's."

var lgtmerExtractor = regexp.MustCompile("@[a-zA-Z0-9_-]+")

type statusInfo struct {
	lgtmers    []string
	sha        string
	repoStatus *github.RepoStatus
}

func parseStatus(sha string, repoStatus *github.RepoStatus) *statusInfo {
	status := &statusInfo{sha: sha, repoStatus: repoStatus, lgtmers: []string{}}

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
	lowerUsername := strings.ToLower(username)
	for _, lgtmer := range s.lgtmers {
		lowerLgtmer := strings.ToLower(lgtmer)
		if lowerLgtmer == lowerUsername || lowerLgtmer == "@"+lowerUsername {
			return true
		}
	}
	return false
}

func newState(lgtmers []string, quorum int) string {
	if len(lgtmers) >= quorum {
		return "success"
	}
	return "failure"
}

func newDescription(lgtmers []string) string {
	switch len(lgtmers) {
	case 0:
		return descriptionNoLGTMers
	case 1:
		return fmt.Sprintf("%s has approved this PR.", lgtmers[0])
	case 2:
		return fmt.Sprintf("%s and %s have approved this PR.", lgtmers[0], lgtmers[1])
	default:
		lastIndex := len(lgtmers) - 1
		return fmt.Sprintf("%s, and %s have approved this PR.",
			strings.Join(lgtmers[0:lastIndex], ", "), lgtmers[lastIndex])
	}
}

func statusStateAndDescription(lgtmers []string, quorum int) (state string, description string) {
	return newState(lgtmers, quorum), newDescription(lgtmers)
}

func (s statusInfo) NewStatus(owner string, quorum int) *github.RepoStatus {
	state, description := statusStateAndDescription(s.lgtmers, quorum)
	return &github.RepoStatus{
		Context:     github.String(lgtmContext(owner)),
		State:       github.String(state),
		Description: github.String(description),
	}
}
