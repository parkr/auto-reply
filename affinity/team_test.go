package affinity

import (
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestTeamRandomCaptainLogins(t *testing.T) {
	team := Team{Captains: []*github.User{
		{Login: github.String("parkr")},
		{Login: github.String("envygeeks")},
		{Login: github.String("mattr-")},
	}}
	selections := team.RandomCaptainLogins(1)
	assert.Len(t, selections, 1)
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[0])

	selections = team.RandomCaptainLogins(2)
	assert.Len(t, selections, 2)
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[0])
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[1])

	selections = team.RandomCaptainLogins(3)
	assert.Len(t, selections, 3)
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[0])
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[1])
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[2])

	selections = team.RandomCaptainLogins(4)
	assert.Len(t, selections, 3)
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[0])
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[1])
	assert.Contains(t, []string{"parkr", "envygeeks", "mattr-"}, selections[2])
}
