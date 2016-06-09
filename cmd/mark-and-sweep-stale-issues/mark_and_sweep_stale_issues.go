package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/common"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/labeler"
)

var (
	staleableLabels = []string{
		"bug",
		"discussion",
		"documentation",
		"downstream-issue",
		"downstream:github-pages",
		"duplicate",
		"enhancement",
		"feature",
		"suggestion",
		"needs-work",
		"not-reproduced",
		"old-stable",
		"pending-feedback",
		"plugin-feature",
		"question",
		"stale",
		"support",
		"undetermined",
		"upstream:issue",
		"wont-fix",
		`¯\_(ツ)_/¯`,
	}

	repos = []*github.Repository{
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-import"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("github-metadata"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-redirect-from"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-feed"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-compose"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-watch"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-sitemap"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-sass-converter"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jemoji"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-gist"),
		},
		&github.Repository{
			Owner: &github.User{Login: github.String("jekyll")},
			Name:  github.String("jekyll-coffeescript"),
		},
	}

	oneMonthAgo = time.Now().AddDate(0, -1, 0)

	staleIssuesListOptions = &github.IssueListByRepoOptions{
		State:       "open",
		Sort:        "updated",
		Direction:   "asc",
		ListOptions: github.ListOptions{PerPage: 200},
	}

	staleIssueComment = &github.IssueComment{
		Body: github.String(`
This issue has been automatically marked as stale because it has not been commented on for at least
one month.

The resources of the Jekyll team are limited, and so we are asking for your help.

If you can still reproduce this error on the <code>3.1-stable</code> or <code>master</code> branch,
please reply with all of the information you have about it in order to keep the issue open.

If this is a feature request, please consider building it first as a plugin. Jekyll 3 introduced
[hooks](http://jekyllrb.com/docs/plugins/#hooks) which provide convenient access points throughout
the Jekyll build pipeline whereby most needs can be fulfilled. If this is something that cannot be
built as a plugin, then please provide more information about why in order to keep this issue open.

Thank you for all your contributions.
`),
	}
)

func main() {
	var actuallyDoIt bool
	flag.BoolVar(&actuallyDoIt, "f", false, "Whether to actually mark the issues or close them.")
	flag.Parse()

	client := ctx.NewClient()

	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go markAndSweep(&wg, client, repo, actuallyDoIt)
	}
	wg.Wait()
}

func markAndSweep(wg *sync.WaitGroup, client *github.Client, repo *github.Repository, actuallyDoIt bool) {
	owner, name, nonStaleIssues := *repo.Owner.Login, *repo.Name, 0

	issues, resp, err := client.Issues.ListByRepo(owner, name, staleIssuesListOptions)
	err = common.ErrorFromResponse(resp, err)
	if err != nil {
		log.Fatalf("could not list issues for %s/%s: %v", owner, name, err)
	}

	if len(issues) == 0 {
		log.Printf("no issues for %s/%s", owner, name)
		wg.Done()
		return
	}

	for _, issue := range issues {
		if isStale(issue) {
			if hasStaleLabel(issue) {
				// Close.
				if actuallyDoIt {
					number := *issue.Number
					log.Printf("%s is stale & notified (closing).", linkify(owner, name, number))
					_, resp, err := client.Issues.Edit(
						owner,
						name,
						number,
						&github.IssueRequest{State: github.String("closed")},
					)
					err = common.ErrorFromResponse(resp, err)
					if err != nil {
						log.Fatalf("!!! could not close issue %s: %v", linkify(owner, name, number), err)
					}
				} else {
					log.Printf("%s is stale & notified (dry-run).", linkify(owner, name, *issue.Number))
				}
			} else {
				// Mark as stale.
				if actuallyDoIt {
					log.Printf("%s is stale (marking).", linkify(owner, name, *issue.Number))
					labeler.AddLabels(client, owner, name, *issue.Number, []string{"stale"})
					client.Issues.CreateComment(owner, name, *issue.Number, staleIssueComment)
				} else {
					log.Printf("%s is stale (dry-run).", linkify(owner, name, *issue.Number))
				}
			}
		} else {
			nonStaleIssues += 1
		}
	}

	log.Printf("%s -- ignored non-stale issues: %d", linkify(owner, name, -1), nonStaleIssues)

	wg.Done()
}

func linkify(owner, name string, number int) string {
	if number == -1 {
		return fmt.Sprintf("https://github.com/%s/%s", owner, name)
	} else {
		return fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, name, number)
	}
}

func isStale(issue github.Issue) bool {
	return issue.PullRequestLinks == nil && !isUpdatedInLastMonth(*issue.UpdatedAt) && hasStaleableLabel(issue)
}

func isUpdatedInLastMonth(updatedAt time.Time) bool {
	return updatedAt.Unix() >= oneMonthAgo.Unix()
}

func hasStaleableLabel(issue github.Issue) bool {
	if issue.Labels == nil {
		return true
	}

	if len(issue.Labels) == 0 {
		return true
	}

	for _, staleableLabel := range staleableLabels {
		for _, issueLabel := range issue.Labels {
			if *issueLabel.Name == staleableLabel {
				return true
			}
		}
	}

	return false
}

func hasStaleLabel(issue github.Issue) bool {
	if issue.Labels == nil {
		return false
	}

	for _, label := range issue.Labels {
		if *label.Name == "stale" {
			return true
		}
	}

	return false
}
