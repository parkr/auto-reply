package githubsearch

type ScopeLevel int64

const (
	_ ScopeLevel = iota

	// Limits the search scope to the title, i.e. `in:title`
	TitleScope ScopeLevel = 1 << (10 * iota)
	// Limits the search scope to the body, i.e. `in:body`
	BodyScope
	// Limits the search scope to the comments, i.e. `in:comments`
	CommentsScope
)

func (t ScopeLevel) String() string {
	switch t {
	case TitleScope:
		return "in:title"
	case BodyScope:
		return "in:body"
	case CommentsScope:
		return "in:comments"
	default:
		return ""
	}
}
