package affinity

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

type Handler struct {
	repos []Repo
	teams []Team
}

func (h *Handler) enabledForRepo(owner, name string) bool {
	for _, repo := range h.repos {
		if repo.Owner == owner && repo.Name == name {
			return true
		}
	}

	return false
}

func (h *Handler) GetRepos() []Repo {
	return h.repos
}

func (h *Handler) AddRepo(owner, name string) {
	if h.repos == nil {
		h.repos = []Repo{}
	}

	if h.enabledForRepo(owner, name) {
		return
	}

	h.repos = append(h.repos, Repo{Owner: owner, Name: name})
}

func (h *Handler) GetTeams() []Team {
	return h.teams
}

func (h *Handler) AddTeam(context *ctx.Context, teamID int64) error {
	if h.teams == nil {
		h.teams = []Team{}
	}

	if _, err := h.GetTeam(teamID); err == nil {
		return nil // already have it!
	}

	team, err := NewTeam(context, teamID)
	if err != nil {
		return err
	}

	h.teams = append(h.teams, team)
	return nil
}

func (h *Handler) GetTeam(teamID int64) (Team, error) {
	for _, team := range h.teams {
		if team.ID == teamID {
			return team, nil
		}
	}
	return Team{}, fmt.Errorf("GetTeam: team with ID=%d not found", teamID)
}

func (h *Handler) RequestReviewFromAffinityTeamCaptains(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("RequestReviewFromAffinityTeamCaptains: not a pull request event")
	}

	context.SetAuthor(*event.Sender.Login)
	context.SetIssue(*event.Repo.Owner.Login, *event.Repo.Name, *event.Number)

	if !h.enabledForRepo(context.Issue.Owner, context.Issue.Repo) {
		return context.NewError("RequestReviewFromAffinityTeamCaptains: not enabled for %s", context.Issue)
	}

	if *event.Action != "opened" {
		return context.NewError("RequestReviewFromAffinityTeamCaptains: not an 'opened' PR event")
	}

	context.IncrStat("affinity.pull_request", []string{"task:request_review"})

	return requestReviewFromTeamCaptains(context, *h, *event.PullRequest.Body, 2)
}

func (h *Handler) AssignPRToAffinityTeamCaptain(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("AssignPRToAffinityTeamCaptain: not a pull request event")
	}

	context.SetAuthor(*event.Sender.Login)
	context.SetIssue(*event.Repo.Owner.Login, *event.Repo.Name, *event.Number)

	if !h.enabledForRepo(context.Issue.Owner, context.Issue.Repo) {
		return context.NewError("AssignPRToAffinityTeamCaptain: not enabled for %s", context.Issue)
	}

	if *event.Action != "opened" {
		return context.NewError("AssignPRToAffinityTeamCaptain: not an 'opened' PR event")
	}

	if event.PullRequest.Assignee != nil {
		context.IncrStat("affinity.error.already_assigned", nil)
		return context.NewError("AssignPRToAffinityTeamCaptain: PR already assigned")
	}

	if context.GitHubAuthedAs(*event.Sender.Login) {
		return fmt.Errorf("bozo. you can't reply to your own comment!")
	}

	context.IncrStat("affinity.pull_request", nil)

	return assignTeamCaptains(context, *h, *event.PullRequest.Body, 1)
}

func (h *Handler) AssignIssueToAffinityTeamCaptain(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.IssuesEvent)
	if !ok {
		return context.NewError("AssignIssueToAffinityTeamCaptain: not an issue event")
	}

	context.SetAuthor(*event.Sender.Login)
	context.SetIssue(*event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number)

	if !h.enabledForRepo(context.Issue.Owner, context.Issue.Repo) {
		return context.NewError("AssignIssueToAffinityTeamCaptain: not enabled for %s", context.Issue)
	}

	if *event.Action != "opened" {
		return context.NewError("AssignIssueToAffinityTeamCaptain: not an 'opened' issue event")
	}

	if event.Assignee != nil {
		context.IncrStat("affinity.error.already_assigned", nil)
		return context.NewError("AssignIssueToAffinityTeamCaptain: issue already assigned")
	}

	if context.GitHubAuthedAs(*event.Sender.Login) {
		return fmt.Errorf("bozo. you can't reply to your own comment!")
	}

	context.IncrStat("affinity.issue", nil)

	return assignTeamCaptains(context, *h, *event.Issue.Body, 1)
}

func (h *Handler) AssignIssueToAffinityTeamCaptainFromComment(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.IssueCommentEvent)
	if !ok {
		return context.NewError("AssignIssueToAffinityTeamCaptainFromComment: not an issue comment event")
	}

	context.SetAuthor(*event.Sender.Login)
	context.SetIssue(*event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number)

	if !h.enabledForRepo(context.Issue.Owner, context.Issue.Repo) {
		return context.NewError("AssignIssueToAffinityTeamCaptainFromComment: not enabled for %s", context.Issue)
	}

	if *event.Action == "deleted" {
		return context.NewError("AssignIssueToAffinityTeamCaptainFromComment: deleted issue comment event")
	}

	if event.Issue.Assignee != nil {
		return context.NewError("AssignIssueToAffinityTeamCaptainFromComment: issue already assigned")
	}

	if context.GitHubAuthedAs(*event.Sender.Login) {
		return fmt.Errorf("bozo. you can't reply to your own comment!")
	}

	context.IncrStat("affinity.issue_comment", nil)

	return assignTeamCaptains(context, *h, *event.Comment.Body, 1)
}
