package dashboard

import (
	"fmt"
	"log"
)

type RubyGem struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Downloads        int    `json:"downloads"`
	HomepageURI      string `json:"homepage_uri"`
	DocumentationURI string `json:"documentation_uri"`
}

func GetRubyGem(gem string) (*RubyGem, error) {
	if gem == "" {
		return nil, nil
	}

	info := &RubyGem{}
	err := getRetry(5, fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", gem), info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func rubygem(gem string) chan *RubyGem {
	rubyGemChan := make(chan *RubyGem, 1)

	go func() {
		info, err := GetRubyGem(gem)
		if err != nil {
			log.Printf("error fetching rubygems info for %s: %v", gem, err)
		}
		rubyGemChan <- info
		close(rubyGemChan)
	}()

	return rubyGemChan
}
