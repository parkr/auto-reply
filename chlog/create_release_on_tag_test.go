package chlog

import (
	"testing"
)

func TestVersionTagRegexpMatchString(t *testing.T) {
	cases := map[string]bool{
		"lgtm":              false,
		"v2.3.0":            true,
		"3.2.0":             false,
		"v13.24.52":         true,
		"v3.2.0.pre.beta12": true,
	}
	for input, expected := range cases {
		if actual := versionTagRegexp.MatchString(input); actual != expected {
			t.Fatalf("versionTagRegexp expected '%v' but got '%v' for `%s`", expected, actual, input)
		}
	}
}

func TestExtractVersion(t *testing.T) {
	cases := map[string]string{
		"lgtm":              "",
		"v2.3.0":            "2.3.0",
		"3.2.0":             "",
		"v13.24.52":         "13.24.52",
		"v3.2.0.pre.beta12": "3.2.0.pre.beta12",
	}
	for input, expected := range cases {
		if actual := extractVersion(input); actual != expected {
			t.Fatalf("extractVersion expected '%v' but got '%v' for `%s`", expected, actual, input)
		}
	}
}
