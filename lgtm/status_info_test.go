package lgtm

import (
	"fmt"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestParseStatus(t *testing.T) {
	cases := []struct {
		sha             string
		description     string
		expectedLgtmers []string
		expectedQuorum  int
	}{
		{"deadbeef", "", []string{}, 0},
		{"deadbeef", "Waiting for approval from at least 2 maintainers.", []string{}, 2},
		{"deadbeef", "Waiting for approval from at least 22 maintainers.", []string{}, 22},
		{"deadbeef", "Approved by @parkr. Requires 1 more LGTM.", []string{"@parkr"}, 2},
		{"deadbeef", "@parkr have approved this PR. Requires 32 more LGTM's.", []string{"@parkr"}, 33},
		{"deadbeef", "@parkr, and @envygeeks have approved this PR.", []string{"@parkr", "@envygeeks"}, 2},
		{"deadbeef", "@mattr-, @parkr, and @BenBalter have approved this PR. Requires no more LGTM's.", []string{"@mattr-", "@parkr", "@BenBalter"}, 3},
	}
	for _, test := range cases {
		parsed := parseStatus(test.sha, &github.RepoStatus{Description: github.String(test.description)})
		assert.Equal(t,
			test.expectedLgtmers, parsed.lgtmers,
			fmt.Sprintf("parsing description: %q", test.description))
		assert.Equal(t,
			test.expectedQuorum, parsed.quorum,
			fmt.Sprintf("parsing description: %q", test.description))
		assert.Equal(t, test.sha, parsed.sha)
	}
}

func TestStatusInfoIsLGTMer(t *testing.T) {
	cases := []struct {
		info             statusInfo
		lgtmerInQuestion string
		islgtmer         bool
	}{
		{statusInfo{}, "@parkr", false},
		{statusInfo{lgtmers: []string{"@parkr"}}, "@parkr", true},
		{statusInfo{lgtmers: []string{"@parkr"}}, "@mattr-", false},
		{statusInfo{lgtmers: []string{"@parkr", "@mattr-"}}, "@mattr-", true},
		{statusInfo{lgtmers: []string{"@parkr", "@mattr-"}}, "@parkr-", false},
		{statusInfo{lgtmers: []string{"@parkr", "@mattr-"}}, "@parkr", true},
		{statusInfo{lgtmers: []string{"@parkr", "@mattr-"}}, "@PARKR", true},
		{statusInfo{lgtmers: []string{"@benBalter", "@mattr-"}}, "@benbalter", true},
	}
	for _, test := range cases {
		assert.Equal(t,
			test.islgtmer, test.info.IsLGTMer(test.lgtmerInQuestion),
			fmt.Sprintf("asking about: %q for lgtmers: %q", test.lgtmerInQuestion, test.info.lgtmers))
	}
}

func TestNewState(t *testing.T) {
	cases := []struct {
		lgtmers  []string
		quorum   int
		expected string
	}{
		{[]string{}, 0, "success"},
		{[]string{}, 1, "failure"},
		{[]string{}, 2, "failure"},
		{[]string{"@parkr"}, 0, "success"},
		{[]string{"@parkr"}, 1, "success"},
		{[]string{"@parkr"}, 2, "failure"},
		{[]string{"@parkr", "@mattr-"}, 0, "success"},
		{[]string{"@parkr", "@mattr-"}, 1, "success"},
		{[]string{"@parkr", "@mattr-"}, 2, "success"},
	}
	for _, test := range cases {
		info := statusInfo{lgtmers: test.lgtmers, quorum: test.quorum}
		assert.Equal(t,
			test.expected, info.newState(),
			fmt.Sprintf("with lgtmers: %q and quorum: %d", test.lgtmers, test.quorum))
	}
}

func TestNewDescription(t *testing.T) {
	cases := []struct {
		lgtmers     []string
		quorum      int
		description string
	}{
		{nil, 0, "No approval is required."},
		{nil, 1, "Awaiting approval from at least 1 maintainer."},
		{[]string{}, 2, "Awaiting approval from at least 2 maintainers."},
		{[]string{"@parkr"}, 2, "Approved by @parkr. Requires 1 more LGTM."},
		{[]string{"@parkr", "@envygeeks"}, 2, "Approved by @parkr and @envygeeks."},
		{[]string{"@mattr-", "@envygeeks", "@parkr"}, 5, "Approved by @mattr-, @envygeeks, and @parkr. Requires 2 more LGTM's."},
	}
	for _, test := range cases {
		info := statusInfo{lgtmers: test.lgtmers, quorum: test.quorum}
		actual := info.newDescription()
		assert.Equal(t, test.description, actual)
		assert.True(t, len(actual) <= 140, fmt.Sprintf("%q must be <= 140 chars.", actual))
	}
}

func TestLGTMsRequiredDescription(t *testing.T) {
	cases := []struct {
		lgtmers  []string
		quorum   int
		expected string
	}{
		{nil, 0, ""},
		{nil, 1, "Requires 1 more LGTM."},
		{[]string{}, 2, "Requires 2 more LGTM's."},
		{[]string{"@parkr"}, 2, "Requires 1 more LGTM."},
		{[]string{"@parkr", "@envygeeks"}, 2, ""},
		{[]string{"@mattr-", "@envygeeks", "@parkr"}, 5, "Requires 2 more LGTM's."},
	}
	for _, test := range cases {
		info := statusInfo{lgtmers: test.lgtmers, quorum: test.quorum}
		actual := info.newLGTMsRequiredDescription()
		assert.Equal(t, test.expected, actual)
		assert.True(t, len(actual) <= 140, fmt.Sprintf("%q must be <= 140 chars.", actual))
	}
}

func TestNewApprovedByDescription(t *testing.T) {
}

func TestStatusInfoNewRepoStatus(t *testing.T) {
	cases := []struct {
		owner          string
		lgtmers        []string
		quorum         int
		expContext     string
		expState       string
		expDescription string
	}{
		{"octocat", []string{}, 0, "octocat/lgtm", "success", "No approval is required."},
		{"parkr", []string{}, 0, "parkr/lgtm", "success", "No approval is required."},
		{"jekyll", []string{}, 1, "jekyll/lgtm", "failure", "Awaiting approval from at least 1 maintainer."},
		{"jekyll", []string{"@parkr"}, 1, "jekyll/lgtm", "success", "Approved by @parkr."},
		{"jekyll", []string{"@parkr"}, 2, "jekyll/lgtm", "failure", "Approved by @parkr. Requires 1 more LGTM."},
		{"jekyll", []string{"@parkr", "@envygeeks"}, 1, "jekyll/lgtm", "success", "Approved by @parkr and @envygeeks."},
		{"jekyll", []string{"@parkr", "@envygeeks"}, 2, "jekyll/lgtm", "success", "Approved by @parkr and @envygeeks."},
		{"jekyll", []string{"@parkr", "@mattr-", "@envygeeks"}, 6, "jekyll/lgtm", "failure", "Approved by @parkr, @mattr-, and @envygeeks. Requires 3 more LGTM's."},
	}
	for _, test := range cases {
		status := statusInfo{lgtmers: test.lgtmers, quorum: test.quorum}
		newStatus := status.NewRepoStatus(test.owner)
		assert.Equal(t,
			test.expContext, *newStatus.Context,
			fmt.Sprintf("with lgtmers: %q and quorum: %d", test.lgtmers, test.quorum))
		assert.Equal(t,
			test.expState, *newStatus.State,
			fmt.Sprintf("with lgtmers: %q and quorum: %d", test.lgtmers, test.quorum))
		assert.Equal(t,
			test.expDescription, *newStatus.Description,
			fmt.Sprintf("with lgtmers: %q and quorum: %d", test.lgtmers, test.quorum))
		assert.True(t, len(*newStatus.Description) <= 140, fmt.Sprintf("%q must be <= 140 chars.", *newStatus.Description))
	}
}
