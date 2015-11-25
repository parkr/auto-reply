package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/parkr/auto-reply/Godeps/_workspace/src/github.com/google/go-github/github"
	"github.com/parkr/auto-reply/common"
)

var defaultListOptions = &github.ListOptions{Page: 0, PerPage: 200}

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

func main() {
	flag.Parse()
	client := common.NewClient()
	query := "state:open comments:0 user:jekyll"
	result, _, err := client.Search.Issues(query, &github.SearchOptions{
		Sort:        "created",
		ListOptions: *defaultListOptions,
	})
	haltIfError(err)
	log.Println(result)
	fmt.Printf("%d Issues:\n", *result.Total)
	for _, issue := range result.Issues {
		fmt.Printf("%-20s %-4d %s | %s\n",
			repoNameFromURL(*issue.HTMLURL),
			*issue.Number,
			issue.CreatedAt.Format("2006-01-02"),
			*issue.Title,
		)
	}
}
