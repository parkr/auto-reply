// ctx is magic; it is basically my own "context" package before I realized that "context" existed.
// ctx.Context is the main construct. It keeps track of information pertinent to the request.
// It should all eventually be replaced by context.Context from the Go stdlib.
package ctx

import (
	gocontext "context"
	"fmt"
	"log"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/google/go-github/v42/github"
)

type Context struct {
	GitHub   *github.Client
	Statsd   *statsd.Client
	RubyGems *rubyGemsClient
	Repo     repoRef
	Issue    issueRef

	currentlyAuthedGitHubUser *github.User
}

func (c *Context) NewError(format string, args ...interface{}) error {
	c.Log(format, args...)
	return fmt.Errorf(format, args...)
}

func (c *Context) Log(format string, args ...interface{}) {
	log.Println(fmt.Sprintf(format, args...))
}

func (c *Context) Context() gocontext.Context {
	return gocontext.Background()
}

func NewDefaultContext() *Context {
	return &Context{
		GitHub:   NewClient(),
		Statsd:   NewStatsd(),
		RubyGems: NewRubyGemsClient(),
	}
}

func WithIssue(owner, repo string, num int) *Context {
	context := NewDefaultContext()
	context.SetRepo(owner, repo)
	context.SetIssue(owner, repo, num)
	return context
}

func WithRepo(owner, repo string) *Context {
	context := NewDefaultContext()
	context.SetRepo(owner, repo)
	return context
}
