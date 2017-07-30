package triage

import (
	"html/template"
	"time"

	"github.com/google/go-github/github"
)

var pretendTemplateInfo = templateInfo{
	RepoName:  "jekyll/test",
	IssueType: "issue",
	IssuesGroupedByLabel: []IssueGrouping{
		{
			Label: "documentation",
			Issues: Issues{
				github.Issue{
					ID:      github.Int(100000),
					Number:  github.Int(125),
					State:   github.String("open"),
					Locked:  github.Bool(true),
					Title:   github.String("Update boolean documentation"),
					Labels:  []github.Label{},
					User:    &github.User{Login: github.String("acontributor")},
					HTMLURL: github.String("https://github.com/jekyll/test/issues/125"),
					Assignees: []*github.User{
						{Login: github.String("areviewer")},
						{Login: github.String("anotherreviewer")},
					},
				},
			},
		},
	},
}

type templateInfo struct {
	RepoName, IssueType  string
	IssuesGroupedByLabel []IssueGrouping
}

func (t templateInfo) LastUpdated() string {
	if len(t.IssuesGroupedByLabel) == 0 {
		return "never"
	}
	return t.IssuesGroupedByLabel[0].lastUpdated.UTC().Format(time.UnixDate)
}

func (t templateInfo) Total() int {
	total := 0
	for _, grouping := range t.IssuesGroupedByLabel {
		total += len(grouping.Issues)
	}
	return total
}

var (
	triageTmpl *template.Template

	triageTmplText = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
  <meta content="origin-when-cross-origin" name="referrer" />
  <link crossorigin="anonymous" href="https://assets-cdn.github.com/assets/frameworks-98550932b9f11a849da143d2dbc9dfaa977a17656514d323ae9ce0d6fa688b60.css" integrity="sha256-mFUJMrnxGoSdoUPS28nfqpd6F2VlFNMjrpzg1vpoi2A=" media="all" rel="stylesheet" />
  <link crossorigin="anonymous" href="https://assets-cdn.github.com/assets/github-a17b8c9d020ded73daa7ee1a3844b4512f12076634d9249861ddf86dc33da66e.css" integrity="sha256-oXuMnQIN7XPap+4aOES0US8SB2Y02SSYYd34bcM9pm4=" media="all" rel="stylesheet" />
  <title>Dashboard Triage</title>
  <style type="text/css">
  pre {white-space: pre-wrap; padding: 1%;}
  hr {border: none; border-top: 2px solid #000; height: 5px; border-bottom: 1px solid #000;}
  </style>
</head>
<body>

<pre><b>{{.RepoName}} {{.IssueType}} triage</b>
Last updated <b>{{.LastUpdated}}</b>

<b>{{.Total}} pending {{.IssueType}}s</b>

Filter by using the <b>type</b> (issue/pr/all), <b>label</b> (see below), or <b>repo</b> (nwo, e.g. "jekyll/jekyll") URL parameters. You can reorder by creation date using <b>order</b> key (asc/desc). Example: <a href="/triage?type=issue&label=documentation&order=desc">view only "documentation" issues starting with newest</a>.

<hr><b><font size='+1'>Pending {{.IssueType}}s</font></b>

{{range .IssuesGroupedByLabel}}
<b>{{.Label}}</b> ({{len .Issues}} issues){{range .Issues}}
    <a href="{{.HTMLURL}}" target="_blank" title="issue {{.Number}}">{{issueType .}} {{.Number}}</a>{{printf "\t"}}{{.Title}}
                    {{.User.GetLogin}} →{{range .Assignees}} {{.Login}}{{end}}{{if len .Assignees | eq 0}} ???{{end}}, {{daysAgo .GetUpdatedAt}}/{{daysAgo .GetCreatedAt}} days, waiting for {{if hasLabel . "pending-feedback"}}author{{else}}reviewer{{end}}
{{end}}
    {{if len .Issues | eq 0}}no issues!{{end}}
{{end}}
</pre>

</body>
</html>
`
)

func init() {
	var err error
	triageTmpl, err = template.New("triage").Funcs(template.FuncMap{
		"daysAgo":   daysAgo,
		"hasLabel":  hasLabel,
		"issueType": issueTypeForIssue,
	}).Parse(triageTmplText)
	if err != nil {
		panic(err)
	}
}

func daysAgo(t time.Time) int {
	return int(time.Now().Sub(t).Seconds() / 86400)
}

func hasLabel(issue github.Issue, desired string) bool {
	for _, label := range issue.Labels {
		if label.GetName() == desired {
			return true
		}
	}
	return false
}

func issueTypeForIssue(issue github.Issue) template.HTML {
	if issue.PullRequestLinks != nil {
		return `<svg aria-hidden="true" class="octicon octicon-git-pull-request" height="16" version="1.1" viewBox="0 0 12 16" width="12"><path fill-rule="evenodd" d="M11 11.28V5c-.03-.78-.34-1.47-.94-2.06C9.46 2.35 8.78 2.03 8 2H7V0L4 3l3 3V4h1c.27.02.48.11.69.31.21.2.3.42.31.69v6.28A1.993 1.993 0 0 0 10 15a1.993 1.993 0 0 0 1-3.72zm-1 2.92c-.66 0-1.2-.55-1.2-1.2 0-.65.55-1.2 1.2-1.2.65 0 1.2.55 1.2 1.2 0 .65-.55 1.2-1.2 1.2zM4 3c0-1.11-.89-2-2-2a1.993 1.993 0 0 0-1 3.72v6.56A1.993 1.993 0 0 0 2 15a1.993 1.993 0 0 0 1-3.72V4.72c.59-.34 1-.98 1-1.72zm-.8 10c0 .66-.55 1.2-1.2 1.2-.65 0-1.2-.55-1.2-1.2 0-.65.55-1.2 1.2-1.2.65 0 1.2.55 1.2 1.2zM2 4.2C1.34 4.2.8 3.65.8 3c0-.65.55-1.2 1.2-1.2.65 0 1.2.55 1.2 1.2 0 .65-.55 1.2-1.2 1.2z"/></svg>`
	}
	return `<svg aria-hidden="true" class="octicon octicon-issue-opened" height="16" version="1.1" viewBox="0 0 14 16" width="14"><path fill-rule="evenodd" d="M7 2.3c3.14 0 5.7 2.56 5.7 5.7s-2.56 5.7-5.7 5.7A5.71 5.71 0 0 1 1.3 8c0-3.14 2.56-5.7 5.7-5.7zM7 1C3.14 1 0 4.14 0 8s3.14 7 7 7 7-3.14 7-7-3.14-7-7-7zm1 3H6v5h2V4zm0 6H6v2h2v-2z"/></svg>`
}
