package common

import (
	"io"
	"log"
	"os"

	"github.com/parkr/auto-reply/Godeps/_workspace/src/github.com/google/go-github/github"
	"github.com/parkr/auto-reply/Godeps/_workspace/src/golang.org/x/oauth2"
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

func ClearJSONRepoOrgField(reader io.Reader) []byte {
	// workaround for https://github.com/google/go-github/issues/131
	var o map[string]interface{}
	dec := json.NewDecoder(reader)
	dec.UseNumber()
	dec.Decode(&o)
	if o != nil {
		repo := o["repository"]
		if repo != nil {
			if repo, ok := repo.(map[string]interface{}); ok {
				delete(repo, "organization")
			}
		}
	}
	b, _ := json.MarshalIndent(o, "", "  ")
	return b
}
