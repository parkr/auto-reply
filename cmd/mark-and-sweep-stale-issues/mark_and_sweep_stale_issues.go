// A command-line utility to mark and sweep Jekyll issues for staleness.
package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/stale"
	"golang.org/x/sync/errgroup"
)

type repo struct {
	Owner, Name string
}

var (
	nonStaleableLabels = []string{
		"has-pull-request",
		"pinned",
		"security",
	}

	defaultRepos = []repo{
		repo{"jekyll", "jekyll"},
		repo{"jekyll", "jekyll-admin"},
		repo{"jekyll", "jekyll-import"},
		repo{"jekyll", "github-metadata"},
		repo{"jekyll", "jekyll-redirect-from"},
		repo{"jekyll", "jekyll-feed"},
		repo{"jekyll", "jekyll-compose"},
		repo{"jekyll", "jekyll-watch"},
		repo{"jekyll", "jekyll-seo-tag"},
		repo{"jekyll", "jekyll-sitemap"},
		repo{"jekyll", "jekyll-sass-converter"},
		repo{"jekyll", "jemoji"},
		repo{"jekyll", "jekyll-gist"},
		repo{"jekyll", "jekyll-coffeescript"},
		repo{"jekyll", "plugins"},
	}

	twoMonthsAgo = time.Now().AddDate(0, -2, 0)

	staleIssuesListOptions = &github.IssueListByRepoOptions{
		State:       "open",
		Sort:        "updated",
		Direction:   "asc",
		ListOptions: github.ListOptions{PerPage: 200},
	}

	staleJekyllIssueComment = &github.IssueComment{
		Body: github.String(`
This issue has been automatically marked as stale because it has not been commented on for at least two months.

The resources of the Jekyll team are limited, and so we are asking for your help.

If this is a **bug** and you can still reproduce this error on the <code>3.3-stable</code> or <code>master</code> branch, please reply with all of the information you have about it in order to keep the issue open.

If this is a **feature request**, please consider building it first as a plugin. Jekyll 3 introduced [hooks](http://jekyllrb.com/docs/plugins/#hooks) which provide convenient access points throughout the Jekyll build pipeline whereby most needs can be fulfilled. If this is something that cannot be built as a plugin, then please provide more information about why in order to keep this issue open.

This issue will automatically be closed in two months if no further activity occurs. Thank you for all your contributions.
`),
	}

	staleNonJekyllIssueComment = &github.IssueComment{
		Body: github.String(`
This issue has been automatically marked as stale because it has not been commented on for at least two months.

The resources of the Jekyll team are limited, and so we are asking for your help.

If this is a **bug** and you can still reproduce this error on the <code>master</code> branch, please reply with all of the information you have about it in order to keep the issue open.

If this is a feature request, please consider whether it can be accomplished in another way. If it cannot, please elaborate on why it is core to this project and why you feel more than 80% of users would find this beneficial.

This issue will automatically be closed in two months if no further activity occurs. Thank you for all your contributions.
`),
	}
)

func main() {
	var actuallyDoIt bool
	flag.BoolVar(&actuallyDoIt, "f", false, "Whether to actually mark the issues or close them.")
	var inputRepos string
	flag.StringVar(&inputRepos, "repos", "", "Specify a list of comma-separated repo name/owner pairs, e.g. 'jekyll/jekyll-admin'.")
	flag.Parse()

	if ctx.NewDefaultContext().GitHub == nil {
		log.Fatalln("cannot proceed without github client")
	}

	var repos []repo
	if inputRepos != "" {
		for _, nwo := range strings.Split(inputRepos, ",") {
			pieces := strings.Split(nwo, "/")
			repos = append(repos, repo{Owner: pieces[0], Name: pieces[1]})
		}
	} else {
		repos = defaultRepos
	}

	wg, _ := errgroup.WithContext(context.Background())
	for _, repo := range repos {
		repo := repo
		wg.Go(func() error {
			return stale.MarkAndCloseForRepo(
				ctx.WithRepo(repo.Owner, repo.Name),
				stale.Configuration{
					Perform:             actuallyDoIt,
					ExemptLabels:        nonStaleableLabels,
					DormantDuration:     time.Since(twoMonthsAgo),
					NotificationComment: staleIssueComment(repo.Owner, repo.Name),
				},
			)
		})
	}
	if err := wg.Wait(); err != nil {
		log.Fatal("error: ", err)
	}
}

func staleIssueComment(repoOwner, repoName string) *github.IssueComment {
	if repoName == "jekyll" {
		return staleJekyllIssueComment
	} else {
		return staleNonJekyllIssueComment
	}
}
