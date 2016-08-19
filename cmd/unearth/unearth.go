// A command-line utility to run a search query and display results.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var (
	context *ctx.Context

	defaultListOptions = &github.ListOptions{Page: 0, PerPage: 200}
)

func haltIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func repoNameFromURL(url string) string {
	return strings.Join(
		strings.SplitN(
			strings.Replace(url, "https://github.com/", "", 1),
			"/",
			-1)[1:2],
		"/",
	)
}

func issuesForQuery(query string) {
	result, _, err := context.GitHub.Search.Issues(query, &github.SearchOptions{
		Sort:        "created",
		ListOptions: *defaultListOptions,
	})
	haltIfError(err)
	fmt.Printf("Query '%s' found %d issues:\n", query, *result.Total)
	for _, issue := range result.Issues {
		fmt.Printf("%-20s %-4d %s | %s\n",
			repoNameFromURL(*issue.HTMLURL),
			*issue.Number,
			issue.CreatedAt.Format("2006-01-02"),
			*issue.Title,
		)
	}
}

func main() {
	flag.Parse()
	context = ctx.NewDefaultContext()

	if flag.NArg() < 1 {
		// Default queries.
		issuesForQuery("state:open comments:0 user:jekyll")
		fmt.Print("\n\n")
		issuesForQuery("state:open comments:>10 user:jekyll")
	} else {
		// User-specified query.
		issuesForQuery(strings.Join(flag.Args(), " "))
	}
}
