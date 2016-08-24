// affinity assigns issues based on team mentions and those team captains.
// The idea is to separate the work of triaging of issues and pull requests
// out to a larger pool of people to make it less of a burden to be involved.
package affinity

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var explanation = `We are utilizing a new workflow in our issues and pull requests. Affinity teams have been setup to allow community members to hear about pull requests that may be interesting to them. When a new issue or pull request comes in, we are asking that the author mention the appropriate affinity team. I then assign a random "team captain" or two to the issue who is in charge of triaging it until it is closed or passing it off to another captain. In order to move forward with this new workflow, we need to know: which of the following teams best fits your issue or contribution?`

func assignTeamCaptains(context *ctx.Context, handler Handler, body string, assigneeCount int) error {
	if context.Issue.IsEmpty() {
		context.IncrStat("affinity.error.no_ref")
		return context.NewError("assignTeamCaptains: issue reference was not set; bailing")
	}

	team, err := findAffinityTeam(body, handler.teams)
	if err != nil {
		context.IncrStat("affinity.error.no_team")
		return askForAffinityTeam(context, handler.teams)
	}

	context.Log("all affinity team captains: %q", team.Captains)
	victims := team.RandomCaptainLogins(assigneeCount)
	context.Log("selected affinity team captains for %s: %q", context.Issue, victims)
	_, _, err = context.GitHub.Issues.AddAssignees(
		context.Issue.Owner,
		context.Issue.Repo,
		context.Issue.Num,
		victims,
	)
	if err != nil {
		context.IncrStat("affinity.error.github_api")
		return context.NewError("assignTeamCaptains: problem assigning: %v", err)
	}

	context.IncrStat("affinity.success")
	context.Log("assignTeamCaptains: assigned %q to %s", victims, context.Issue)
	return nil
}

func findAffinityTeam(body string, allTeams []Team) (Team, error) {
	for _, team := range allTeams {
		if strings.Contains(body, team.Mention) {
			return team, nil
		}
	}
	return Team{}, fmt.Errorf("findAffinityTeam: no matching team")
}

func askForAffinityTeam(context *ctx.Context, allTeams []Team) error {
	_, _, err := context.GitHub.Issues.CreateComment(
		context.Issue.Owner,
		context.Issue.Repo,
		context.Issue.Num,
		&github.IssueComment{Body: github.String(buildAffinityTeamMessage(context, allTeams))},
	)
	if err != nil {
		return context.NewError("askForAffinityTeam: could not leave comment: %v", err)
	}
	return nil
}

func buildAffinityTeamMessage(context *ctx.Context, allTeams []Team) string {
	var prefix string
	if context.Issue.Author != "" {
		prefix = fmt.Sprintf("Hey, @%s!", context.Issue.Author)
	} else {
		prefix = "Hello!"
	}

	teams := []string{}
	for _, team := range allTeams {
		teams = append(teams, fmt.Sprintf(
			"- `%s` – %s",
			team.Mention, team.Description,
		))
	}

	return fmt.Sprintf(
		"%s %s\n\n%s\n\nMention one of these teams in a comment below and we'll get this sorted. Thanks!",
		prefix, explanation, strings.Join(teams, "\n"),
	)
}
