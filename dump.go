package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Function to check if a file matches a pattern from the whitelist or blacklist
func matchesPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error matching pattern %s: %v\n", pattern, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// Function to read patterns from a file
func readPatternsFromFile(filePath string) ([]string, error) {
	var patterns []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if pattern != "" && !strings.HasPrefix(pattern, "#") {
			patterns = append(patterns, pattern)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// Function to dump project data to JSON
func dumpProjectToJSON(rootPath, outputPath string, whitelistPatterns, blacklistPatterns []string) {
	var projectData ProjectData

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			projectData.Directories = append(projectData.Directories, relativePath)
		} else if isTextFile(filepath.Ext(info.Name())) {
			if matchesPattern(relativePath, blacklistPatterns) {
				return nil
			}
			if len(whitelistPatterns) > 0 && !matchesPattern(relativePath, whitelistPatterns) {
				return nil
			}
			content, err := readFileContent(path)
			if err != nil {
				return err
			}
			projectData.Files = append(projectData.Files, FileData{Path: relativePath, Content: content})
		}
		return nil
	})
	check(err, "walking the file path")

	data, err := json.MarshalIndent(projectData, "", "  ")
	check(err, "marshaling JSON")

	err = ioutil.WriteFile(outputPath, data, 0644)
	check(err, "writing JSON to file")
}

func main() {
	rootPath := flag.String("path", ".", "The root path of the project")
	outputPath := flag.String("output", "project_data.json", "The output JSON file")
	whitelistFile := flag.String("whitelist", "", "Path to the whitelist file (optional)")
	blacklistFile := flag.String("blacklist", "", "Path to the blacklist file (optional)")

	flag.Parse()

	absRootPath, err := filepath.Abs(strings.TrimSpace(*rootPath))
	check(err, "getting absolute root path")

	var whitelistPatterns, blacklistPatterns []string

	if *whitelistFile != "" {
		whitelistPatterns, err = readPatternsFromFile(*whitelistFile)
		check(err, "reading whitelist file")
	}

	if *blacklistFile != "" {
		blacklistPatterns, err = readPatternsFromFile(*blacklistFile)
		check(err, "reading blacklist file")
	}

	dumpProjectToJSON(absRootPath, *outputPath, whitelistPatterns, blacklistPatterns)

	fmt.Printf("Project data dumped to JSON file: %s\n", *outputPath)
}
