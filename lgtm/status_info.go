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
	quorum     int
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

func (s statusInfo) newState() string {
	if len(s.lgtmers) >= s.quorum {
		return "success"
	}
	return "failure"
}

func (s statusInfo) newDescription() string {
	switch len(s.lgtmers) {
	case 0:
		return descriptionNoLGTMers
	case 1:
		return fmt.Sprintf("%s has approved this PR. %d LGTM's are required.", s.lgtmers[0], s.quorum)
	case 2:
		return fmt.Sprintf("%s and %s have approved this PR. %d LGTM's are required.", s.lgtmers[0], s.lgtmers[1], s.quorum)
	default:
		lastIndex := len(s.lgtmers) - 1
		return fmt.Sprintf("%s, and %s have approved this PR. %d LGTM's are required.",
			strings.Join(s.lgtmers[0:lastIndex], ", "), s.lgtmers[lastIndex], s.quorum)
	}
}

func (s statusInfo) NewRepoStatus(owner string) *github.RepoStatus {
	return &github.RepoStatus{
		Context:     github.String(lgtmContext(owner)),
		State:       github.String(s.newState()),
		Description: github.String(s.newDescription()),
	}
}
