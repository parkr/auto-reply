package dependencies

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/parkr/auto-reply/ctx"
)

type RubyDependency struct {
	name       string
	constraint version.Constraints
	latest     *version.Version
	isOutdated *bool
}

func (d *RubyDependency) String() string {
	return fmt.Sprintf(
		"name:%+v constraint:%+v latest %+v isOutdated:%v",
		d.name, d.constraint, d.latest, *d.isOutdated,
	)
}

func (d *RubyDependency) GetName() string {
	return d.name
}

func (d *RubyDependency) GetConstraint() version.Constraints {
	return d.constraint
}

func (d *RubyDependency) GetLatestVersion(context *ctx.Context) *version.Version {
	if d.latest != nil {
		return d.latest
	}

	versionStr, err := context.RubyGems.GetLatestVersion(d.name)
	if err != nil {
		context.Log("dependencies: could not fetch latest version on rubygems for %s: %v", d.name, err)
		return nil
	}

	ver, err := version.NewVersion(versionStr)
	if err != nil {
		context.Log("dependencies: could not parse version %+v for %s: %v", versionStr, d.name, err)
		return nil
	}

	d.latest = ver
	return d.latest
}

func (d *RubyDependency) IsOutdated(context *ctx.Context) bool {
	if d.isOutdated != nil {
		return *d.isOutdated
	}

	latestVersion := d.GetLatestVersion(context)
	if latestVersion == nil {
		return false
	}

	isOutdated := !d.GetConstraint().Check(latestVersion)
	d.isOutdated = &isOutdated
	return *d.isOutdated
}
