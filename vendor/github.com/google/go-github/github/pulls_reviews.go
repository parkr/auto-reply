package github

import "time"

// No docs are out for this quite yet.
type PullRequestReview struct {
	ID          *int       `json:"id,omitempty"`
	User        *User      `json:"user,omitempty"`
	Body        *string    `json:"body,omitempty"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`

	// State can be "approved", "rejected", or "commented".
	State *string `json:"state,omitempty"`
}
