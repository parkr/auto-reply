package ctx

import (
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const githubAccessTokenEnvVar = "GITHUB_ACCESS_TOKEN"

func (c *Context) GitHubAuthedAs(login string) bool {
	return c.CurrentlyAuthedGitHubUser().GetLogin() == login
}

func (c *Context) CurrentlyAuthedGitHubUser() *github.User {
	if c.currentlyAuthedGitHubUser == nil {
		currentlyAuthedUser, _, err := c.GitHub.Users.Get(c.Context(), "")
		if err != nil {
			c.Log("couldn't fetch currently-auth'd user: %v", err)
			return nil
		}
		c.currentlyAuthedGitHubUser = currentlyAuthedUser
	}

	return c.currentlyAuthedGitHubUser
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
