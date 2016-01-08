package changelog

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"strings"
)

var (
	versionRegexp    = regexp.MustCompile(`## (?i:(HEAD|\d+.\d+(.\d+)?)( / (\d{4}-\d{2}-\d{2}))?)`)
	subheaderRegexp  = regexp.MustCompile(`### ([0-9A-Za-z_ ]+)`)
	changeLineRegexp = regexp.MustCompile(`\* (.+)( \(((#[0-9]+)|(@[[:word:]]+))\))?`)

	verbose = false
)

// SetVerbose sets the verbose flag to the value passed.
// If true is passed, verbose logging will be enabled.
func SetVerbose(v bool) {
	verbose = v
}

func logVerbose(args ...interface{}) {
	if verbose == true {
		log.Println(args...)
	}
}

func parseChangelog(file io.Reader, history *Changelog) error {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	currentHeader := ""
	currentSubHeader := ""
	var currentLine *ChangeLine
	for scanner.Scan() {
		txt := scanner.Text()
		logVerbose(txt)
		logVerbose("isHeader", versionRegexp.MatchString(txt))
		if versionRegexp.MatchString(txt) {
			matches := versionRegexp.FindStringSubmatch(txt)
			logVerbose("headerMatches:", matches, len(matches))
			currentHeader = matches[1]
			currentSubHeader = ""
			logVerbose("currentHeader: '%s'", currentHeader)
			var date string
			if len(matches) == 5 {
				date = matches[4]
			}
			history.Versions = append(history.Versions, &Version{
				Version:     currentHeader,
				Date:        date,
				History:     []*ChangeLine{},
				Subsections: []*Subsection{},
			})
			continue
		}

		logVerbose("isSubHeader", subheaderRegexp.MatchString(txt))
		if subheaderRegexp.MatchString(txt) {
			matches := subheaderRegexp.FindStringSubmatch(txt)
			logVerbose("subHeaderMatches:", matches, len(matches))
			currentSubHeader = matches[1]
			logVerbose("currentSubHeader: '%s'", currentSubHeader)
			history.AddSubsection(currentHeader, currentSubHeader)
			continue
		}

		logVerbose("isChangeLine", changeLineRegexp.MatchString(txt))
		if changeLineRegexp.MatchString(txt) {
			matches := changeLineRegexp.FindStringSubmatch(txt)
			logVerbose("changeLineMatches:", matches, len(matches))
			line := &ChangeLine{
				Summary:   matches[1],
				Reference: matches[3],
			}
			logVerbose("newChangeLine:", line)
			currentLine = line
			if currentSubHeader == "" {
				history.AddLineToVersion(currentHeader, line)
			} else {
				history.AddLineToSubsection(currentHeader, currentSubHeader, line)
			}
			continue
		} else {
			if strings.TrimSpace(txt) != "" && currentLine != nil {
				currentLine.Summary += " " + strings.TrimSpace(txt)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("error reading history:", err)
	}
	return nil
}
