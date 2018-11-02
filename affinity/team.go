package affinity

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

func NewTeam(context *ctx.Context, teamId int64) (Team, error) {
	team := Team{ID: teamId}
	if err := team.fetchMetadata(context); err != nil {
		return Team{}, err
	}
	if err := team.fetchCaptains(context); err != nil {
		return Team{}, err
	}

	return team, nil
}

type Team struct {
	// The team ID.
	ID int64

	// The org the team belongs to
	Org string

	// The name of the team.
	Name string

	// The mention this should match, e.g. "@jekyll/documentation"
	Mention string

	// The description of the repo.
	Description string

	// Team captains, requires at least the Login field
	Captains []*github.User
}

func (t Team) String() string {
	return fmt.Sprintf("Team{ID=%d Org=%s Name=%s Mention=%s Description=%s Captains=%q",
		t.ID,
		t.Org,
		t.Name,
		t.Mention,
		t.Description,
		usersByLogin(t.Captains),
	)
}

func (t Team) RandomCaptainLogins(num int) []string {
	rand.Seed(time.Now().UnixNano())

	selectionmap := map[string]bool{}

	// Just return all of them.
	if len(t.Captains) <= num {
		return usersByLogin(t.Captains)
	}

	// Find a random selection.
	for {
		selection := t.Captains[rand.Intn(len(t.Captains))]
		selectionmap[selection.GetLogin()] = true

		if len(selectionmap) == num {
			break
		}
	}

	selections := []string{}
	for login := range selectionmap {
		selections = append(selections, login)
	}
	return selections
}

func (t Team) RandomCaptainLoginsExcluding(excludedLogin string, count int) []string {
	var selections []string

	// If the pool of captains isn't big enough to hit count without the excluded login,
	// then just return all the other captains.
	if len(t.Captains)-1 <= count {
		for _, user := range t.Captains {
			if user.GetLogin() != excludedLogin {
				selections = append(selections, user.GetLogin())
			}
		}
		return selections
	}

	// We can only ever exclude 1 login, so just get count + 1 random logins
	// and pick them one by one until we get our desired count.
	for _, login := range t.RandomCaptainLogins(count + 1) {
		if login != excludedLogin {
			selections = append(selections, login)
		}
		if len(selections) == count {
			break
		}
	}

	return selections
}

func (t *Team) fetchCaptains(context *ctx.Context) error {
	users, _, err := context.GitHub.Teams.ListTeamMembers(
		context.Context(),
		t.ID,
		&github.TeamListTeamMembersOptions{
			Role:        "maintainer",
			ListOptions: github.ListOptions{Page: 0, PerPage: 100},
		},
	)
	if err != nil {
		return err
	}

	for _, user := range users {
		if !context.GitHubAuthedAs(user.GetLogin()) {
			t.Captains = append(t.Captains, user)
		}
	}

	return nil
}

func (t *Team) fetchMetadata(context *ctx.Context) error {
	team, _, err := context.GitHub.Teams.GetTeam(context.Context(), t.ID)
	if err != nil {
		return err
	}

	t.Org = *team.Organization.Login
	t.Name = *team.Name
	t.Mention = fmt.Sprintf("@%s/%s", t.Org, *team.Slug)
	t.Description = *team.Description
	return nil
}
