package githubsearch

type CommitStatus int64

const (
	_       CommitStatus = iota
	Pending CommitStatus = 1 << (10 * iota)
	Failure
	Success
)

func (c CommitStatus) String() string {
	switch c {
	case Pending:
		return "pending"
	case Failure:
		return "failure"
	case Success:
		return "success"
	default:
		return ""
	}
}
