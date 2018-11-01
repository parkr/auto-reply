// jekyll is the configuration of handlers and such specific to the org's requirements. This is what you should copy and customize.
package jekyll

import (
	"fmt"

	"github.com/parkr/auto-reply/affinity"
	"github.com/parkr/auto-reply/autopull"
	"github.com/parkr/auto-reply/chlog"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/hooks"
	"github.com/parkr/auto-reply/labeler"
	"github.com/parkr/auto-reply/lgtm"
	"github.com/parkr/auto-reply/travis"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/jekyll/deprecate"
	"github.com/parkr/auto-reply/jekyll/issuecomment"
)

var jekyllOrgEventHandlers = hooks.EventHandlerMap{
	hooks.CreateEvent: {chlog.CreateReleaseOnTagHandler},
	hooks.IssuesEvent: {deprecate.DeprecateOldRepos},
	hooks.IssueCommentEvent: {
		issuecomment.PendingFeedbackUnlabeler,
		issuecomment.StaleUnlabeler,
		chlog.MergeAndLabel,
	},
	hooks.PullRequestEvent: {
		labeler.IssueHasPullRequestLabeler,
		labeler.PendingRebaseNeedsWorkPRUnlabeler,
	},
	hooks.ReleaseEvent: {chlog.CloseMilestoneOnRelease},
	hooks.StatusEvent:  {statStatus, travis.FailingFmtBuildHandler},
}

func statStatus(context *ctx.Context, payload interface{}) error {
	status, ok := payload.(*github.StatusEvent)
	if !ok {
		return context.NewError("statStatus: not an status event")
	}

	context.SetIssue(*status.Repo.Owner.Login, *status.Repo.Name, -1)

	if context.Statsd != nil {
		statName := fmt.Sprintf("status.%s", *status.State)
		context.Log("context.Statsd.Count(%s, 1, []string{context:%s, repo:%s}, 1)", statName, *status.Context, context.Issue)
		return context.Statsd.Incr(
			statName,
			[]string{
				"context:" + *status.Context,
				"repo:" + context.Issue.String(),
			},
			float64(1.0), // rate..?
		)
	}
	return nil
}

func jekyllAffinityHandler(context *ctx.Context) *affinity.Handler {
	handler := &affinity.Handler{}

	handler.AddRepo("jekyll", "jekyll")
	handler.AddRepo("jekyll", "minima")

	handler.AddTeam(context, 1961060) // @jekyll/build
	handler.AddTeam(context, 1961072) // @jekyll/documentation
	handler.AddTeam(context, 1961061) // @jekyll/ecosystem
	handler.AddTeam(context, 1961065) // @jekyll/performance
	handler.AddTeam(context, 1961059) // @jekyll/stability
	handler.AddTeam(context, 1116640) // @jekyll/windows

	context.Log("affinity teams: %+v", handler.GetTeams())
	context.Log("affinity team repos: %+v", handler.GetRepos())

	return handler
}

func newLgtmHandler() *lgtm.Handler {
	handler := &lgtm.Handler{}

	handler.AddRepo("jekyll", "jekyll", 2)
	handler.AddRepo("jekyll", "jekyll-coffeescript", 2)
	handler.AddRepo("jekyll", "jekyll-compose", 1)
	handler.AddRepo("jekyll", "jekyll-commonmark", 1)
	handler.AddRepo("jekyll", "jekyll-docs", 1)
	handler.AddRepo("jekyll", "jekyll-feed", 1)
	handler.AddRepo("jekyll", "jekyll-gist", 2)
	handler.AddRepo("jekyll", "jekyll-import", 1)
	handler.AddRepo("jekyll", "jekyll-mentions", 2)
	handler.AddRepo("jekyll", "jekyll-opal", 2)
	handler.AddRepo("jekyll", "jekyll-paginate", 2)
	handler.AddRepo("jekyll", "jekyll-redirect-from", 2)
	handler.AddRepo("jekyll", "jekyll-sass-converter", 2)
	handler.AddRepo("jekyll", "jekyll-seo-tag", 1)
	handler.AddRepo("jekyll", "jekyll-sitemap", 2)
	handler.AddRepo("jekyll", "jekyll-textile-converter", 2)
	handler.AddRepo("jekyll", "jekyll-watch", 2)
	handler.AddRepo("jekyll", "github-metadata", 2)
	handler.AddRepo("jekyll", "jemoji", 1)
	handler.AddRepo("jekyll", "mercenary", 1)
	handler.AddRepo("jekyll", "minima", 1)
	handler.AddRepo("jekyll", "plugins", 1)

	return handler
}

func NewJekyllOrgHandler(context *ctx.Context) *hooks.GlobalHandler {
	affinityHandler := jekyllAffinityHandler(context)
	jekyllOrgEventHandlers.AddHandler(hooks.IssuesEvent, affinityHandler.AssignIssueToAffinityTeamCaptain)
	jekyllOrgEventHandlers.AddHandler(hooks.IssueCommentEvent, affinityHandler.AssignIssueToAffinityTeamCaptainFromComment)
	jekyllOrgEventHandlers.AddHandler(hooks.PullRequestEvent, affinityHandler.AssignPRToAffinityTeamCaptain)
	jekyllOrgEventHandlers.AddHandler(hooks.PullRequestEvent, affinityHandler.RequestReviewFromAffinityTeamCaptains)

	lgtmHandler := newLgtmHandler()
	jekyllOrgEventHandlers.AddHandler(hooks.PullRequestReviewEvent, lgtmHandler.PullRequestReviewHandler)

	autopullHandler := autopull.Handler{}
	autopullHandler.AcceptAllRepos(true)
	jekyllOrgEventHandlers.AddHandler(hooks.PushEvent, autopullHandler.CreatePullRequestFromPush)

	return &hooks.GlobalHandler{
		Context:       context,
		EventHandlers: jekyllOrgEventHandlers,
	}
}
