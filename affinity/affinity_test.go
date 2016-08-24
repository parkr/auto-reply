package affinity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleLongComment = `On the site documentation section, links to documentation sections always point to the jekyllrb.com website, this means that users testing changes might get confused because they will see the official external website page instead of their local website upon clicking those links.


**Please check if this change doesn't break the official website on https://jekyllrb.com before accepting the pull request.**

----------

@jekyll/documentation`

func TestFindAffinityTeam(t *testing.T) {
	allTeams := []Team{
		{ID: 456, Mention: "@jekyll/documentation"},
		{ID: 789, Mention: "@jekyll/ecosystem"},
		{ID: 101, Mention: "@jekyll/performance"},
		{ID: 213, Mention: "@jekyll/stability"},
		{ID: 141, Mention: "@jekyll/windows"},
		{ID: 123, Mention: "@jekyll/build"},
	}

	examples := []struct {
		body           string
		matchingTeamID int
	}{
		{exampleLongComment, 456},
		{"@jekyll/documentation @jekyll/build", 456},
		{"@jekyll/windows @jekyll/documentation", 456},
		{"@jekyll/windows", 141},
	}
	for _, example := range examples {
		matchingTeam, err := findAffinityTeam(example.body, allTeams)
		assert.NoError(t, err)
		assert.Equal(t, matchingTeam.ID, example.matchingTeamID,
			"expected the following to match %d team: `%s`", example.matchingTeamID, example.body)
	}
}
