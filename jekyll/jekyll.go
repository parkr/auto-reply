package jekyll

import (
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/hooks"

	"github.com/parkr/auto-reply/jekyll/deprecate"
)

var jekyllOrgEventHandlers = map[hooks.EventType][]hooks.EventHandler{
	hooks.IssuesEvent: {deprecate.DeprecateOldRepos},
}

func NewJekyllOrgHandler(context *ctx.Context) *hooks.GlobalHandler {
	return &hooks.GlobalHandler{
		Context:       context,
		EventHandlers: jekyllOrgEventHandlers,
	}
}
