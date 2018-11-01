package githubsearch

import "strings"

// All parameters for searching issues and pull requests.
// https://help.github.com/articles/searching-issues-and-pull-requests/
type IssueSearchParameters struct {
	// The issue type, used to scope to only issues or PR's. Blank by
	// default.
	Type IssueType

	// Which fields to search in, i.e. 'title,' 'body,' 'comments'.
	Scope ScopeLevel

	// Visibility of the repository, i.e. 'public' or 'private'.
	Visibility Visibility

	// The author of the issue, e.g. 'defunkt'.
	Author string

	// Assignees of interest, e.g. 'defunkt'.
	Assignees []string

	// Search for mentions of a user, e.g. 'defunkt'.
	Mentions string

	// Issues with comments from a user, e.g. 'defunkt'.
	Commenter string

	// A logical OR of author, assignees, mentions, commenter.
	// Searches for a user who is the author or assignee of an issue, or is
	// mentioned in or has commented on the issue. Search 4 fields at once!
	Involves string

	// Team mentions, e.g. 'bootstrap/maintainers'.
	Team string

	// State of an issue/PR, i.e. whether it's open, closed, merged, or unmerged.
	State State

	// Issue must have these labels (a logical AND on the contents of the slice).
	Labels []string

	// Milestone the issue belongs to, e.g. 'bug fix'.
	Milestone string

	// Project board to search in form 'user/repo/project', e.g.
	// 'github/linguist/1' is project board 1 in https://github.com/github/linguist.
	ProjectBoard string

	// Issues which have blank fields, e.g. 'no:label', 'no:assignee', etc
	MissingField FieldName

	// Language of the repository, e.g. 'ruby', 'python', 'go'.
	Language string

	// CI status of the latest commit, e.g. 'pending', 'success', 'failure'.
	Status CommitStatus

	// Head branch name, e.g. 'head:patch-1'
	HeadBranchName string

	// Base branch name (usually 'master'), e.g. 'base:master'
	BaseBranchName string

	// When the issue was created, allows modifiers.
	CreatedAt *TimeParameters

	// When the issue was updated, allows modifiers.
	UpdatedAt *TimeParameters

	// When it was merged, allows modifiers.
	MergedAt *TimeParameters

	// When it was closed, allows modifiers.
	ClosedAt *TimeParameters

	// How many comments are on the issue, allows modifiers.
	NumComments *NumericalParameters

	// User account who owns the repository/repositories, e.g. 'defunkt'.
	User string

	// Organization account who owns the repository/repositories, e.g. 'github'.
	Organization string

	// Specific repository to search.
	Repository *RepositoryName

	// Reviewed status
	Reviewed ReviewStatus

	// User who reviewed the PR, e.g. 'defunkt'
	ReviewedBy string

	// User whose review was requested, e.g. 'defunkt'
	ReviewRequested string

	// Team whose review was requested, e.g. 'jekyll/core'.
	TeamReviewRequested string

	// A specific piece of text to search for, e.g. 'cat'.
	Query string
}

// QueryParts returns a slice of strings which can be input into the GitHub
// Search API directly, e.g. ['user:defunkt', 'comments:>10']
func (p IssueSearchParameters) QueryParts() []string {
	var parts []string

	if p.Type.String() != "" {
		parts = append(parts, p.Type.String())
	}

	if p.Scope.String() != "" {
		parts = append(parts, p.Scope.String())
	}

	if p.Visibility.String() != "" {
		parts = append(parts, p.Visibility.String())
	}

	if p.Author != "" {
		parts = append(parts, "author:"+p.Author)
	}

	if len(p.Assignees) != 0 {
		for _, assignee := range p.Assignees {
			parts = append(parts, "assignee:"+assignee)
		}
	}

	if p.Mentions != "" {
		parts = append(parts, "mentions:"+p.Mentions)
	}

	if p.Commenter != "" {
		parts = append(parts, "commenter:"+p.Commenter)
	}

	if p.Involves != "" {
		parts = append(parts, "involves:"+p.Involves)
	}

	if p.Team != "" {
		parts = append(parts, "team:"+p.Team)
	}

	if p.State.String() != "" {
		parts = append(parts, p.State.String())
	}

	if len(p.Labels) != 0 {
		for _, label := range p.Labels {
			parts = append(parts, "label:"+quoteMultiWordStrings(label))
		}
	}

	if p.Milestone != "" {
		parts = append(parts, "milestone:"+quoteMultiWordStrings(p.Milestone))
	}

	if p.ProjectBoard != "" {
		parts = append(parts, "project:"+p.ProjectBoard)
	}

	if p.MissingField.String() != "" {
		parts = append(parts, "no:"+p.MissingField.String())
	}

	if p.Language != "" {
		parts = append(parts, "language:"+p.Language)
	}

	if p.Status.String() != "" {
		parts = append(parts, "status:"+p.Status.String())
	}

	if p.HeadBranchName != "" {
		parts = append(parts, "head:"+quoteMultiWordStrings(p.HeadBranchName))
	}

	if p.BaseBranchName != "" {
		parts = append(parts, "base:"+quoteMultiWordStrings(p.BaseBranchName))
	}

	if p.CreatedAt != nil {
		parts = append(parts, "created:"+p.CreatedAt.String())
	}

	if p.UpdatedAt != nil {
		parts = append(parts, "updated:"+p.UpdatedAt.String())
	}

	if p.MergedAt != nil {
		parts = append(parts, "merged:"+p.MergedAt.String())
	}

	if p.ClosedAt != nil {
		parts = append(parts, "closed:"+p.ClosedAt.String())
	}

	if p.NumComments != nil {
		parts = append(parts, "comments:"+p.NumComments.String())
	}

	if p.User != "" {
		parts = append(parts, "user:"+p.User)
	}

	if p.Organization != "" {
		parts = append(parts, "org:"+p.Organization)
	}

	if p.Repository != nil {
		parts = append(parts, "repo:"+p.Repository.String())
	}

	if p.Reviewed.String() != "" {
		parts = append(parts, p.Reviewed.String())
	}

	if p.ReviewedBy != "" {
		parts = append(parts, "reviewed-by:"+p.ReviewedBy)
	}

	if p.ReviewRequested != "" {
		parts = append(parts, "review-requested:"+p.ReviewRequested)
	}

	if p.TeamReviewRequested != "" {
		parts = append(parts, "team-review-requested:"+p.TeamReviewRequested)
	}

	if p.Query != "" {
		parts = append(parts, p.Query)
	}

	return parts
}

func (p IssueSearchParameters) String() string {
	return strings.Join(p.QueryParts(), " ")
}

func quoteMultiWordStrings(input string) string {
	if strings.ContainsRune(input, ' ') {
		return "\"" + input + "\""
	}
	return input
}
