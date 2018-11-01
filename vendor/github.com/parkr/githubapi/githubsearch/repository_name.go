package githubsearch

type RepositoryName struct {
	Owner string
	Name  string
}

func (r RepositoryName) String() string {
	if r.Owner == "" && r.Name == "" {
		return ""
	}
	if r.Owner == "" && r.Name != "" {
		return r.Name
	}
	if r.Owner != "" && r.Name == "" {
		return r.Owner
	}
	return r.Owner + "/" + r.Name
}
