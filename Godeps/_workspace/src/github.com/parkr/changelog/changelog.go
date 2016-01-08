package changelog

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Changelog represents a changelog in its entirety, containing all the
// versions that are tracked in the changelog. For supported formats, see
// the documentation for Version.
type Changelog struct {
	Versions []*Version
}

// A Markdown string representation of the Changelog.
func (c *Changelog) String() string {
	return join(c.Versions, "\n\n") + "\n"
}

func (c *Changelog) getVersion(versionNum string) *Version {
	for _, v := range c.Versions {
		if v.Version == versionNum {
			return v
		}
	}
	return nil
}

func (c *Changelog) getSubsection(versionNum, subsectionName string) *Subsection {
	for _, s := range c.getVersion(versionNum).Subsections {
		if s.Name == subsectionName {
			return s
		}
	}
	return nil
}

// AddSubsection adds a Subsection with the given name to the given version.
func (c *Changelog) AddSubsection(versionNum string, subsection string) {
	version := c.getVersion(versionNum)
	version.Subsections = append(version.Subsections, &Subsection{Name: subsection})
}

// AddLineToVersion adds a ChangeLine to the given version's direct
// history. This is only to be used when it is inappropriate to add it to a
// subsection, or the version's changes don't warrant subsections.
func (c *Changelog) AddLineToVersion(versionNum string, line *ChangeLine) {
	c.addToChangelines(&c.getVersion(versionNum).History, line)
}

// AddLineToSubsection adds a ChangeLine to the given version's
// subsection's history.
//
// For example, this could be used to add a change to v1.4.2's "Major
// Enhancements" subsection.
func (c *Changelog) AddLineToSubsection(versionNum, subsectionName string, line *ChangeLine) {
	s := c.getSubsection(versionNum, subsectionName)
	c.addToChangelines(&s.History, line)
}

func (c *Changelog) addToChangelines(lines *[]*ChangeLine, line *ChangeLine) {
	*lines = append(*lines, line)
}

// Version contains the data for the changes for a given version. It can
// have both direct history and subsections.
// Acceptable formats:
//
//     ## 2.4.1
//     ## 2.4.1 / 2015-04-23
//
// The version currently cannot be prefixed with a `v`, but a date is
// optional.
type Version struct {
	Version     string
	Date        string
	History     []*ChangeLine
	Subsections []*Subsection
}

// String returns the markdown representation for the version.
func (v *Version) String() string {
	out := fmt.Sprintf("## %s", v.Version)
	if v.Date != "" {
		out += " / " + v.Date
	}
	if len(v.History) > 0 {
		out += "\n\n" + join(v.History, "\n")
	}
	if len(v.Subsections) > 0 {
		out += "\n\n" + join(v.Subsections, "\n\n")
	}
	return out
}

// Subsection contains the data for a given subsection.
// Acceptable format:
//
//     ### Subsection Name Here
//
// Common subsections are "Major Enhancements," and "Bug Fixes."
type Subsection struct {
	Name    string
	History []*ChangeLine
}

// String returns the markdown representation of the subsection.
func (s *Subsection) String() string {
	if len(s.History) > 0 {
		return fmt.Sprintf(
			"### %s\n\n%s",
			s.Name,
			join(s.History, "\n"),
		)
	}
	return ""
}

// ChangeLine contains the data for a single change.
// Acceptable formats:
//
//     * This is a change (#1234)
//     * This is another change. (@parkr)
//     * This is a change w/o a reference.
//
// The references must be encased in parentheses, and only one reference is
// currently supported.
type ChangeLine struct {
	// What the change entails.
	Summary string
	// Reference can be either a username (e.g. @parkr) or a PR number
	// (e.g. #1234).
	Reference string
}

// String returns the markdown representation of the ChangeLine.
// E.g. "  * Added documentation. (#123)"
func (l *ChangeLine) String() string {
	if l.Reference == "" {
		return fmt.Sprintf(
			"  * %s",
			l.Summary,
		)
	}
	return fmt.Sprintf(
		"  * %s (%s)",
		l.Summary,
		l.Reference,
	)
}

// join calls the .String() function of each element in the slice it's
// passed, then joins those strings by the given separator.
func join(lines interface{}, sep string) string {
	s := reflect.ValueOf(lines)
	if s.Kind() != reflect.Slice {
		panic("join given a non-slice type")
	}

	ret := make([]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		vals := s.Index(i).MethodByName("String").Call(nil)
		ret[i] = vals[0].String()
	}

	return strings.Join(ret, sep)
}

// NewChangelog builds a changelog from the file at the provided filename.
func NewChangelog(filename string) (*Changelog, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return NewChangelogFromReader(file)
}

// NewChangelogFromReader builds a changelog from the contents read in
// through the reader it's passed.
func NewChangelogFromReader(reader io.Reader) (*Changelog, error) {
	history := &Changelog{Versions: []*Version{}}
	err := parseChangelog(reader, history)
	if err != nil {
		return nil, err
	}
	return history, nil
}
