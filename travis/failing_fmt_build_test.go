package travis

import "testing"

func TestBuildIDFromTargetURL(t *testing.T) {
	expected := int64(215667761)
	targetURL := "https://travis-ci.org/jekyll/jekyll/builds/215667761"
	actual, err := buildIDFromTargetURL(targetURL)
	if err != nil {
		t.Errorf("failed: expected no error but got %+v", err)
	}
	if actual != expected {
		t.Errorf("failed: expected %d from %q, but got: %d", expected, targetURL, actual)
	}
}
