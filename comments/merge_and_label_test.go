package comments

import (
	"io/ioutil"
	"testing"

	"github.com/parkr/auto-reply/Godeps/_workspace/src/github.com/stretchr/testify/assert"
)

func TestParseMergeRequestComment(t *testing.T) {
	comments := []struct {
		comment string
		isReq   bool
		label   string
	}{
		{"it looked like you could merge it", false, ""},
		{"@jekyllbot: merge", true, ""},
		{"@jekyllbot: :shipit:", true, ""},
		{"@jekyllbot: :ship:", true, ""},
		{"@jekyllbot: merge +minor-enhancement", true, "minor-enhancements"},
		{"@jekyllbot: merge +Bug Fix\n", true, "bug-fixes"},
	}
	for _, c := range comments {
		isReq, label := parseMergeRequestComment(c.comment)
		assert.Equal(t, c.isReq, isReq, "'%s' should have isReq=%v", c.comment, c.isReq)
		assert.Equal(t, c.label, label, "'%s' should have label=%v", c.comment, c.label)
	}
}

func TestBase64Decode(t *testing.T) {
	encoded, err := ioutil.ReadFile("history_contents.enc")
	assert.NoError(t, err)
	decoded := base64Decode(string(encoded))
	assert.Contains(t, decoded, "### Minor Enhancements")
}

func TestAddMergeReference(t *testing.T) {
	historyFile := addMergeReference("", "Development Fixes", "Some great change", 1)
	assert.Equal(t, "## HEAD\n\n### Development Fixes\n\n  * Some great change (#1)\n", historyFile)

	historyFile = addMergeReference(
		"## HEAD",
		"Development Fixes", "Another great change!!!!!!!", 1)
	assert.Equal(t, "## HEAD\n\n### Development Fixes\n\n  * Another great change!!!!!!! (#1)\n", historyFile)

	historyFile = addMergeReference(
		"## HEAD\n\n### Development Fixes\n\n  * Some great change (#1)\n",
		"Development Fixes", "Another great change!!!!!!!", 1)
	assert.Equal(t, "## HEAD\n\n### Development Fixes\n\n  * Some great change (#1)\n  * Another great change!!!!!!! (#1)\n", historyFile)
}
