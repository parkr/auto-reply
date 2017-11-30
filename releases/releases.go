package releases

import (
	"errors"
	"fmt"
	"sort"

	"github.com/google/go-github/github"
	"github.com/hashicorp/go-version"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/jekyll"
)

func LatestRelease(context *ctx.Context, repo jekyll.Repository) (*github.RepositoryRelease, error) {
	releases, _, err := context.GitHub.Repositories.ListReleases(context.Context(), repo.Owner(), repo.Name, &github.ListOptions{PerPage: 300})
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, nil
	}

	versions := []*version.Version{}
	for _, release := range releases {
		v, err := version.NewVersion(release.GetTagName())
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	// After this, the versions are properly sorted
	sort.Sort(sort.Reverse(version.Collection(versions)))

	desiredVersionTagName := "v" + versions[0].String()

	for _, release := range releases {
		if release.GetTagName() == desiredVersionTagName {
			return release, nil
		}
	}

	return nil, errors.New("couldn't determine the latest version")
}

func CommitsSinceRelease(context *ctx.Context, repo jekyll.Repository, latestRelease *github.RepositoryRelease) (int, error) {
	comparison, _, err := context.GitHub.Repositories.CompareCommits(
		context.Context(),
		repo.Owner(), repo.Name,
		latestRelease.GetTagName(), "master",
	)
	if err != nil {
		return -1, fmt.Errorf("error fetching commit comparison for %s...master for %s: %v", latestRelease.GetTagName(), repo, err)
	}

	return comparison.GetTotalCommits(), nil
}
