package githubsearch

type Visibility int64

const (
	_      Visibility = iota
	Public Visibility = 1 << (10 * iota)
	Private
)

func (v Visibility) String() string {
	switch v {
	case Public:
		return "is:public"
	case Private:
		return "is:private"
	default:
		return ""
	}
}
