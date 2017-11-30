package jekyll

type Repository struct {
	Name string
}

// Always the Jekyll org.
func (r Repository) Owner() string {
	return "jekyll"
}

// String returns NWO.
func (r Repository) String() string {
	return r.Owner() + "/" + r.Name
}

var DefaultRepos = []Repository{
	{Name: "github-metadata"},
	{Name: "jekyll"},
	{Name: "jekyll-admin"},
	{Name: "jekyll-coffeescript"},
	{Name: "jekyll-compose"},
	{Name: "jekyll-feed"},
	{Name: "jekyll-gist"},
	{Name: "jekyll-import"},
	{Name: "jekyll-redirect-from"},
	{Name: "jekyll-sass-converter"},
	{Name: "jekyll-seo-tag"},
	{Name: "jekyll-sitemap"},
	{Name: "jekyll-watch"},
	{Name: "jemoji"},
	{Name: "minima"},
	{Name: "plugins"},
}
