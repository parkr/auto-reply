package comments

import (
	"errors"
	"fmt"
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
			labels, _, err := client.Issues.ListLabelsForMilestone(
				*event.Repo.Owner.Login,
				*event.Repo.Name,
				*event.Issue.Number,
				nil,
			)
			if err != nil {
				return err
			}
			fmt.Printf("labels from GitHub = %v\n", labels)
			changeSectionLabel = sectionForLabel(selectSectionLabel(labels))
		}
		fmt.Printf("changeSectionLabel = '%s'\n", changeSectionLabel)

		// What is the label for the change section? E.g. "Major Enhancement," "Minor Enhancement," etc

		// Merge
		// Read History.markdown, add line to appropriate change section
		// Commit change to History.markdown

		return nil
	}
)

func isAuthorizedCommenter(user *github.User) bool {
	return *user.Login == "parkr"
}

func parseMergeRequestComment(commentBody string) (bool, string) {
	matches := mergeCommentRegexp.FindAllStringSubmatch(commentBody, -1)
	if matches == nil {
		return false, ""
	}

	var label string
	if labelFromComment := matches[0][3]; labelFromComment != "" {
		label = downcaseAndHyphenize(labelFromComment)
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
