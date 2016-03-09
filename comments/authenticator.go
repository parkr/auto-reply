package comments

import (
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

var (
	teamsCache             = map[string][]github.Team{}
	teamHasPushAccessCache = map[string]*github.Repository{}
	teamMembershipCache    = map[string]bool{}
)

type Authenticator struct {
	client *github.Client
}

func isAuthorizedCommenter(client *github.Client, event github.IssueCommentEvent) bool {
	auth := Authenticator{client}
	orgTeams := auth.teamsForOrg(*event.Repo.Owner.Name)
	for _, team := range orgTeams {
		if auth.isTeamMember(*team.ID, *event.Comment.User.Login) &&
			auth.teamHasPushAccess(*team.ID, *event.Repo.Owner.Name, *event.Repo.Name) {
			return true
		}
	}
	return false
}

func (auth *Authenticator) isTeamMember(teamId int, login string) bool {
	cacheKey := auth.cacheKeyIsTeamMember(teamId, login)
	if _, ok := teamMembershipCache[cacheKey]; !ok {
		newOk, _, err := auth.client.Organizations.IsTeamMember(teamId, login)
		if err != nil {
			log.Printf("ERROR performing IsTeamMember(%d, \"%s\"): %v", teamId, login, err)
			return false
		}
		teamMembershipCache[cacheKey] = newOk
	}
	return teamMembershipCache[cacheKey]
}

func (auth *Authenticator) teamHasPushAccess(teamId int, owner, repo string) bool {
	cacheKey := auth.cacheKeyTeamHashPushAccess(teamId, owner, repo)
	if _, ok := teamHasPushAccessCache[cacheKey]; !ok {
		repository, _, err := auth.client.Organizations.IsTeamRepo(teamId, owner, repo)
		if err != nil {
			log.Printf("ERROR performing IsTeamRepo(%d, \"%s\", \"%s\"): %v", teamId, repo, err)
			return false
		}
		if repository == nil {
			return false
		}
		teamHasPushAccessCache[cacheKey] = repository
	}
	permissions := *teamHasPushAccessCache[cacheKey].Permissions
	return permissions["push"] || permissions["admin"]
}

func (auth *Authenticator) teamsForOrg(org string) []github.Team {
	if _, ok := teamsCache[org]; !ok {
		teamz, _, err := auth.client.Organizations.ListTeams(org, &github.ListOptions{
			PerPage: 100,
		})
		if err != nil {
			log.Printf("ERROR performing ListTeams(\"%s\"): %v", org, err)
			return nil
		}
		teamsCache[org] = teamz
	}
	return teamsCache[org]
}

func (auth *Authenticator) cacheKeyIsTeamMember(teamId int, login string) string {
	return fmt.Sprintf("%d_%s", teamId, login)
}

func (auth *Authenticator) cacheKeyTeamHashPushAccess(teamId int, owner, repo string) string {
	return fmt.Sprintf("%d_%s_%s", teamId, owner, repo)
}
