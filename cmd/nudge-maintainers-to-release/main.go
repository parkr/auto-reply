package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/jekyll"
	"github.com/parkr/auto-reply/releases"
	"github.com/parkr/auto-reply/search"
	"github.com/parkr/auto-reply/sentry"
	"github.com/parkr/githubapi/githubsearch"
	"golang.org/x/sync/errgroup"
)

var (
	defaultRepos = jekyll.DefaultRepos

	threeMonthsAgoUnix = time.Now().AddDate(0, -3, 0).Unix()

	issueTitle  = "Time for a new release"
	issueLabels = []string{"release"}

	issueBodyTemplate = template.Must(template.New("issueBodyTemplate").Parse(`
Hello, maintainers! :wave:

By my calculations, it's time for a new release of {{.Repo.Name}}. {{if gt .CommitsOnMasterSinceLatestRelease 100}}There have been {{.CommitsOnMasterSinceLatestRelease}} commits{{else}}It's been over 3 months{{end}} since the last release, {{.LatestRelease.TagName}}.

What else is left to be done before a new release can be made? Please make sure to update History.markdown too if it's not already updated.

Thanks! :revolving_hearts: :sparkles:
`))
)

type templateInfo struct {
	Repo                              jekyll.Repository
	CommitsOnMasterSinceLatestRelease int
	LatestRelease                     *github.RepositoryRelease
}

func main() {
	var perform bool
	flag.BoolVar(&perform, "f", false, "Whether to actually file issues.")
	var inputRepos string
	flag.StringVar(&inputRepos, "repos", "", "Specify a list of comma-separated repo name/owner pairs, e.g. 'jekyll/jekyll-import'.")
	flag.Parse()

	// Get latest 10 releases.
	// Sort releases by semver version, taking highest one.
	//
	// Has there been 100 commits since this release? If so, make an issue.
	// Has at least 1 commit been made since this release & is this release at least 2 month old? If so, make an issue.

	var repos []jekyll.Repository
	if inputRepos == "" {
		repos = defaultRepos
	}

	log.SetPrefix("nudge-maintainers-to-release: ")

	sentryClient, err := sentry.NewClient(map[string]string{
		"app":          "nudge-maintainers-to-release",
		"inputRepos":   inputRepos,
		"actuallyDoIt": fmt.Sprintf("%t", perform),
	})
	if err != nil {
		panic(err)
	}
	sentryClient.Recover(func() error {
		context := ctx.NewDefaultContext()
		if context.GitHub == nil {
			return errors.New("cannot proceed without github client")
		}

		if inputRepos != "" {
			repos = []jekyll.Repository{}
			for _, inputRepo := range strings.Split(inputRepos, ",") {
				pieces := strings.Split(inputRepo, "/")
				if len(pieces) != 2 {
					return fmt.Errorf("input repo %q is improperly formed", inputRepo)
				}
				repos = append(repos, jekyll.NewRepository(pieces[0], pieces[1]))
			}
		}

		wg, _ := errgroup.WithContext(context.Context())
		for _, repo := range repos {
			repo := repo
			wg.Go(func() error {
				latestRelease, err := releases.LatestRelease(context, repo)
				if err != nil {
					log.Printf("%s error fetching latest release: %+v", repo, err)
					return err
				}
				if latestRelease == nil {
					log.Printf("%s has no releases", repo)
					return nil
				}

				commitsSinceLatestRelease, err := releases.CommitsSinceRelease(context, repo, latestRelease)
				if err != nil {
					log.Printf("%s error fetching commits since latest release: %+v", repo, err)
					return err
				}

				if commitsSinceLatestRelease > 100 || (commitsSinceLatestRelease >= 3 && latestRelease.GetCreatedAt().Unix() <= threeMonthsAgoUnix) {
					if perform {
						err := fileIssue(context, templateInfo{
							Repo:                              repo,
							LatestRelease:                     latestRelease,
							CommitsOnMasterSinceLatestRelease: commitsSinceLatestRelease,
						})
						if err != nil {
							log.Printf("%s: nudged maintainers (release=%s commits=%d released_on=%s)",
								repo,
								latestRelease.GetTagName(),
								commitsSinceLatestRelease,
								latestRelease.GetCreatedAt(),
							)
						}
						return err
					} else {
						log.Printf("%s is in need of a nudge (release=%s commits=%d released_on=%s)",
							repo,
							latestRelease.GetTagName(),
							commitsSinceLatestRelease,
							latestRelease.GetCreatedAt(),
						)
					}
				} else {
					log.Printf("%s is NOT in need of a nudge: (release=%s commits=%d released_on=%s)",
						repo,
						latestRelease.GetTagName(),
						commitsSinceLatestRelease,
						latestRelease.GetCreatedAt(),
					)
				}

				return nil
			})
		}
		return wg.Wait()
	})
}

func fileIssue(context *ctx.Context, issueInfo templateInfo) error {
	if issue := getReleaseNudgeIssue(context, issueInfo.Repo); issue != nil {
		return fmt.Errorf("%s: issue already exists: %s", issueInfo.Repo, issue.GetHTMLURL())
	}

	var body bytes.Buffer
	if err := issueBodyTemplate.Execute(&body, issueInfo); err != nil {
		return fmt.Errorf("%s: error executing template: %+v", issueInfo.Repo, err)
	}

	issue, _, err := context.GitHub.Issues.Create(
		context.Context(),
		issueInfo.Repo.Owner(), issueInfo.Repo.Name(),
		&github.IssueRequest{
			Title:  &issueTitle,
			Labels: &issueLabels,
			Body:   github.String(body.String()),
		},
	)
	if err != nil {
		return fmt.Errorf("%s: error filing issue: %+v", issueInfo.Repo, err)
	}

	log.Printf("%s filed %s", issueInfo.Repo, issue.GetHTMLURL())
	return nil
}

func getReleaseNudgeIssue(context *ctx.Context, repo jekyll.Repository) *github.Issue {
	query := githubsearch.IssueSearchParameters{
		Type:       githubsearch.Issue,
		Scope:      githubsearch.TitleScope,
		Author:     context.CurrentlyAuthedGitHubUser().GetLogin(),
		State:      githubsearch.Open,
		Repository: &githubsearch.RepositoryName{Owner: repo.Owner(), Name: repo.Name()},
		Query:      issueTitle,
	}
	issues, err := search.GitHubIssues(context, query)
	if err != nil {
		log.Printf("%s: error searching %s: %+v", repo, query, err)
		return nil
	}
	if len(issues) > 0 {
		return &(issues[0])
	}
	return nil
}
