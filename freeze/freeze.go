package freeze

import (
	"time"
)

var (
	TooOld    = time.Now().Add(-365 * 24 * time.Hour).Format("2006-01-02")
	LabelName = "frozen-due-to-age"
)
