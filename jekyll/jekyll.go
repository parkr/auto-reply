package jekyll

import (
	"github.com/parkr/auto-reply/autopull"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/hooks"
	"github.com/parkr/auto-reply/labeler"
	"github.com/parkr/auto-reply/lgtm"

	"github.com/parkr/auto-reply/jekyll/deprecate"
	"github.com/parkr/auto-reply/jekyll/issuecomment"
)

var lgtmEnabledRepos = []lgtm.Repo{
	//{Owner: "jekyll", Name: "jekyll", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-import", Quorum: 1},
	{Owner: "jekyll", Name: "jekyll-feed", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-sitemap", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-mentions", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-watch", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-compose", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-paginate", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-gist", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-coffeescript", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-opal", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-sass-converter", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-textile-converter", Quorum: 2},
	{Owner: "jekyll", Name: "jekyll-redirect-from", Quorum: 2},
	{Owner: "jekyll", Name: "github-metadata", Quorum: 2},
	{Owner: "jekyll", Name: "jemoji", Quorum: 1},
	{Owner: "jekyll", Name: "mercenary", Quorum: 1},
}

var jekyllOrgEventHandlers = map[hooks.EventType][]hooks.EventHandler{
	hooks.IssuesEvent: {deprecate.DeprecateOldRepos},
	hooks.IssueCommentEvent: {
		issuecomment.PendingFeedbackUnlabeler, issuecomment.StaleUnlabeler,
		issuecomment.MergeAndLabel, lgtm.NewIssueCommentHandler(lgtmEnabledRepos),
	},
	hooks.PushEvent: {autopull.AutomaticallyCreatePullRequest("jekyll/jekyll")},
	hooks.PullRequestEvent: {
		labeler.PendingRebaseNeedsWorkPRUnlabeler,
		lgtm.NewPullRequestHandler(lgtmEnabledRepos),
	},
}

func NewJekyllOrgHandler(context *ctx.Context) *hooks.GlobalHandler {
	return &hooks.GlobalHandler{
		Context:       context,
		EventHandlers: jekyllOrgEventHandlers,
	}
}
