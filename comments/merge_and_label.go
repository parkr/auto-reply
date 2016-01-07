package comments

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/parkr/auto-reply/Godeps/_workspace/src/github.com/google/go-github/github"
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
		// commitMsg := fmt.Sprintf("Merge pull request %v", number)
		// _, _, mergeErr := client.PullRequests.Merge(owner, repo, number, commitMsg)
		// if err != nil {
		//     fmt.Printf("comments: error merging %v\n", err)
		//     return err
		// }

		// Delete branch
		//ref := fmt.Sprintf("heads/%s", branch)
		//res, deleteBranchErr := client.Git.DeleteRef(owner, repo, ref)

		// Read History.markdown, add line to appropriate change section
		historyFileContents := getHistoryContents(client, owner, repo)
		log.Println(historyFileContents)

		// Commit change to History.markdown

		return nil
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

	if strings.HasPrefix(label, "development") {
		return "development-fixes"
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

func getHistoryContents(client *github.Client, owner, repo string) string {
	content, _, _, err := client.Repositories.GetContents(owner, repo, "History.markdown", nil)
	if err != nil {
		fmt.Printf("comments: error getting History.markdown %v\n", err)
		return ""
	}
	return *content.Content
}
