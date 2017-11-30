package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/jekyll"
	"github.com/parkr/auto-reply/releases"
	"github.com/parkr/auto-reply/sentry"
	"golang.org/x/sync/errgroup"
)

var (
	defaultRepos = jekyll.DefaultRepos

	twoMonthsAgoUnix = time.Now().AddDate(0, -2, 0).Unix()

	issueLabels = []string{"release"}
)

func main() {
	var perform bool
	flag.BoolVar(&perform, "f", false, "Whether to actually file issues.")
	var inputRepos string
	flag.StringVar(&inputRepos, "repos", "", "Specify a list of comma-separated repo name/owner pairs, e.g. 'jekyll/jekyll-admin'.")
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

				if commitsSinceLatestRelease > 100 {
					if perform {
						return fileIssue(context, repo, latestRelease, "Over 100 commits have been made since the last release.")
					} else {
						log.Printf("%s would have been nudged for commits=%d", repo, commitsSinceLatestRelease)
					}
				} else if commitsSinceLatestRelease >= 1 && latestRelease.GetCreatedAt().Unix() <= twoMonthsAgoUnix {
					if perform {
						return fileIssue(context, repo, latestRelease, "The last release was over 2 months ago and there are unreleased commits on master.")
					} else {
						log.Printf("%s would have been nudged for date=%s commits=%d", repo, latestRelease.GetCreatedAt(), commitsSinceLatestRelease)
					}
				} else {
					log.Printf("%s is not in need of a release", repo)
				}

				return nil
			})
		}
		return wg.Wait()
	})
}

func fileIssue(context *ctx.Context, repo jekyll.Repository, latestRelease *github.RepositoryRelease, reason string) error {
	// TODO: does one already exist?
	issue, _, err := context.GitHub.Issues.Create(
		context.Context(),
		repo.Owner(), repo.Name(),
		&github.IssueRequest{
			Title:  github.String("Time for a new release!"),
			Labels: &issueLabels,
			Body: github.String(strings.TrimSpace(fmt.Sprintf(`
Hello, fine maintainers!

You've made some wonderful progress! %s

Would you mind shipping a new release soon so our users can enjoy the optimizations the community has made on master?

Thanks! :revolving_hearts: :sparkles:
`, reason))),
		},
	)
	if err != nil {
		return err
	}

	log.Printf("%s filed %s", repo, issue.HTMLURL)
	return nil
}
