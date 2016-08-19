// ctx is magic; it is basically my own "context" package before I realied that "context" existed.
// ctx.Context is the main construct. It keeps track of information pertinent to the request.
// It should all eventually be replaced by context.Context from the Go stdlib.
package ctx

import (
	"fmt"
	"log"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const githubAccessTokenEnvVar = "GITHUB_ACCESS_TOKEN"

var (
	countRate float64 = 1
	noTags            = []string{}
)

// issueRef is used to refer to an issue or pull request
type issueRef struct {
	Author      string
	Owner, Repo string
	Num         int
}

func (r issueRef) String() string {
	return fmt.Sprintf("%s/%s#%d", r.Owner, r.Repo, r.Num)
}

func (r issueRef) IsEmpty() bool {
	return r.Owner == "" || r.Repo == "" || r.Num == 0
}

type Context struct {
	GitHub *github.Client
	Statsd *statsd.Client
	Issue  issueRef
}

func (c *Context) IncrStat(name string) {
	c.CountStat(name, 1)
}

func (c *Context) CountStat(name string, value int64) {
	if c.Statsd != nil {
		c.Statsd.Count(name, value, noTags, countRate)
	}
}

func (c *Context) NewError(format string, args ...interface{}) error {
	c.Log(format, args...)
	return fmt.Errorf(format, args...)
}

func (c *Context) Log(format string, args ...interface{}) {
	log.Println(fmt.Sprintf(format, args...))
}

func (c *Context) SetIssue(owner, repo string, num int) {
	c.Issue = issueRef{
		Owner: owner,
		Repo:  repo,
		Num:   num,
	}
}

func (c *Context) SetAuthor(author string) {
	c.Issue.Author = author
}

func NewDefaultContext() *Context {
	return &Context{
		GitHub: NewClient(),
		Statsd: NewStatsd(),
	}
}

func WithIssue(owner, repo string, num int) *Context {
	context := NewDefaultContext()
	context.SetIssue(owner, repo, num)
	return context
}

func GitHubToken() string {
	return os.Getenv(githubAccessTokenEnvVar)
}

func NewClient() *github.Client {
	if token := GitHubToken(); token != "" {
		return github.NewClient(oauth2.NewClient(
			oauth2.NoContext,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: GitHubToken()},
			),
		))
	} else {
		log.Fatalf("%s required", githubAccessTokenEnvVar)
		return nil
	}
}

func NewStatsd() *statsd.Client {
	client, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	client.Namespace = "autoreply."
	return client
}
