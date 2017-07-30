// deprecate closes issues on deprecated repos and leaves a nice comment for the user.
package deprecate

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var (
	deprecatedRepos = map[string]string{
		"jekyll/jekyll-help": `This repository is no longer maintained. If you're still experiencing this problem, please search for your issue on [Jekyll Talk](https://talk.jekyllrb.com/), our new community forum. If it isn't there, feel free to post to the Help category and someone will assist you. Thanks!`,
	}
)

func DeprecateOldRepos(context *ctx.Context, event interface{}) error {
	issue, ok := event.(*github.IssuesEvent)
	if !ok {
		return context.NewError("DeprecateOldRepos: not an issue event")
	}

	if *issue.Action != "opened" {
		return context.NewError("DeprecateOldRepos: issue event's action is not 'opened'")
	}

	owner, name, number := *issue.Repo.Owner.Login, *issue.Repo.Name, *issue.Issue.Number
	if message, ok := deprecatedRepos[*issue.Repo.FullName]; ok {
		err := commentAndClose(context, owner, name, number, message)
		if err != nil {
			return err
		}
	}

	return nil
}

func commentAndClose(context *ctx.Context, owner, name string, number int, message string) error {
	ref := fmt.Sprintf("%s/%s#%d", owner, name, number)
	_, _, err := context.GitHub.Issues.CreateComment(
		context.Context(),
		owner, name, number,
		&github.IssueComment{Body: github.String(message)},
	)
	if err != nil {
		return context.NewError("DeprecateOldRepos: error commenting on %s: %v", ref, err)
	}
	_, _, err = context.GitHub.Issues.Edit(
		context.Context(),
		owner, name, number,
		&github.IssueRequest{State: github.String("closed")},
	)
	if err != nil {
		return context.NewError("DeprecateOldRepos: error closing %s: %v", ref, err)
	}
	return nil
}
