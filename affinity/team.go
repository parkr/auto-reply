package affinity

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/auth"
	"github.com/parkr/auto-reply/ctx"
)

func NewTeam(context *ctx.Context, teamId int) (Team, error) {
	team := Team{ID: teamId}
	if err := team.FetchMetadata(context); err != nil {
		return Team{}, err
	}
	if err := team.FetchCaptains(context); err != nil {
		return Team{}, err
	}

	return team, nil
}

type Team struct {
	// The team ID.
	ID int

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

func (t Team) RandomCaptainLogins(num int) []string {
	rand.Seed(time.Now().UnixNano())

	selections := []string{}

	// Just return all of them.
	if len(t.Captains) <= num {
		for _, captain := range t.Captains {
			selections = append(selections, *captain.Login)
		}
		return selections
	}

	// Find a random selection.
OuterLoop:
	for {
		selection := t.Captains[rand.Intn(len(t.Captains))]
		for _, previous := range selections {
			if *selection.Login == previous {
				continue OuterLoop
			}
		}
		selections = append(selections, *selection.Login)

		if len(selections) == num {
			break
		}
	}
	return selections
}

func (t *Team) FetchCaptains(context *ctx.Context) error {
	users, _, err := context.GitHub.Organizations.ListTeamMembers(t.ID, &github.OrganizationListTeamMembersOptions{
		Role:        "maintainer",
		ListOptions: github.ListOptions{Page: 0, PerPage: 100},
	})
	if err != nil {
		return err
	}

	t.Captains = users

	if t.Org != "" {
		allMembers, _, err := context.GitHub.Organizations.ListTeamMembers(t.ID, &github.OrganizationListTeamMembersOptions{
			Role:        "all",
			ListOptions: github.ListOptions{Page: 0, PerPage: 100},
		})
		if err != nil {
			return err
		}

		for _, user := range allMembers {
			if auth.UserIsOrgOwner(context, t.Org, *user.Login) {
				t.Captains = append(t.Captains, user)
			}
		}
	} else {
		context.Log("Team.FetchCaptains: cannot fetch org owners without Team.Org value")
	}

	return nil
}

func (t *Team) FetchMetadata(context *ctx.Context) error {
	team, _, err := context.GitHub.Organizations.GetTeam(t.ID)
	if err != nil {
		return err
	}

	t.Org = *team.Organization.Login
	t.Name = *team.Name
	t.Mention = fmt.Sprintf("@%s/%s", t.Org, *team.Slug)
	t.Description = *team.Description
	return nil
}

func (t *Team) IsTeamCaptain(login string) bool {
	for _, captain := range t.Captains {
		if *captain.Login == login {
			return true
		}
	}
	return false
}
