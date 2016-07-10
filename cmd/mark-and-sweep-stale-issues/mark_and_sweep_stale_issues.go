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
	nonStaleableLabels = []string{
		"has-pull-request",
		"security",
	}

	repos = []*github.Repository{
		repo("jekyll", "jekyll"),
		repo("jekyll", "jekyll-import"),
		repo("jekyll", "github-metadata"),
		repo("jekyll", "jekyll-redirect-from"),
		repo("jekyll", "jekyll-feed"),
		repo("jekyll", "jekyll-compose"),
		repo("jekyll", "jekyll-watch"),
		repo("jekyll", "jekyll-sitemap"),
		repo("jekyll", "jekyll-sass-converter"),
		repo("jekyll", "jemoji"),
		repo("jekyll", "jekyll-gist"),
		repo("jekyll", "jekyll-coffeescript"),
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

If this is a **bug** and you can still reproduce this error on the <code>3.1-stable</code> or <code>master</code> branch, please reply with all of the information you have about it in order to keep the issue open.

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
	owner, name, nonStaleIssues, failedIssues := *repo.Owner.Login, *repo.Name, 0, 0

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
			err := handleStaleIssue(client, repo, issue, actuallyDoIt)
			if err != nil {
				failedIssues += 1
			}
		} else {
			nonStaleIssues += 1
		}
	}

	log.Printf("%s -- ignored non-stale issues: %d", linkify(owner, name, -1), nonStaleIssues)
	log.Printf("%s -- failed issues: %d", linkify(owner, name, -1), failedIssues)

	wg.Done()
}

func handleStaleIssue(client *github.Client, repo *github.Repository, issue *github.Issue, actuallyDoIt bool) error {
	owner, name, number := *repo.Owner.Login, *repo.Name, *issue.Number
	issueRef := linkify(owner, name, number)

	if hasStaleLabel(issue) {
		// Close.
		if actuallyDoIt {
			number := *issue.Number
			log.Printf("%s is stale & notified (closing).", issueRef)
			_, resp, err := client.Issues.Edit(
				owner,
				name,
				number,
				&github.IssueRequest{State: github.String("closed")},
			)
			err = common.ErrorFromResponse(resp, err)
			if err != nil {
				log.Printf("%s !!! could not close issue: %v", issueRef, err)
				return err
			}
		} else {
			log.Printf("%s is stale & notified (dry-run).", issueRef)
		}
	} else {
		// Mark as stale.
		if actuallyDoIt {
			log.Printf("%s is stale (marking).", issueRef)
			err := labeler.AddLabels(client, owner, name, number, []string{"stale"})
			if err != nil {
				log.Printf("%s !!! could not add stale label: %v", issueRef, err)
				return err
			}

			_, resp, err := client.Issues.CreateComment(owner, name, number, staleIssueComment(repo))
			err = common.ErrorFromResponse(resp, err)
			if err != nil {
				log.Printf("%s !!! could not leave comment: %v", issueRef, err)
				return err
			}
		} else {
			log.Printf("%s is stale (dry-run).", issueRef)
		}
	}

	return nil
}

func linkify(owner, name string, number int) string {
	if number == -1 {
		return fmt.Sprintf("https://github.com/%s/%s", owner, name)
	} else {
		return fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, name, number)
	}
}

func isStale(issue *github.Issue) bool {
	return issue.PullRequestLinks == nil && !isUpdatedInLast2Months(*issue.UpdatedAt) && isStaleable(issue)
}

func isUpdatedInLast2Months(updatedAt time.Time) bool {
	return updatedAt.Unix() >= twoMonthsAgo.Unix()
}

func isStaleable(issue *github.Issue) bool {
	if issue.Labels == nil {
		return true
	}

	if len(issue.Labels) == 0 {
		return true
	}

	for _, staleableLabel := range nonStaleableLabels {
		for _, issueLabel := range issue.Labels {
			if *issueLabel.Name == staleableLabel {
				return false
			}
		}
	}

	return true
}

func hasStaleLabel(issue *github.Issue) bool {
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

func staleIssueComment(repo *github.Repository) *github.IssueComment {
	if *repo.Name == "jekyll" {
		return staleJekyllIssueComment
	} else {
		return staleNonJekyllIssueComment
	}
}

func repo(owner, name string) *github.Repository {
	return &github.Repository{
		Owner: &github.User{Login: github.String(owner)},
		Name:  github.String(name),
	}
}
