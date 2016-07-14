package lgtm

import (
	"testing"
)

func TestLGTMBodyRegexp(t *testing.T) {
	cases := map[string]bool{
		"lgtm":                                  true,
		"LGTM":                                  true,
		"LGTM.":                                 true,
		"LGTM!":                                 true,
		"LGTM!!!":                               true,
		"@jekyllbot: LGTM":                      true,
		"Yeah, this LGTM.":                      true,
		"Then I'll LGTM it.":                    false,
		"I'd love to get a LGTM for this.":      false,
		"@envygeeks, can you give this a LGTM?": false,
	}
	for input, expected := range cases {
		if actual := lgtmBodyRegexp.MatchString(input); actual != expected {
			t.Fatalf("lgtmBodyRegexp expected '%v' but got '%v' for `%s`", expected, actual, input)
		}
	}
}
