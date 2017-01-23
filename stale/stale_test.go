package stale

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestIsUpdatedWithinDuration(t *testing.T) {
	twoMonthsAgo := time.Now().AddDate(0, -2, 0)
	dormantDuration := time.Since(twoMonthsAgo)
	config := Configuration{DormantDuration: dormantDuration}

	cases := []struct {
		updatedAtDate                      time.Time
		isUpdatedWithinDurationReturnValue bool
	}{
		{time.Now().AddDate(-1, 0, 0), false},
		{time.Now().AddDate(0, -2, -1), false},
		{time.Now().AddDate(0, -1, -30), true},
		{time.Now().AddDate(0, -1, -29), true},
		{time.Now(), true},
	}

	for _, testCase := range cases {
		issue := &github.Issue{UpdatedAt: &testCase.updatedAtDate}
		assert.Equal(t,
			testCase.isUpdatedWithinDurationReturnValue,
			isUpdatedWithinDuration(issue, config),
			fmt.Sprintf(
				"date='%s' config.DormantDuration='%s' time.Since(date)='%s'",
				testCase.updatedAtDate,
				config.DormantDuration,
				time.Since(testCase.updatedAtDate)),
		)
	}
}
