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
	}{
		{"deadbeef", "", []string{}},
		{"deadbeef", "This pull request has not received any LGTM's.", []string{}},
		{"deadbeef", "@parkr has approved this PR.", []string{"@parkr"}},
		{"deadbeef", "@parkr have approved this PR.", []string{"@parkr"}},
		{"deadbeef", "@parkr, and @envygeeks have approved this PR.", []string{"@parkr", "@envygeeks"}},
		{"deadbeef", "@mattr-, @parkr, and @BenBalter have approved this PR.", []string{"@mattr-", "@parkr", "@BenBalter"}},
	}
	for _, test := range cases {
		parsed := parseStatus(test.sha, &github.RepoStatus{Description: github.String(test.description)})
		assert.Equal(t,
			test.expectedLgtmers, parsed.lgtmers,
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
		assert.Equal(t,
			test.expected, newState(test.lgtmers, test.quorum),
			fmt.Sprintf("with lgtmers: %q and quorum: %d", test.lgtmers, test.quorum))
	}
}

func TestNewDescription(t *testing.T) {
	cases := []struct {
		lgtmers     []string
		description string
	}{
		{nil, descriptionNoLGTMers},
		{[]string{}, descriptionNoLGTMers},
		{[]string{"@parkr"}, "@parkr has approved this PR."},
		{[]string{"@parkr", "@envygeeks"}, "@parkr and @envygeeks have approved this PR."},
		{[]string{"@mattr-", "@envygeeks", "@parkr"}, "@mattr-, @envygeeks, and @parkr have approved this PR."},
	}
	for _, testCase := range cases {
		assert.Equal(t, testCase.description, newDescription(testCase.lgtmers))
	}
}

func TestStatusInfoNewStatus(t *testing.T) {
	cases := []struct {
		owner          string
		lgtmers        []string
		quorum         int
		expContext     string
		expState       string
		expDescription string
	}{
		{"octocat", []string{}, 0, "octocat/lgtm", "success", descriptionNoLGTMers},
		{"parkr", []string{}, 0, "parkr/lgtm", "success", descriptionNoLGTMers},
		{"jekyll", []string{}, 1, "jekyll/lgtm", "failure", descriptionNoLGTMers},
		{"jekyll", []string{"@parkr"}, 1, "jekyll/lgtm", "success", "@parkr has approved this PR."},
		{"jekyll", []string{"@parkr"}, 2, "jekyll/lgtm", "failure", "@parkr has approved this PR."},
		{"jekyll", []string{"@parkr", "@envygeeks"}, 1, "jekyll/lgtm", "success", "@parkr and @envygeeks have approved this PR."},
		{"jekyll", []string{"@parkr", "@envygeeks"}, 2, "jekyll/lgtm", "success", "@parkr and @envygeeks have approved this PR."},
	}
	for _, test := range cases {
		status := statusInfo{lgtmers: test.lgtmers}
		newStatus := status.NewStatus(test.owner, test.quorum)
		assert.Equal(t,
			test.expContext, *newStatus.Context,
			fmt.Sprintf("with lgtmers: %q and quorum: %d", test.lgtmers, test.quorum))
		assert.Equal(t,
			test.expState, *newStatus.State,
			fmt.Sprintf("with lgtmers: %q and quorum: %d", test.lgtmers, test.quorum))
		assert.Equal(t,
			test.expDescription, *newStatus.Description,
			fmt.Sprintf("with lgtmers: %q and quorum: %d", test.lgtmers, test.quorum))
	}
}
