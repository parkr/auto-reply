package lgtm

import (
	"testing"
)

func TestLGTMBodyRegexp(t *testing.T) {
	cases := map[string]bool{
		"lgtm":               true,
		"LGTM":               true,
		"LGTM.":              true,
		"@jekyllbot: LGTM":   true,
		"Yeah, this LGTM.":   true,
		"Then I'll LGTM it.": false,
	}
	for input, expected := range cases {
		if actual := lgtmBodyRegexp.MatchString(input); actual != expected {
			t.Fatalf("lgtmBodyRegexp expected '%v' but got '%v' for `%s`", expected, actual, input)
		}
	}
}

func TestGenerateDescription(t *testing.T) {
	cases := []struct {
		lgtmers     []string
		description string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"parkr"}, ""},
	}
	for _, testCase := range cases {
		actual := generateDescription(testCase.lgtmers)
		if actual != testCase.description {
			t.Fatalf("description for %+v should be `%s`, not: `%s`",
				testCase.lgtmers, testCase.description, actual)
		}
	}
}
