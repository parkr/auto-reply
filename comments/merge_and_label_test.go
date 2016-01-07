package comments

import (
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
		{"@jekyllbot: merge +minor-enhancement", true, "minor-enhancement"},
		{"@jekyllbot: merge +Bug Fix\n", true, "bug-fix"},
	}
	for _, c := range comments {
		isReq, label := parseMergeRequestComment(c.comment)
		assert.Equal(t, c.isReq, isReq, "'%s' should have isReq=%v", c.comment, c.isReq)
		assert.Equal(t, c.label, label, "'%s' should have label=%v", c.comment, c.label)
	}
}
