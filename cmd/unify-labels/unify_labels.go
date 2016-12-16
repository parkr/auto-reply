// unify-labels is a CLI which will add, rename, or change the color of labels so they match the Jekyll org's requirements.
package main

import (
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var desiredLabels = []*github.Label{
	{Name: github.String("bug"), Color: github.String("d41313")},
	{Name: github.String("discussion"), Color: github.String("006b75")},
	{Name: github.String("documentation"), Color: github.String("006b75")},
	{Name: github.String("enhancement"), Color: github.String("009800")},
	{Name: github.String("feature"), Color: github.String("009800")},
	{Name: github.String("fix"), Color: github.String("eb6420")},
	{Name: github.String("github"), Color: github.String("222222")},
	{Name: github.String("has-pull-request"), Color: github.String("fbca04")},
	{Name: github.String("help-wanted"), Color: github.String("fbca04")},
	{Name: github.String("internal"), Color: github.String("ededed")},
	{Name: github.String("pending-feedback"), Color: github.String("fbca04")},
	{Name: github.String("pending-rebase"), Color: github.String("eb6420")},
	{Name: github.String("release"), Color: github.String("d4c5f9")},
	{Name: github.String("security"), Color: github.String("e11d21")},
	{Name: github.String("stale"), Color: github.String("bfd4f2")},
	{Name: github.String("suggestion"), Color: github.String("0052cc")},
	{Name: github.String("support"), Color: github.String("5319e7")},
	{Name: github.String("tests"), Color: github.String("d4c5f9")},
	{Name: github.String("undetermined"), Color: github.String("fe3868")},
	{Name: github.String("ux"), Color: github.String("006b75")},
	{Name: github.String("windows"), Color: github.String("fbca04")},
	{Name: github.String("wont-fix"), Color: github.String("e11d21")},
}
var listOpts = github.ListOptions{PerPage: 100}

func allPossibleNames(name string) []string {
	return []string{
		name,
		strings.Replace(name, "-", "", -1),
		strings.Replace(name, "-", " ", -1),
		strings.ToLower(name),
		strings.Title(name),
		strings.Title(strings.Replace(name, "-", "", -1)),
		strings.Title(strings.Replace(name, "-", " ", -1)),
	}
}

func findLabel(labels []*github.Label, desiredLabel *github.Label) *github.Label {
	possibleNames := allPossibleNames(*desiredLabel.Name)
	for _, possibleName := range possibleNames {
		for _, label := range labels {
			if *label.Name == possibleName {
				return label
			}
		}
	}

	return nil
}

func main() {
	context := ctx.NewDefaultContext()
	repos, _, err := context.GitHub.Repositories.List("jekyll", &github.RepositoryListOptions{
		Type: "owner", Sort: "full_name", Direction: "asc", ListOptions: listOpts,
	})
	if err != nil {
		log.Fatalln("error fetching repos:", err)
	}

	for _, repo := range repos {
		owner, repoName := *repo.Owner.Login, *repo.Name
		log.Printf("Processing %s", *repo.FullName)
		labels, _, err := context.GitHub.Issues.ListLabels(owner, repoName, &listOpts)
		if err != nil {
			log.Fatalf("error fetching labels for %s: %v", *repo.FullName, err)
		}

		for _, desiredLabel := range desiredLabels {
			matchedLabel := findLabel(labels, desiredLabel)

			if matchedLabel == nil {
				log.Printf("%s: creating %s with color %s",
					*repo.FullName, *desiredLabel.Name, *desiredLabel.Color)
				// It doesn't exist. Create and continue.
				_, _, err := context.GitHub.Issues.CreateLabel(owner, repoName, desiredLabel)
				if err != nil {
					log.Fatalf("error creating '%s' for %s: %v", *desiredLabel.Name, *repo.FullName, err)
				}
				continue
			}

			// It does exist, but possibly with incorrect info. Update it.
			if *matchedLabel.Name != *desiredLabel.Name || *matchedLabel.Color != *desiredLabel.Color {
				log.Printf("%s: updating %s with data: %v",
					*repo.FullName, *matchedLabel.Name, github.Stringify(desiredLabel))
				_, _, err := context.GitHub.Issues.EditLabel(owner, repoName, *matchedLabel.Name, desiredLabel)
				if err != nil {
					log.Fatalf("error updating '%s' for %s: %v", *matchedLabel.Name, *repo.FullName, err)
				}
				continue
			}
		}
	}
}
