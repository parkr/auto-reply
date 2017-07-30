package triage

import (
	"context"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/google/go-github/github"
)

func New(client *github.Client, labelsofInterest []string) *Triager {
	return &Triager{
		Client:           client,
		LabelsOfInterest: labelsofInterest,
		repoTypeCache:    map[string][]IssueGrouping{},
	}
}

type Triager struct {
	Client           *github.Client
	LabelsOfInterest []string

	repoTypeCache map[string][]IssueGrouping
}

func (t *Triager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("reset") != "" {
		t.repoTypeCache = map[string][]IssueGrouping{}
		log.Println("resetting repoTypeCache...")
		http.Redirect(w, r, "/triage", 302)
		return
	}

	repo := r.FormValue("repo")
	if repo == "" {
		repo = "jekyll/jekyll"
	}

	issueType := r.FormValue("type")
	if issueType == "" {
		issueType = "all"
	}

	label := r.FormValue("label")
	// If blank, then we pull all of them.

	order := r.FormValue("order")
	// If blank, then we don't do any sorting.

	err := triageTmpl.Execute(w, t.getTemplateInfo(repo, issueType, label, order))
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func (t *Triager) getTemplateInfo(repo, issueType, label, order string) templateInfo {
	key := repo + "____" + issueType
	if _, ok := t.repoTypeCache[key]; !ok {
		t.repoTypeCache[key] = t.fetchIssues(repo, issueType)
	}

	desiredGroupings := []IssueGrouping{}
	if label == "" {
		desiredGroupings = t.repoTypeCache[key]
	} else {
		for _, labelGrouping := range t.repoTypeCache[key] {
			if labelGrouping.Label == label {
				desiredGroupings = append(desiredGroupings, labelGrouping)
				break
			}
		}
	}

	if issueType == "all" {
		issueType = "issues & pull request"
	}

	if order != "" {
		log.Printf("Ordering issues by CreatedAt %s", order)
		for _, labelGrouping := range desiredGroupings {
			if order == "desc" {
				// Sorts them in descending order by CreatedAt date.
				sort.Stable(sort.Reverse(labelGrouping.Issues))
			} else {
				// Sorts them in ascending order by CreatedAt date.
				sort.Stable(labelGrouping.Issues)
			}
		}
	}

	return templateInfo{
		RepoName:             repo,
		IssueType:            issueType,
		IssuesGroupedByLabel: desiredGroupings,
	}
}

func (t Triager) fetchIssues(repo, issueType string) []IssueGrouping {
	log.Printf("Fetching issues of type %s for %s", issueType, repo)

	// Get all issues or pull requests (depending on issueType) for repo
	issues := []github.Issue{}
	query := "repo:" + repo + " is:open"
	if issueType != "" && issueType != "all" {
		query += " is:" + issueType
	}
	opts := &github.SearchOptions{
		Sort:        "created",
		Order:       "asc",
		ListOptions: github.ListOptions{PerPage: 500},
	}
	for {
		log.Printf("Running query %q page %d", query, opts.ListOptions.Page)
		result, resp, err := t.Client.Search.Issues(context.Background(), query, opts)
		if err != nil {
			return []IssueGrouping{{Label: "error: " + err.Error()}}
		}

		for _, issue := range result.Issues {
			issues = append(issues, issue)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	// Create groupings
	triageGroup := &IssueGrouping{Label: "triage", Issues: Issues{}}
	grouping := []*IssueGrouping{}
	for _, label := range t.LabelsOfInterest {
		grouping = append(grouping, &IssueGrouping{Label: label, Issues: Issues{}})
	}

	// Group by each label of interest.
	for _, issue := range issues {
		matchedALabelGroup := false
		for _, labelGroup := range grouping {
			for _, label := range issue.Labels {
				if label.GetName() == labelGroup.Label {
					labelGroup.Issues = append(labelGroup.Issues, issue)
					matchedALabelGroup = true
					break
				}
			}
		}
		// If it didn't match another grouping label, then it's not been properly triaged.
		if matchedALabelGroup == false {
			triageGroup.Issues = append(triageGroup.Issues, issue)
		}
	}

	triageGroup.lastUpdated = time.Now()
	unmodifiable := []IssueGrouping{*triageGroup}
	for _, group := range grouping {
		group.lastUpdated = time.Now()
		unmodifiable = append(unmodifiable, *group)
	}

	for _, labelGroup := range unmodifiable {
		log.Printf("Label group %q has %d issues", labelGroup.Label, len(labelGroup.Issues))
	}

	log.Printf("Done fetching issues of type %s for %s... found %d issues", issueType, repo, len(issues))

	return unmodifiable
}
