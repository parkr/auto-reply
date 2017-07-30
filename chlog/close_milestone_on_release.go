package chlog

import (
	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

func CloseMilestoneOnRelease(context *ctx.Context, payload interface{}) error {
	release, ok := payload.(*github.ReleaseEvent)
	if !ok {
		return context.NewError("chlog.CloseMilestoneOnRelease: not a release event")
	}

	if *release.Action != "published" {
		return context.NewError("chlog.CloseMilestoneOnRelease: not a published release")
	}

	if *release.Release.Prerelease || *release.Release.Draft {
		return context.NewError("chlog.CloseMilestoneOnRelease: a prerelease or draft release")
	}

	owner, repo := *release.Repo.Owner.Login, *release.Repo.Name

	milestones, _, err := context.GitHub.Issues.ListMilestones(context.Context(), owner, repo, &github.MilestoneListOptions{
		State:       "open",
		ListOptions: github.ListOptions{Page: 0, PerPage: 200},
	})
	if err != nil {
		return context.NewError("chlog.CloseMilestoneOnRelease: couldn't fetch milestones for %s/%s: %+v", owner, repo, err)
	}

	for _, milestone := range milestones {
		if *milestone.Title == *release.Release.TagName {
			context.Log("chlog.CloseMilestoneOnRelease: found milestone (%d)", *milestone.Number)

			_, _, err := context.GitHub.Issues.EditMilestone(
				context.Context(), owner, repo, *milestone.Number, &github.Milestone{State: github.String("closed")})
			if err != nil {
				return context.NewError("chlog.CloseMilestoneOnRelease: couldn't close milestone for %s/%s: %+v", owner, repo, err)
			}
		}
	}

	context.Log("chlog.CloseMilestoneOnRelease: no milestone with title '%s' on %s/%s", *release.Release.TagName, owner, repo)

	return nil
}
