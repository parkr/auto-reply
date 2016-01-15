package comments

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/parkr/auto-reply/Godeps/_workspace/src/github.com/google/go-github/github"
	"github.com/parkr/auto-reply/Godeps/_workspace/src/github.com/parkr/changelog"
)

var (
	mergeCommentRegexp = regexp.MustCompile("@[a-zA-Z-_]+: (merge|:shipit:|:ship:)( \\+([a-zA-Z-_ ]+))?")

	HandlerMergeAndLabel = func(client *github.Client, event github.IssueCommentEvent) error {
		// Is this a pull request?
		if !isPullRequest(event) {
			return errors.New("not a pull request")
		}

		var changeSectionLabel string
		isReq, labelFromComment := parseMergeRequestComment(*event.Comment.Body)

		// Is It a merge request comment?
		if !isReq {
			return errors.New("not a merge request comment")
		}

		log.Println(event)

		owner, repo, number := *event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number

		// Does the user have merge/label abilities?
		if !isAuthorizedCommenter(event.Comment.User) {
			return errors.New("commenter isn't allowed to merge")
		}

		// Should it be labeled?
		if labelFromComment != "" {
			// Apply label
			changeSectionLabel = sectionForLabel(labelFromComment)
		} else {
			// Get changeSectionLabel from issue labels!
			labels, _, err := client.Issues.ListLabelsForMilestone(owner, repo, number, nil)
			if err != nil {
				return err
			}
			fmt.Printf("labels from GitHub = %v\n", labels)
			changeSectionLabel = sectionForLabel(selectSectionLabel(labels))
		}
		fmt.Printf("changeSectionLabel = '%s'\n", changeSectionLabel)

		// Merge
		commitMsg := fmt.Sprintf("Merge pull request %v", number)
		_, _, mergeErr := client.PullRequests.Merge(owner, repo, number, commitMsg)
		if mergeErr != nil {
			fmt.Printf("comments: error merging %v\n", mergeErr)
			return mergeErr
		}

		// Delete branch
		repoInfo, _, getRepoErr := client.PullRequests.Get(owner, repo, number)
		if getRepoErr != nil {
			fmt.Printf("comments: error fetching pull request: %v\n", getRepoErr)
			return getRepoErr
		}

		// Delete branch
		if deletableRef(repoInfo, owner) {
			ref := fmt.Sprintf("heads/%s", *repoInfo.Head.Ref)
			_, deleteBranchErr := client.Git.DeleteRef(owner, repo, ref)
			if deleteBranchErr != nil {
				fmt.Printf("comments: error deleting branch %v\n", mergeErr)
			}
		}

		// Read History.markdown, add line to appropriate change section
		historyFileContents, historySHA := getHistoryContents(client, owner, repo)

		// Add to
		newHistoryFileContents := addMergeReference(historyFileContents, changeSectionLabel, *repoInfo.Title, number)

		// Commit change to History.markdown
		commitErr := commitHistoryFile(client, historySHA, owner, repo, number, newHistoryFileContents)
		if commitErr != nil {
			fmt.Printf("comments: error committing updated history %v\n", mergeErr)
		}
		return commitErr
	}
)

func isAuthorizedCommenter(user *github.User) bool {
	return *user.Login == "parkr"
}

func parseMergeRequestComment(commentBody string) (bool, string) {
	matches := mergeCommentRegexp.FindAllStringSubmatch(commentBody, -1)
	if matches == nil || matches[0] == nil {
		return false, ""
	}

	var label string
	if len(matches[0]) >= 4 {
		if labelFromComment := matches[0][3]; labelFromComment != "" {
			label = downcaseAndHyphenize(labelFromComment)
		}
	}

	return true, normalizeLabel(label)
}

func downcaseAndHyphenize(label string) string {
	return strings.Replace(strings.ToLower(label), " ", "-", -1)
}

func normalizeLabel(label string) string {
	if strings.HasPrefix(label, "major") {
		return "major-enhancements"
	}

	if strings.HasPrefix(label, "minor") {
		return "minor-enhancements"
	}

	if strings.HasPrefix(label, "bug") {
		return "bug-fixes"
	}

	if strings.HasPrefix(label, "dev") {
		return "development-fixes"
	}

	if strings.HasPrefix(label, "site") {
		return "site-enhancements"
	}

	return label
}

func sectionForLabel(label string) string {
	switch label {
	case "major-enhancements":
		return "Major Enhancements"
	case "minor-enhancements":
		return "Minor Enhancements"
	case "bug-fixes":
		return "Bug Fixes"
	case "development-fixes":
		return "Development Fixes"
	case "site-enhancements":
		return "Site Enhancements"
	default:
		return label
	}
}

func selectSectionLabel(labels []github.Label) string {
	for _, label := range labels {
		if sectionForLabel(*label.Name) != *label.Name {
			return *label.Name
		}
	}
	return ""
}

func containsChangeLabel(commentBody string) bool {
	_, labelFromComment := parseMergeRequestComment(commentBody)
	return labelFromComment != ""
}

func getHistoryContents(client *github.Client, owner, repo string) (content, sha string) {
	contents, _, _, err := client.Repositories.GetContents(
		owner,
		repo,
		"History.markdown",
		&github.RepositoryContentGetOptions{Ref: "heads/master"},
	)
	if err != nil {
		fmt.Printf("comments: error getting History.markdown %v\n", err)
		return "", ""
	}
	return base64Decode(*contents.Content), *contents.SHA
}

func base64Decode(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Printf("comments: error decoding string: %v\n", err)
		return ""
	}
	return string(decoded)
}

func addMergeReference(historyFileContents, changeSectionLabel, prTitle string, number int) string {
	changes, err := changelog.NewChangelogFromReader(strings.NewReader(historyFileContents))
	if historyFileContents == "" {
		err = nil
		changes = &changelog.Changelog{
			Versions: []*changelog.Version{},
		}
	}
	if err != nil {
		fmt.Printf("comments: error %v\n", err)
		return historyFileContents
	}
	reference := fmt.Sprintf("#%d", number)

	// Find HEAD, or create
	var version *changelog.Version
	for _, v := range changes.Versions {
		if v.Version == "HEAD" {
			version = v
		}
	}
	if version == nil {
		version = &changelog.Version{
			Version:     "HEAD",
			Subsections: []*changelog.Subsection{},
		}
		changes.Versions = append([]*changelog.Version{version}, changes.Versions...)
	}

	// Find Subsection, or create
	var subsection *changelog.Subsection
	for _, s := range version.Subsections {
		if s.Name == changeSectionLabel {
			subsection = s
		}
	}
	if subsection == nil {
		subsection = &changelog.Subsection{
			Name:    changeSectionLabel,
			History: []*changelog.ChangeLine{},
		}
		version.Subsections = append([]*changelog.Subsection{subsection}, version.Subsections...)
	}

	// Find changeline, only create if does not exist.
	for _, c := range subsection.History {
		if c.Reference == reference {
			return historyFileContents
		}
	}
	changeLine := &changelog.ChangeLine{
		Summary:   prTitle,
		Reference: reference,
	}
	subsection.History = append(subsection.History, changeLine)

	return changes.String()
}

func deletableRef(pr *github.PullRequest, owner string) bool {
	return *pr.Head.Repo.Owner.Login == owner && *pr.Head.Ref != "master" && *pr.Head.Ref != "gh-pages"
}

func commitHistoryFile(client *github.Client, historySHA, owner, repo string, number int, newHistoryFileContents string) error {
	repositoryContentsOptions := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Update history to reflect merge of #%d [ci skip]", number)),
		Content: []byte(newHistoryFileContents),
		SHA:     github.String(historySHA),
		Committer: &github.CommitAuthor{
			Name:  github.String("jekyllbot"),
			Email: github.String("jekyllbot@jekyllrb.com"),
		},
	}
	updateResponse, _, err := client.Repositories.UpdateFile(owner, repo, "History.markdown", repositoryContentsOptions)
	if err != nil {
		fmt.Printf("comments: error committing History.markdown: %v\n", err)
		return err
	}
	fmt.Printf("comments: updateResponse: %s\n", updateResponse)
	return nil
}
