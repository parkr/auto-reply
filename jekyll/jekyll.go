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

var jekyllOrgEventHandlers = map[hooks.EventType][]hooks.EventHandler{
	hooks.IssuesEvent: {deprecate.DeprecateOldRepos},
	hooks.IssueCommentEvent: {
		issuecomment.PendingFeedbackUnlabeler, issuecomment.StaleUnlabeler,
		issuecomment.MergeAndLabel, lgtm.IssueCommentHandler,
	},
	hooks.PushEvent:        {autopull.AutomaticallyCreatePullRequest("jekyll/jekyll")},
	hooks.PullRequestEvent: {labeler.PendingRebaseNeedsWorkPRUnlabeler},
}

func NewJekyllOrgHandler(context *ctx.Context) *hooks.GlobalHandler {
	return &hooks.GlobalHandler{
		Context:       context,
		EventHandlers: jekyllOrgEventHandlers,
	}
}
