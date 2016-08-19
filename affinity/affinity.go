// affinity assigns issues based on team mentions and those team captains.
// The idea is to separate triaging of issues and pull requests out
package affinity

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var explanation = `We are utilizing a new workflow in our issues and pull requests. Affinity teams have been setup to allow community members to hear about pull requests that may be interesting to them. When a new issue or pull request comes in, we are asking that the author mention the appropriate affinity team. I then assign a random "team captain" or two to the issue who is in charge of triaging it until it is closed or passing it off to another captain. In order to move forward with this new workflow, we need to know: which of the following teams best fits your issue or contribution?`

var Teams []Team

type Team struct {
	// The team ID.
	ID int

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
	return nil
}

func (t *Team) FetchMetadata(context *ctx.Context) error {
	team, _, err := context.GitHub.Organizations.GetTeam(t.ID)
	if err != nil {
		return err
	}

	t.Description = *team.Description
	return nil
}

func NewTeam(context *ctx.Context, teamId int, name, mention string) (Team, error) {
	team := Team{
		ID:      teamId,
		Name:    name,
		Mention: mention,
	}
	if err := team.FetchCaptains(context); err != nil {
		return Team{}, err
	}

	return team, nil
}

func AssignPRToAffinityTeamCaptain(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("AssignPRToAffinityTeamCaptain: not a pull request event")
	}

	context.SetAuthor(*event.Sender.Login)
	context.SetIssue(*event.Repo.Owner.Login, *event.Repo.Name, *event.Number)

	if *event.Action != "opened" {
		return context.NewError("AssignPRToAffinityTeamCaptain: not an 'opened' PR event")
	}

	if event.PullRequest.Assignee != nil {
		context.IncrStat("affinity.error.already_assigned")
		return context.NewError("AssignPRToAffinityTeamCaptain: PR already assigned")
	}

	context.IncrStat("affinity.pull_request")

	return assignTeamCaptains(context, *event.PullRequest.Body, 2)
}

func AssignIssueToAffinityTeamCaptain(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.IssuesEvent)
	if !ok {
		return context.NewError("AssignIssueToAffinityTeamCaptain: not an issue event")
	}

	context.SetAuthor(*event.Sender.Login)
	context.SetIssue(*event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number)

	if *event.Action != "opened" {
		return context.NewError("AssignIssueToAffinityTeamCaptain: not an 'opened' issue event")
	}

	if event.Assignee != nil {
		context.IncrStat("affinity.error.already_assigned")
		return context.NewError("AssignIssueToAffinityTeamCaptain: issue already assigned")
	}

	context.IncrStat("affinity.issue")

	return assignTeamCaptains(context, *event.Issue.Body, 1)
}
func AssignIssueToAffinityTeamCaptainFromComment(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.IssueCommentEvent)
	if !ok {
		return context.NewError("AssignIssueToAffinityTeamCaptainFromComment: not an issue comment event")
	}

	context.SetAuthor(*event.Sender.Login)
	context.SetIssue(*event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number)

	if *event.Action == "deleted" {
		return context.NewError("AssignIssueToAffinityTeamCaptainFromComment: deleted issue comment event")
	}

	if event.Issue.Assignee != nil {
		return context.NewError("AssignIssueToAffinityTeamCaptainFromComment: issue already assigned")
	}

	context.IncrStat("affinity.issue_comment")

	return assignTeamCaptains(context, *event.Comment.Body, 1)
}

func findAffinityTeam(body string) (Team, error) {
	for _, team := range Teams {
		if strings.Contains(body, team.Mention) {
			return team, nil
		}
	}
	return Team{}, fmt.Errorf("findAffinityTeam: no matching team")
}

func askForAffinityTeam(context *ctx.Context) error {
	_, _, err := context.GitHub.Issues.CreateComment(
		context.Issue.Owner,
		context.Issue.Repo,
		context.Issue.Num,
		&github.IssueComment{Body: github.String(buildAffinityTeamMessage(context))},
	)
	if err != nil {
		return context.NewError("askForAffinityTeam: could not leave comment: %v", err)
	}
	return nil
}

func buildAffinityTeamMessage(context *ctx.Context) string {
	var prefix string
	if context.Issue.Author != "" {
		prefix = fmt.Sprintf("Hey, @%s!", context.Issue.Author)
	} else {
		prefix = "Hello!"
	}

	teams := []string{}
	for _, team := range Teams {
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

func assignTeamCaptains(context *ctx.Context, body string, assigneeCount int) error {
	if context.Issue.IsEmpty() {
		context.IncrStat("affinity.error.no_ref")
		return context.NewError("assignTeamCaptains: issue reference was not set; bailing")
	}

	team, err := findAffinityTeam(body)
	if err != nil {
		context.IncrStat("affinity.error.no_team")
		return askForAffinityTeam(context)
	}

	victims := team.RandomCaptainLogins(assigneeCount)
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
