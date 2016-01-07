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
			changeSectionLabel = labelFromComment
		} else {
			// Get changeSectionLabel from issue labels!
			changeSectionLabel = ""
		}
		fmt.Println(changeSectionLabel)

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

func parseMergeRequestComment(commentBody string) (isRequest bool, label string) {
	matches := mergeCommentRegexp.FindAllStringSubmatch(commentBody, -1)
	if matches == nil {
		return
	}

	isRequest = true

	if labelFromComment := matches[0][3]; labelFromComment != "" {
		label = downcaseAndHyphenize(labelFromComment)
	}

	return
}

func downcaseAndHyphenize(label string) string {
	return strings.Replace(strings.ToLower(label), " ", "-", -1)
}

func containsChangeLabel(commentBody string) bool {
	_, labelFromComment := parseMergeRequestComment(commentBody)
	return labelFromComment != ""
}
