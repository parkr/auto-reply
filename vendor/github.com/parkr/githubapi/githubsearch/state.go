package githubsearch

type State int64

const (
	_    State = iota
	Open State = 1 << (10 * iota)
	Closed
	Merged
	Unmerged
)

func (s State) String() string {
	switch s {
	case Open:
		return "is:open"
	case Closed:
		return "is:closed"
	case Merged:
		return "is:merged"
	case Unmerged:
		return "is:unmerged"
	default:
		return ""
	}
}
