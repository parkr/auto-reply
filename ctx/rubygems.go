package ctx

import (
	"github.com/golang/groupcache/lru"
	"github.com/jekyll/dashboard"
)

type rubyGemsClient struct {
	authToken  string
	baseAPIUrl string
	cache      *lru.Cache
}

func NewRubyGemsClient() *rubyGemsClient {
	return &rubyGemsClient{
		baseAPIUrl: "https://rubygems.org/api/v1",
		cache:      lru.New(500),
	}
}

func (c *rubyGemsClient) GetGem(gemName string) (*dashboard.RubyGem, error) {
	if val, ok := c.cache.Get(gemName); ok {
		return val.(*dashboard.RubyGem), nil
	}

	return dashboard.GetRubyGem(gemName)
}

func (c *rubyGemsClient) GetLatestVersion(gemName string) (string, error) {
	gem, err := c.GetGem(gemName)
	if err != nil {
		return "", err
	}
	return gem.Version, nil
}
