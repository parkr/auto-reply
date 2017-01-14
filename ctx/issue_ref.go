package ctx

import "fmt"

// issueRef is used to refer to an issue or pull request
type issueRef struct {
	Author      string
	Owner, Repo string
	Num         int
}

func (r issueRef) String() string {
	if r.Num < 0 {
		return fmt.Sprintf("%s/%s", r.Owner, r.Repo)
	} else {
		return fmt.Sprintf("%s/%s#%d", r.Owner, r.Repo, r.Num)
	}
}

func (r issueRef) IsEmpty() bool {
	return r.Owner == "" || r.Repo == "" || r.Num == 0
}

func (c *Context) SetIssue(owner, repo string, num int) {
	c.Issue = issueRef{
		Owner: owner,
		Repo:  repo,
		Num:   num,
	}
}

func (c *Context) SetAuthor(author string) {
	c.Issue.Author = author
}
