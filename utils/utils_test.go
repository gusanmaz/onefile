package utils

import (
	"testing"

	"github.com/sabhiram/go-gitignore"
)

func TestMatchesPatterns(t *testing.T) {
	patterns := []string{"*.txt", "!important.txt", "temp/"}
	gitIgnore := ignore.CompileIgnoreLines(patterns...)

	testCases := []struct {
		path           string
		includeGit     bool
		includeNonText bool
		expected       bool
	}{
		{"file.txt", true, true, false},
		{"important.txt", true, true, true},
		{"file.go", true, true, true},
		{".git/config", false, true, false},
		{".git/config", true, true, true},
		{"temp/file.go", true, true, false},
	}

	for _, tc := range testCases {
		result := MatchesPatterns(tc.path, gitIgnore, tc.includeGit, tc.includeNonText)
		if result != tc.expected {
			t.Errorf("MatchesPatterns(%q, %v, %v) = %v; want %v",
				tc.path, tc.includeGit, tc.includeNonText, result, tc.expected)
		}
	}
}
