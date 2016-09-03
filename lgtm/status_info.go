package lgtm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

var lgtmerExtractor = regexp.MustCompile("@[a-zA-Z0-9_-]+")
var remainingLGTMsExtractor = regexp.MustCompile(`Waiting for approval from at least (\d+)|Requires (\d+) more LGTM('s)?`)

type statusInfo struct {
	lgtmers    []string
	quorum     int
	sha        string
	repoStatus *github.RepoStatus
}

func parseStatus(sha string, repoStatus *github.RepoStatus) *statusInfo {
	status := &statusInfo{sha: sha, repoStatus: repoStatus, lgtmers: []string{}}

	if repoStatus.Description != nil {
		// Extract LGTMers.
		lgtmersExtracted := lgtmerExtractor.FindAllStringSubmatch(*repoStatus.Description, -1)
		if len(lgtmersExtracted) > 0 {
			for _, lgtmerWrapping := range lgtmersExtracted {
				for _, lgtmer := range lgtmerWrapping {
					status.lgtmers = append(status.lgtmers, lgtmer)
				}
			}
		}

		status.quorum = len(status.lgtmers)

		// Extract additional quorum. :)
		extractedRemainingLGTMs := remainingLGTMsExtractor.FindAllStringSubmatch(*repoStatus.Description, -1)
		if len(extractedRemainingLGTMs) > 0 && len(extractedRemainingLGTMs[0]) > 2 {
			remainingLGTMsString := extractedRemainingLGTMs[0][1]
			if remainingLGTMsString == "" {
				remainingLGTMsString = extractedRemainingLGTMs[0][2]
			}

			remainingLGTMs, err := strconv.Atoi(remainingLGTMsString)
			if err != nil {
				fmt.Printf("lgtm.parseStatus: error parsing %q for remaining LGTM's: %v", *repoStatus.Description, err)
			}
			status.quorum += remainingLGTMs
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
	return "pending"
}

// newDescription produces the LGTM status description based on the LGTMers
// and quorum values specified for this statusInfo.
func (s statusInfo) newDescription() string {
	if s.quorum == 0 {
		return "No approval is required."
	}

	if len(s.lgtmers) == 0 {
		message := fmt.Sprintf("Awaiting approval from at least %d maintainer", s.quorum)
		if s.quorum > 1 {
			message += "s"
		}
		return message + "."
	}

	if requiredLGTMsDesc := s.newLGTMsRequiredDescription(); requiredLGTMsDesc != "" {
		return s.newApprovedByDescription() + " " + requiredLGTMsDesc
	} else {
		return s.newApprovedByDescription()
	}
}

func (s statusInfo) newLGTMsRequiredDescription() string {
	remaining := s.quorum - len(s.lgtmers)

	switch {
	case remaining <= 0:
		return ""
	case remaining == 1:
		return "Requires 1 more LGTM."
	default:
		return fmt.Sprintf("Requires %d more LGTM's.", remaining)
	}
}

func (s statusInfo) newApprovedByDescription() string {
	switch len(s.lgtmers) {
	case 0:
		return "Not yet approved by any maintainers."
	case 1:
		return fmt.Sprintf("Approved by %s.", s.lgtmers[0])
	case 2:
		return fmt.Sprintf("Approved by %s and %s.", s.lgtmers[0], s.lgtmers[1])
	default:
		lastIndex := len(s.lgtmers) - 1
		return fmt.Sprintf("Approved by %s, and %s.",
			strings.Join(s.lgtmers[0:lastIndex], ", "), s.lgtmers[lastIndex])
	}
}

func (s statusInfo) NewRepoStatus(owner string) *github.RepoStatus {
	return &github.RepoStatus{
		Context:     github.String(lgtmContext(owner)),
		State:       github.String(s.newState()),
		Description: github.String(s.newDescription()),
	}
}
