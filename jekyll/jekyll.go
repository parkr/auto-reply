// jekyll is the configuration of handlers and such specific to the org's requirements. This is what you should copy and customize.
package jekyll

import (
	"github.com/parkr/auto-reply/affinity"
	"github.com/parkr/auto-reply/autopull"
	"github.com/parkr/auto-reply/chlog"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/hooks"
	"github.com/parkr/auto-reply/labeler"
	"github.com/parkr/auto-reply/lgtm"

	"github.com/parkr/auto-reply/jekyll/deprecate"
	"github.com/parkr/auto-reply/jekyll/issuecomment"
)

var lgtmEnabledRepos = []lgtm.Repo{
	{Owner: "jekyll", Name: "jekyll", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-coffeescript", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-compose", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-docs", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-feed", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-gist", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-import", Quorum: 1},
	{Owner: "jekyll", Name: "jekyll-mentions", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-opal", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-paginate", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-redirect-from", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-sass-converter", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-sitemap", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-textile-converter", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-watch", Quorum: 2},
	{Owner: "jekyll", Name: "github-metadata", Quorum: 2},
	{Owner: "jekyll", Name: "jemoji", Quorum: 1},
	{Owner: "jekyll", Name: "mercenary", Quorum: 1},
	{Owner: "jekyll", Name: "minima", Quorum: 1},
}

var jekyllOrgEventHandlers = hooks.EventHandlerMap{
	hooks.CreateEvent: {chlog.CreateReleaseOnTagHandler},
	hooks.IssuesEvent: {deprecate.DeprecateOldRepos},
	hooks.IssueCommentEvent: {
		issuecomment.PendingFeedbackUnlabeler,
		issuecomment.StaleUnlabeler,
		chlog.MergeAndLabel,
		lgtm.NewIssueCommentHandler(lgtmEnabledRepos),
	},
	hooks.PushEvent: {autopull.AutomaticallyCreatePullRequest("jekyll/jekyll")},
	hooks.PullRequestEvent: {
		labeler.PendingRebaseNeedsWorkPRUnlabeler,
		lgtm.NewPullRequestHandler(lgtmEnabledRepos),
	},
}

func jekyllAffinityHandler(context *ctx.Context) *affinity.Handler {
	handler := &affinity.Handler{}

	//handler.AddRepo("jekyll", "jekyll")

	handler.AddTeam(context, 1961060) // @jekyll/build
	handler.AddTeam(context, 1961072) // @jekyll/documentation
	handler.AddTeam(context, 1961061) // @jekyll/ecosystem
	handler.AddTeam(context, 1961065) // @jekyll/performance
	handler.AddTeam(context, 1961059) // @jekyll/stability
	handler.AddTeam(context, 1116640) // @jekyll/windows

	context.Log("affinity teams: %q", handler.GetTeams())
	context.Log("affinity team repos: %q", handler.GetRepos())

	return handler
}

func NewJekyllOrgHandler(context *ctx.Context) *hooks.GlobalHandler {
	affinityHandler := jekyllAffinityHandler(context)
	jekyllOrgEventHandlers.AddHandler(hooks.IssuesEvent, affinityHandler.AssignIssueToAffinityTeamCaptain)
	jekyllOrgEventHandlers.AddHandler(hooks.IssueCommentEvent, affinityHandler.AssignIssueToAffinityTeamCaptainFromComment)
	jekyllOrgEventHandlers.AddHandler(hooks.PullRequestEvent, affinityHandler.AssignPRToAffinityTeamCaptain)

	return &hooks.GlobalHandler{
		Context:       context,
		EventHandlers: jekyllOrgEventHandlers,
	}
}
