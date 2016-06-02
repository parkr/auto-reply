package ctx

import (
	"log"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const githubAccessTokenEnvVar = "GITHUB_ACCESS_TOKEN"

type Context struct {
	GitHub *github.Client
	Statsd *statsd.Client
}

func NewDefaultContext() *Context {
	return &Context{
		GitHub: NewClient(),
		Statsd: NewStatsd(),
	}
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
