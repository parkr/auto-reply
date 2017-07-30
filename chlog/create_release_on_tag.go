package chlog

import (
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
)

var versionTagRegexp = regexp.MustCompile(`v(\d+\.\d+\.\d+)(\.pre\.(beta|rc)\d+)?`)

func CreateReleaseOnTagHandler(context *ctx.Context, payload interface{}) error {
	create, ok := payload.(*github.CreateEvent)
	if !ok {
		return context.NewError("chlog.CreateReleaseOnTagHandler: not a create event")
	}

	if *create.RefType != "tag" {
		return context.NewError("chlog.CreateReleaseOnTagHandler: not a tag create event")
	}

	version := extractVersion(*create.Ref)
	if version == "" {
		return context.NewError("chlog.CreateReleaseOnTagHandler: not a version tag (%s)", *create.Ref)
	}

	isPreRelease := strings.Index(version, ".pre") >= 0
	desiredRef := version
	if isPreRelease {
		// Working with a pre-release. Use HEAD.
		desiredRef = "HEAD"
	}

	owner, name := *create.Repo.Owner.Login, *create.Repo.Name

	// Read History.markdown, add line to appropriate change section
	historyFileContents, _ := getHistoryContents(context, owner, name)
	changes, err := parseChangelog(historyFileContents)
	if err != nil {
		return context.NewError("chlog.CreateReleaseOnTagHandler: could not parse history file: %v", err)
	}

	versionLog := changes.GetVersion(desiredRef)
	if versionLog == nil {
		return context.NewError("chlog.CreateReleaseOnTagHandler: no '%s' version in history file", desiredRef)
	}

	releaseBodyForVersion := strings.Join(strings.SplitN(versionLog.String(), "\n\n", 2)[1:], "\n")

	_, _, err = context.GitHub.Repositories.CreateRelease(
		context.Context(),
		owner, name,
		&github.RepositoryRelease{
			TagName:    create.Ref,
			Name:       create.Ref,
			Body:       github.String(releaseBodyForVersion),
			Draft:      github.Bool(false),
			Prerelease: github.Bool(isPreRelease),
		})
	if err != nil {
		context.NewError("chlog.CreateReleaseOnTagHandler: error creating release: %v", err)
	}

	return nil
}

func extractVersion(tag string) string {
	if versionTagRegexp.MatchString(tag) {
		return strings.Replace(tag, "v", "", 1)
	}
	return ""
}
