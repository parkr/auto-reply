package githubsearch

type Modifier int64

const (
	Equal    Modifier = iota
	LessThan          = 1 << (10 * iota)
	LessThanOrEqualTo
	GreaterThan
	GreaterThanOrEqualTo
	Range
)

func (m Modifier) String() string {
	switch m {
	case Equal:
		return ""
	case LessThan:
		return "<"
	case LessThanOrEqualTo:
		return "<="
	case GreaterThan:
		return ">"
	case GreaterThanOrEqualTo:
		return ">="
	case Range:
		return ".."
	default:
		return ""
	}
}
