package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GenerateMarkdown(projectData ProjectData, includeGit, includeNonText bool) string {
	var md strings.Builder

	md.WriteString("# Project Structure\n\n")
	md.WriteString("```\n")
	md.WriteString(generateProjectTree(projectData, includeGit, includeNonText))
	md.WriteString("```\n\n")

	md.WriteString("## Shell Commands to Create Project Structure\n\n")
	md.WriteString("```bash\n")
	md.WriteString(GenerateShellCommands(projectData, includeGit, includeNonText))
	md.WriteString("```\n\n")

	md.WriteString("## File Contents\n\n")
	for _, file := range projectData.Files {
		if file.Content != "" && (includeGit || !strings.HasPrefix(file.Path, ".git/")) && (includeNonText || isTextFile(file.Path)) {
			language := getLanguageFromExtension(file.Path)
			md.WriteString(fmt.Sprintf("### %s\n\n```%s\n%s\n```\n\n", file.Path, language, file.Content))
		}
	}

	return md.String()
}

func generateProjectTree(projectData ProjectData, includeGit, includeNonText bool) string {
	var tree strings.Builder
	tree.WriteString(".\n")

	var allPaths []string
	for _, dir := range projectData.Directories {
		if includeGit || !strings.HasPrefix(dir, ".git") {
			allPaths = append(allPaths, dir)
		}
	}
	for _, file := range projectData.Files {
		if (includeGit || !strings.HasPrefix(file.Path, ".git")) && (includeNonText || isTextFile(file.Path)) {
			allPaths = append(allPaths, file.Path)
		}
	}
	sort.Strings(allPaths)

	for i, path := range allPaths {
		parts := strings.Split(path, string(os.PathSeparator))
		for j, part := range parts {
			isLast := i == len(allPaths)-1 && j == len(parts)-1
			prefix := strings.Repeat("│   ", j)
			if isLast {
				tree.WriteString(fmt.Sprintf("%s└── %s\n", prefix, part))
			} else {
				tree.WriteString(fmt.Sprintf("%s├── %s\n", prefix, part))
			}
		}
	}

	return tree.String()
}

func GenerateShellCommands(projectData ProjectData, includeGit, includeNonText bool) string {
	var commands strings.Builder

	for _, dir := range projectData.Directories {
		if includeGit || !strings.HasPrefix(dir, ".git") {
			commands.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", dir))
		}
	}

	for _, file := range projectData.Files {
		if (includeGit || !strings.HasPrefix(file.Path, ".git/")) && (includeNonText || isTextFile(file.Path)) {
			dir := filepath.Dir(file.Path)
			if dir != "." {
				commands.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", dir))
			}
			commands.WriteString(fmt.Sprintf("touch \"%s\"\n", file.Path))
		}
	}

	return commands.String()
}

func SaveAsMarkdown(projectData ProjectData, outputPath string, includeGit, includeNonText bool) error {
	markdown := GenerateMarkdown(projectData, includeGit, includeNonText)
	return ioutil.WriteFile(outputPath, []byte(markdown), 0644)
}
