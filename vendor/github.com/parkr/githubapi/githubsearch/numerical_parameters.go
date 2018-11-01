package githubsearch

import "strconv"

type NumericalParameters struct {
	Count    int
	Modifier Modifier

	// Only used if Modifier is Range
	UpperBound int
}

func (n NumericalParameters) String() string {
	switch n.Modifier {
	case Range:
		return strconv.Itoa(n.Count) + n.Modifier.String() + strconv.Itoa(n.UpperBound)
	case Equal, LessThan, LessThanOrEqualTo, GreaterThan, GreaterThanOrEqualTo:
		return n.Modifier.String() + strconv.Itoa(n.Count)
	default:
		panic("Not sure what to do with modifier: " + n.Modifier.String())
	}
}
