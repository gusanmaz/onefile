package internal

import (
	"bufio"
	"github.com/sabhiram/go-gitignore"
	"os"
	"strings"
)

func loadPatternsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if pattern != "" && !strings.HasPrefix(pattern, "#") {
			patterns = append(patterns, pattern)
		}
	}

	return patterns, scanner.Err()
}

func CreateGitIgnoreMatcher(patterns []string) *ignore.GitIgnore {
	return ignore.CompileIgnoreLines(patterns...)
}

func MatchesPatterns(path string, gitIgnore *ignore.GitIgnore, includeGit, includeNonText bool) bool {
	if !includeGit && (strings.HasPrefix(path, ".git"+string(os.PathSeparator)) || path == ".git") {
		return false
	}
	if gitIgnore.MatchesPath(path) {
		return false
	}
	return includeNonText || isTextFile(path)
}

func ParsePatterns(patterns []string) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "@") {
			filePatterns, err := loadPatternsFromFile(strings.TrimPrefix(pattern, "@"))
			if err != nil {
				return nil, err
			}
			result = append(result, filePatterns...)
		} else {
			result = append(result, pattern)
		}
	}
	return result, nil
}

func getFilePaths(files []FileData) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.Path
	}
	return paths
}
