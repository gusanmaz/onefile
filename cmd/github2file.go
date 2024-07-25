package cmd

import (
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"os"
	"path/filepath"
	"strings"

	"github.com/gusanmaz/onefile/utils"
	"github.com/spf13/cobra"
)

func NewGitHub2FileCmd() *cobra.Command {
	var repoURL, outputType, outputDir, outputName, githubToken string
	var excludePatterns []string
	var allRepos, useGit, includeGit, includeNonText, showExcluded bool
	var cmd = &cobra.Command{
		Use:   "github2file",
		Short: "Fetch a GitHub repository and save as JSON or Markdown",
		Long: `Fetch a GitHub repository and save its structure and contents as JSON or Markdown.
Exclude patterns can be specified directly or by referencing a file with @.
Example: -e "*.go @.gitignore" -e "utils/extension_language_map.json go.mod go.sum"`,
		Run: func(cmd *cobra.Command, args []string) {
			if repoURL == "" && !allRepos {
				fmt.Println("Please provide a GitHub URL or use the -a flag")
				return
			}

			// Process exclude patterns
			var processedPatterns []string
			for _, pattern := range excludePatterns {
				patterns := strings.Fields(pattern)
				processedPatterns = append(processedPatterns, patterns...)
			}

			parsedExcludePatterns, err := utils.ParsePatterns(processedPatterns)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing exclude patterns: %v\n", err)
				return
			}

			gitIgnore := utils.CreateGitIgnoreMatcher(parsedExcludePatterns)

			if allRepos {
				owner, _, _, err := utils.ParseGitHubURL(repoURL)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error parsing GitHub URL: %v\n", err)
					return
				}

				repos, err := utils.FetchUserRepos(owner, githubToken)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error fetching user repositories: %v\n", err)
					return
				}

				for _, repo := range repos {
					fetchAndSaveRepo(fmt.Sprintf("https://github.com/%s/%s", owner, repo.Name), outputType, outputDir, outputName, gitIgnore, useGit, githubToken, includeGit, includeNonText, showExcluded)
				}
			} else {
				fetchAndSaveRepo(repoURL, outputType, outputDir, outputName, gitIgnore, useGit, githubToken, includeGit, includeNonText, showExcluded)
			}
		},
	}

	cmd.Flags().StringVarP(&repoURL, "url", "u", "", "GitHub repository URL")
	cmd.Flags().StringVarP(&outputType, "type", "t", "md", "Output type: json or md")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "d", ".", "Output directory")
	cmd.Flags().StringVarP(&outputName, "output-name", "n", "", "Output file name (without extension)")
	cmd.Flags().StringArrayVarP(&excludePatterns, "exclude", "e", []string{}, "Patterns to exclude files (Use @ for file-based patterns, e.g., @.gitignore)")
	cmd.Flags().BoolVarP(&allRepos, "all-repos", "a", false, "Fetch all repositories for a user")
	cmd.Flags().BoolVarP(&useGit, "use-git", "g", true, "Use git clone if available")
	cmd.Flags().StringVarP(&githubToken, "token", "k", "", "GitHub API token")
	cmd.Flags().BoolVar(&includeGit, "include-git", false, "Include .git files and directories")
	cmd.Flags().BoolVar(&includeNonText, "include-non-text", false, "Include non-text files")
	cmd.Flags().BoolVar(&showExcluded, "show-excluded", false, "Show excluded files in project structure and shell commands")

	return cmd
}

func fetchAndSaveRepo(repoURL, outputType, outputDir, outputName string, gitIgnore *ignore.GitIgnore, useGit bool, githubToken string, includeGit, includeNonText, showExcluded bool) {
	owner, repo, path, err := utils.ParseGitHubURL(repoURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing GitHub URL: %v\n", err)
		return
	}

	if outputName == "" {
		outputName = fmt.Sprintf("%s_%s", owner, repo)
		if path != "" {
			outputName += "_" + strings.ReplaceAll(path, "/", "_")
		}
	}

	projectData, err := utils.FetchGithubRepo(owner, repo, path, gitIgnore, useGit, githubToken, includeGit, includeNonText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching GitHub repo: %v\n", err)
		return
	}

	outputPath := filepath.Join(outputDir, outputName+"."+outputType)

	if outputType == "json" {
		err = utils.SaveAsJSON(projectData, outputPath, includeGit, includeNonText)
	} else if outputType == "md" {
		err = utils.SaveAsMarkdown(projectData, outputPath, includeGit, includeNonText, showExcluded)
	} else {
		fmt.Fprintf(os.Stderr, "Invalid output type. Use 'json' or 'md'\n")
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving output: %v\n", err)
		return
	}

	fmt.Printf("Output file created successfully: %s\n", outputPath)
}
