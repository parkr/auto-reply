// A command-line utility to lock old issues.
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/freeze"
)

type repository struct {
	Owner, Name string
}

var (
	defaultRepos = []repository{
		repository{"jekyll", "jekyll"},
		repository{"jekyll", "jekyll-admin"},
		repository{"jekyll", "jekyll-import"},
		repository{"jekyll", "github-metadata"},
		repository{"jekyll", "jekyll-redirect-from"},
		repository{"jekyll", "jekyll-feed"},
		repository{"jekyll", "jekyll-compose"},
		repository{"jekyll", "jekyll-watch"},
		repository{"jekyll", "jekyll-seo-tag"},
		repository{"jekyll", "jekyll-sitemap"},
		repository{"jekyll", "jekyll-sass-converter"},
		repository{"jekyll", "jemoji"},
		repository{"jekyll", "jekyll-gist"},
		repository{"jekyll", "jekyll-coffeescript"},
		repository{"jekyll", "plugins"},
	}

	sleepBetweenFreezes = 150 * time.Millisecond
)

func main() {
	var actuallyDoIt bool
	flag.BoolVar(&actuallyDoIt, "f", false, "Whether to actually mark the issues or close them.")
	var inputRepos string
	flag.StringVar(&inputRepos, "repos", "", "Specify a list of comma-separated repo name/owner pairs, e.g. 'jekyll/jekyll-admin'.")
	flag.Parse()

	var repos []repository
	if inputRepos == "" {
		repos = defaultRepos
	}

	context := ctx.NewDefaultContext()
	if context.GitHub == nil {
		log.Fatalln("cannot proceed without github client")
	}

	// Support running on just a list of issues. Either a URL or a `owner/name#number` syntax.
	if flag.NArg() > 0 {
		if err := processSingleIssues(context, actuallyDoIt, flag.Args()...); err != nil {
			log.Fatalf("error: %#v", err)
		}
		return
	}

	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(context *ctx.Context, repo repository, actuallyDoIt bool) {
			defer wg.Done()
			if err := processRepo(context, repo.Owner, repo.Name, actuallyDoIt); err != nil {
				log.Printf("%s/%s: error: %#v", repo.Owner, repo.Name, err)
			}
		}(context, repo, actuallyDoIt)
	}

	wg.Wait()
}

func extractIssueInfo(issueName string) (owner, repo string, number int) {
	issueName = strings.TrimPrefix(issueName, "https://github.com/")

	var err error
	pieces := strings.Split(issueName, "/")

	// Ex: `owner/repo#number`
	if len(pieces) == 2 {
		owner = pieces[0]
		morePieces := strings.Split(pieces[1], "#")
		if len(morePieces) == 2 {
			repo = morePieces[0]
			number, err = strconv.Atoi(morePieces[1])
			if err != nil {
				log.Printf("huh? %#v for %s", err, morePieces[1])
			}
		}
		return
	}

	// Ex: `owner/repo/issues/number`
	if len(pieces) == 4 {
		owner = pieces[0]
		repo = pieces[1]
		number, err = strconv.Atoi(pieces[3])
		if err != nil {
			log.Printf("huh? %#v for %s", err, pieces[3])
		}
		return
	}

	return "", "", 0
}

func processSingleIssues(context *ctx.Context, actuallyDoIt bool, issueNames ...string) error {
	issues := []github.Issue{}
	for _, issueName := range issueNames {
		owner, repo, number := extractIssueInfo(issueName)
		if owner == "" || repo == "" || number <= 0 {
			return fmt.Errorf("couldn't extract issue info from '%s': owner=%s repo=%s number=%d",
				issueName, owner, repo, number)
		}

		issues = append(issues, github.Issue{
			Number: github.Int(number),
			Repository: &github.Repository{
				Owner: &github.User{Login: github.String(owner)},
				Name:  github.String(repo),
			},
		})
	}
	return processIssues(context, actuallyDoIt, issues)
}

func processRepo(context *ctx.Context, owner, repo string, actuallyDoIt bool) error {
	start := time.Now()

	issues, err := freeze.AllTooOldIssues(context, owner, repo)
	if err != nil {
		return err
	}

	log.Printf("%s/%s: freezing %d closed issues before %v", owner, repo, len(issues), freeze.TooOld)
	err = processIssues(context, actuallyDoIt, issues)
	log.Printf("%s/%s: finished in %s", owner, repo, time.Since(start))

	return err
}

func processIssues(context *ctx.Context, actuallyDoIt bool, issues []github.Issue) error {
	for _, issue := range issues {
		owner, repo := *issue.Repository.Owner.Login, *issue.Repository.Name
		if actuallyDoIt {
			log.Printf("%s/%s: freezing %s", owner, repo, *issue.HTMLURL)
			if err := freeze.Freeze(context, owner, repo, *issue.Number); err != nil {
				return err
			}
			time.Sleep(sleepBetweenFreezes)
		} else {
			log.Printf("%s/%s: would have frozen %s", owner, repo, *issue.HTMLURL)
			time.Sleep(sleepBetweenFreezes)
		}
	}
	return nil
}
