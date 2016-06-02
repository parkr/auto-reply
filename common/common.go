package common

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const accessTokenEnvVar = "GITHUB_ACCESS_TOKEN"

func GitHubToken() string {
	return os.Getenv(accessTokenEnvVar)
}

func NewClient() *github.Client {
	if token := GitHubToken(); token != "" {
		return github.NewClient(oauth2.NewClient(
			oauth2.NoContext,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
			),
		))
	} else {
		log.Fatalf("%s required", accessTokenEnvVar)
		return nil
	}
}

func SliceLookup(data []string) map[string]bool {
	mapping := map[string]bool{}
	for _, datum := range data {
		mapping[datum] = true
	}
	return mapping
}

func ErrorFromResponse(res *github.Response, err error) error {
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected error code: %d", res.StatusCode)
	}

	return nil
}
