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

var jekyllOrgEventHandlers = map[hooks.EventType][]hooks.EventHandler{
	hooks.CreateEvent: {chlog.CreateReleaseOnTagHandler},
	hooks.IssuesEvent: {
		affinity.AssignIssueToAffinityTeamCaptain,
		deprecate.DeprecateOldRepos,
	},
	hooks.IssueCommentEvent: {
		affinity.AssignIssueToAffinityTeamCaptainFromComment,
		issuecomment.PendingFeedbackUnlabeler,
		issuecomment.StaleUnlabeler,
		chlog.MergeAndLabel,
		lgtm.NewIssueCommentHandler(lgtmEnabledRepos),
	},
	hooks.PushEvent: {autopull.AutomaticallyCreatePullRequest("jekyll/jekyll")},
	hooks.PullRequestEvent: {
		affinity.AssignPRToAffinityTeamCaptain,
		labeler.PendingRebaseNeedsWorkPRUnlabeler,
		lgtm.NewPullRequestHandler(lgtmEnabledRepos),
	},
}

func NewJekyllOrgHandler(context *ctx.Context) *hooks.GlobalHandler {
	affinity.Teams = []affinity.Team{
		affinity.Team{ID: 0, Name: "Build", Mention: "@jekyll/build"},
		affinity.Team{ID: 0, Name: "Documentation", Mention: "@jekyll/documentation"},
		affinity.Team{ID: 0, Name: "Ecosystem", Mention: "@jekyll/ecosystem"},
		affinity.Team{ID: 0, Name: "Performance", Mention: "@jekyll/performance"},
		affinity.Team{ID: 0, Name: "Stability", Mention: "@jekyll/stability"},
		affinity.Team{ID: 0, Name: "Windows", Mention: "@jekyll/windows"},
	}
	return &hooks.GlobalHandler{
		Context:       context,
		EventHandlers: jekyllOrgEventHandlers,
	}
}
