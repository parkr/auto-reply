// check-for-outdated-dependencies takes a repo
package main

import (
	"flag"
	"log"
	"strings"

	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/dependencies"
)

var defaultRepos = strings.Join([]string{
	"jekyll/jekyll",
}, ",")

func main() {
	var depType string
	flag.StringVar(&depType, "type", "ruby", "The type of dependency we're checking (options: ruby)")
	var reposString string
	flag.StringVar(&reposString, "repos", defaultRepos, "Comma-separated list of repos to check, e.g. jekyll/jekyll,jekyll/jekyll-import")
	flag.Parse()
	context := ctx.NewDefaultContext()

	for _, repo := range strings.Split(reposString, ",") {
		pieces := strings.SplitN(repo, "/", 2)
		checker := dependencies.NewRubyDependencyChecker(pieces[0], pieces[1])
		outdated := checker.AllOutdatedDependencies(context)
		for _, dependency := range outdated {
			log.Printf(
				"%s/%s: %s is outdated (constraint: %s, but latest version is %s)",
				pieces[0], pieces[1], dependency.GetName(), dependency.GetConstraint(), dependency.GetLatestVersion(context),
			)
		}
	}
}
