package ctx

import "fmt"

type repoRef struct {
	Owner string
	Name  string
}

func (r repoRef) String() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

func (r repoRef) IsEmpty() bool {
	return r.Owner == "" || r.Name == ""
}

func (c *Context) SetRepo(owner, repo string) {
	c.Repo = repoRef{
		Owner: owner,
		Name:  repo,
	}
}
