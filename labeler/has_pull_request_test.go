package labeler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkedIssues(t *testing.T) {
	assert.Equal(t, []int{13, 14},
		linkedIssues("Fixes #13. Fixes #14"))

	assert.Equal(t, []int{13, 14, 1, 412, 2},
		linkedIssues("Fixes #13. Fixes # Resolves #14 Settles #12 Closes #1. Fixes #412..... Close #2"))

	multilineComment := `Upgrade Rubocop to 0.49.0

Fix #6089
Fix #6101 `
	assert.Equal(t, []int{6089, 6101}, linkedIssues(multilineComment))
}

func TestClosedIssueRegex(t *testing.T) {
	assert.Equal(t, "13",
		fixesIssueMatcher.FindAllStringSubmatch("Fixes #13", -1)[0][1],
		"should have extracted 13 from 'Fixes #13'")

	assert.Equal(t, "13",
		fixesIssueMatcher.FindAllStringSubmatch("Closed #13", -1)[0][1],
		"should match 'close' pattern too 'Closed #13'")

	assert.Equal(t, "13",
		fixesIssueMatcher.FindAllStringSubmatch("rEsoLvEd #13", -1)[0][1],
		"should match mixedcase pattern")
}
