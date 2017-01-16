package dashboard

import (
	"log"
	"os"
	"strings"

	gh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const accessTokenEnvVar = "GITHUB_ACCESS_TOKEN"

var githubClient *gh.Client

type GitHub struct {
	CommitsThisWeek           int    `json:"commits_this_week"`
	OpenPRs                   int    `json:"open_prs"`
	OpenIssues                int    `json:"open_issues"`
	CommitsSinceLatestRelease int    `json:"commits_since_latest_release"`
	LatestReleaseTag          string `json:"latest_release_tag"`
}

func init() {
	githubClient = newGitHubClient()
}

func gitHubToken() string {
	return os.Getenv(accessTokenEnvVar)
}

func newGitHubClient() *gh.Client {
	if token := gitHubToken(); token != "" {
		return gh.NewClient(oauth2.NewClient(
			oauth2.NoContext,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		))
	} else {
		log.Printf("%s required for GitHub", accessTokenEnvVar)
		return nil
	}
}

func github(nwo string) chan *GitHub {
	githubChan := make(chan *GitHub, 1)

	go func() {
		if nwo == "" || githubClient == nil {
			githubChan <- nil
			close(githubChan)
			return
		}
		pieces := strings.Split(nwo, "/")
		owner := pieces[0]
		repo := pieces[1]

		commits, tag := commitsSinceLatestRelease(owner, repo)
		openIssueAndPRCount := openIssues(owner, repo)
		openPRCount := openPRs(nwo)
		githubChan <- &GitHub{
			CommitsThisWeek:           commitsThisWeek(owner, repo),
			OpenPRs:                   openPRCount,
			OpenIssues:                openIssueAndPRCount - openPRCount,
			CommitsSinceLatestRelease: commits,
			LatestReleaseTag:          tag,
		}
		close(githubChan)
	}()

	return githubChan
}

func openIssues(owner, repo string) int {
	repoData, _, err := githubClient.Repositories.Get(owner, repo)
	if err != nil {
		log.Printf("error fetching repo %s/%s: %v", owner, repo, err)
		return -1
	}
	return *repoData.OpenIssuesCount
}

func openPRs(nwo string) int {
	result, _, err := githubClient.Search.Issues(
		"state:open type:pr repo:"+nwo,
		&gh.SearchOptions{Sort: "created", Order: "asc"},
	)
	if err != nil {
		log.Printf("error searching for pr's for %s: %v", nwo, err)
		return -1
	}
	return *result.Total
}

func commitsThisWeek(owner, repo string) int {
	activities, _, err := githubClient.Repositories.ListCommitActivity(owner, repo)
	if err != nil {
		log.Printf("error fetching commits this week for %s/%s: %v", owner, repo, err)
		return -1
	}
	if len(activities) < 1 {
		log.Printf("error fetching commits this week for %s/%s: no results", owner, repo)
		return -1
	}
	return *activities[len(activities)-1].Total
}

func commitsSinceLatestRelease(owner, repo string) (int, string) {
	release, _, err := githubClient.Repositories.GetLatestRelease(owner, repo)
	if err != nil {
		log.Printf("error fetching commits since latest release for %s/%s: %v", owner, repo, err)
		return -1, ""
	}
	comparison, _, err := githubClient.Repositories.CompareCommits(
		owner, repo,
		*release.TagName, "master",
	)
	if err != nil {
		log.Printf("error fetching commit comparison for %s...master for %s/%s: %v", *release.TagName, owner, repo, err)
		return -1, *release.TagName
	}
	return *comparison.TotalCommits, *release.TagName
}
