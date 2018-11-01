package githubsearch

type ReviewStatus int64

const (
	_ ReviewStatus = iota

	// No review.
	NoReview ReviewStatus = 1 << (10 * iota)
	// Reviews are required.
	RequiredReview
	// At least 1 approved review
	ApprovedReview
	// At least 1 changes requested review
	ChangesRequestedReview
)

func (r ReviewStatus) String() string {
	switch r {
	case NoReview:
		return "review:none"
	case RequiredReview:
		return "review:required"
	case ApprovedReview:
		return "review:approved"
	case ChangesRequestedReview:
		return "review:changes_requested"
	default:
		return ""
	}
}
