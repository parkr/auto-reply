package dependencies

import (
	"bufio"
	"encoding/base64"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/parkr/auto-reply/ctx"
)

var (
	dependencyCache = map[string]Dependency{}

	gemNameRegexp = `(\(|\s+)("|')([\w-_]+)("|')`
	versionRegexp = `(,\s*("|')(.+)("|')(\)|\s+)?)?`

	gemspecRegexp                = regexp.MustCompile(`\.add_(runtime|development)_dependency` + gemNameRegexp + versionRegexp)
	gemspecRegexpNameIndex       = 4
	gemspecRegexpConstraintIndex = 8

	gemfileRegexp                = regexp.MustCompile(`gem` + gemNameRegexp + versionRegexp)
	gemfileRegexpNameIndex       = 3
	gemfileRegexpConstraintIndex = 7
)

type lineParserFunc func(line string) Dependency

type Checker interface {
	AllOutdatedDependencies(context *ctx.Context) []Dependency
}

type rubyDependencyChecker struct {
	owner, name  string
	dependencies []Dependency
}

func (r *rubyDependencyChecker) AllOutdatedDependencies(context *ctx.Context) []Dependency {
	err := r.parseFileForDependencies(context, r.name+".gemspec", r.parseGemspecDependency)
	if err != nil {
		context.Log("dependencies: couldn't parse gemspec for %s/%s: %v", r.owner, r.name, err)
	}

	err = r.parseFileForDependencies(context, "Gemfile", r.parseGemfileDependency)
	if err != nil {
		context.Log("dependencies: couldn't parse gemfile for %s/%s: %v", r.owner, r.name, err)
	}

	outdated := []Dependency{}
	for _, dep := range r.dependencies {
		if dep.IsOutdated(context) {
			outdated = append(outdated, dep)
		}
	}

	return outdated
}

func (r *rubyDependencyChecker) hasDependency(name string) bool {
	for _, dep := range r.dependencies {
		if dep.GetName() == name {
			return true
		}
	}
	return false
}

// parseGemspec finds the gemspec for the project and appends any dependencies
// it finds to the list of dependencies stored on the rubyDependencyChecker instance.
func (r *rubyDependencyChecker) parseFileForDependencies(context *ctx.Context, path string, lineParser lineParserFunc) error {
	contents := r.fetchFile(context, path)
	if contents == "" {
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		line := scanner.Text()
		dependency := lineParser(line)
		if dependency != nil && !r.hasDependency(dependency.GetName()) {
			r.dependencies = append(r.dependencies, dependency)
		}
	}

	return scanner.Err()
}

func (r *rubyDependencyChecker) parseGemspecDependency(line string) Dependency {
	match := gemspecRegexp.FindAllStringSubmatch(line, -1)
	if len(match) >= 1 && len(match[0]) >= gemspecRegexpConstraintIndex+1 {
		name := match[0][gemspecRegexpNameIndex]
		constraintStr := match[0][gemspecRegexpConstraintIndex]

		// If there is no constraint, there's no way it could be outdated!
		if constraintStr == "" {
			return nil
		}

		log.Printf("%+v %+v", name, constraintStr)
		constraint, err := version.NewConstraint(constraintStr)
		if err == nil {
			return &RubyDependency{name: name, constraint: constraint}
		} else {
			log.Printf("dependencies: can't parse constraint %+v for dependency %s", constraintStr, name)
		}
	}
	return nil
}

func (r *rubyDependencyChecker) parseGemfileDependency(line string) Dependency {
	match := gemfileRegexp.FindAllStringSubmatch(line, -1)
	if len(match) >= 1 && len(match[0]) >= gemfileRegexpConstraintIndex+1 {
		name := match[0][gemfileRegexpNameIndex]
		constraintStr := match[0][gemfileRegexpConstraintIndex]

		// If there is no constraint, there's no way it could be outdated!
		if constraintStr == "" {
			return nil
		}

		constraint, err := version.NewConstraint(constraintStr)
		if err == nil {
			return &RubyDependency{name: name, constraint: constraint}
		} else {
			log.Printf("dependencies: can't parse constraint %+v for dependency %s", constraintStr, name)
		}
	}
	return nil
}

func (r *rubyDependencyChecker) fetchFile(context *ctx.Context, path string) string {
	contents, _, _, err := context.GitHub.Repositories.GetContents(context.Context(), r.owner, r.name, path, nil)
	if err != nil {
		context.Log("dependencies: error getting %s from %s/%s: %v", path, r.owner, r.name, err)
		return ""
	}
	return base64Decode(*contents.Content)
}

func base64Decode(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		log.Printf("dependencies: error decoding string: %v\n", err)
		return ""
	}
	return string(decoded)
}

func NewRubyDependencyChecker(repoOwner, repoName string) Checker {
	return &rubyDependencyChecker{owner: repoOwner, name: repoName, dependencies: []Dependency{}}
}
