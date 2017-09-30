package filepath

import (
	"fmt"
	"testing"
)

func TestClean(t *testing.T) {
	for _, test := range []struct {
		os       string
		path     string
		expected string
	}{
		{
			os:       "linux",
			path:     "/",
			expected: "/",
		},
		{
			os:       "linux",
			path:     "//",
			expected: "/",
		},
		{
			os:       "linux",
			path:     "/a",
			expected: "/a",
		},
		{
			os:       "linux",
			path:     "/a/",
			expected: "/a",
		},
		{
			os:       "linux",
			path:     "//a",
			expected: "/a",
		},
		{
			os:       "linux",
			path:     ".",
			expected: ".",
		},
		{
			os:       "linux",
			path:     "./c",
			expected: "c",
		},
		{
			os:       "linux",
			path:     ".././a",
			expected: "../a",
		},
		{
			os:       "linux",
			path:     "a/../b",
			expected: "b",
		},
	} {
		t.Run(
			fmt.Sprintf("Clean(%q,%q)", test.os, test.path),
			func(t *testing.T) {
				clean := Clean(test.os, test.path)
				if clean != test.expected {
					t.Errorf("unexpected result: %q (expected %q)\n", clean, test.expected)
				}
			},
		)
	}
}
