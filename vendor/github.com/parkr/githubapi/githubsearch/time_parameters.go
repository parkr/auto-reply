package githubsearch

import "time"

const searchTimeFormat = "2006-01-02T15:04:05-0700"

type TimeParameters struct {
	Time     time.Time
	Modifier Modifier

	// Only valid when using the Range modifier.
	EndTime time.Time
}

func (t TimeParameters) String() string {
	switch t.Modifier {
	case Range:
		if t.EndTime.IsZero() {
			panic("Range modifier must provide an EndTime")
		}
		return t.Time.Format(searchTimeFormat) + t.Modifier.String() + t.EndTime.Format(searchTimeFormat)
	case Equal, LessThan, LessThanOrEqualTo, GreaterThan, GreaterThanOrEqualTo:
		return t.Modifier.String() + t.Time.Format(searchTimeFormat)
	default:
		panic("Not sure what to do with modifier: " + t.Modifier.String())
	}
}
