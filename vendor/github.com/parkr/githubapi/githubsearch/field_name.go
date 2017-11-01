package githubsearch

type FieldName int64

const (
	_     FieldName = iota
	Label FieldName = 1 << (10 * iota)
	Milestone
	Assignee
	Project
)

func (f FieldName) String() string {
	switch f {
	case Label:
		return "label"
	case Milestone:
		return "milestone"
	case Assignee:
		return "assignee"
	case Project:
		return "project"
	default:
		return ""
	}
}
