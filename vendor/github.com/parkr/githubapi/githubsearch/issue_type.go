package githubsearch

type IssueType int64

const (
	// Includes both issues and pull requests.
	All IssueType = iota
	// Limits the search scope to only issues, i.e. `is:issue`
	Issue IssueType = 1 << (10 * iota)
	// Limits the search scope to only pull requests, i.e. `is:pr`
	PullRequest
)

func (t IssueType) String() string {
	switch t {
	case All:
		return ""
	case Issue:
		return "is:issue"
	case PullRequest:
		return "is:pr"
	default:
		return ""
	}
}
