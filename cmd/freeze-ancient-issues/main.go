// A command-line utility to lock old issues.
package main

import (
	"flag"
	"log"
	"sync"
	"time"

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

func processRepo(context *ctx.Context, owner, repo string, actuallyDoIt bool) error {
	start := time.Now()

	issues, err := freeze.AllTooOldIssues(context, owner, repo)
	if err != nil {
		return err
	}

	log.Printf("%s/%s: Freezing %d closed issues before %v", owner, repo, len(issues), freeze.TooOld)
	for _, issue := range issues {
		if actuallyDoIt {
			log.Printf("%s/%s: freezing %s", owner, repo, *issue.HTMLURL)
			if err = freeze.Freeze(context, owner, repo, *issue.Number); err != nil {
				return err
			}
			time.Sleep(sleepBetweenFreezes)
		} else {
			log.Printf("%s/%s: would have frozen %s", owner, repo, *issue.HTMLURL)
			time.Sleep(sleepBetweenFreezes)
		}
	}

	log.Printf("%s/%s: finished in %s", owner, repo, time.Since(start))

	return nil
}
