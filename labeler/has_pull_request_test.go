package labeler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
