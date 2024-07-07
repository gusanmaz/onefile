package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gusanmaz/onefile/internal"
	"github.com/spf13/cobra"
)

func NewGitHub2FileCmd() *cobra.Command {
	var repoURL, outputType, outputDir, outputName string
	var includePatterns, excludePatterns []string
	var allRepos bool
	var cmd = &cobra.Command{
		Use:   "github2file",
		Short: "Fetch a GitHub repository and save as JSON or Markdown",
		Long:  `Fetch a GitHub repository and save its structure and contents as JSON or Markdown.`,
		Run: func(cmd *cobra.Command, args []string) {
			if repoURL == "" && !allRepos {
				fmt.Println("Please provide a GitHub URL or use the -a flag")
				return
			}

			if allRepos {
				owner, _, _, err := internal.ParseGitHubURL(repoURL)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error parsing GitHub URL: %v\n", err)
					return
				}

				repos, err := internal.FetchUserRepos(owner)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error fetching user repositories: %v\n", err)
					return
				}

				for _, repo := range repos {
					fetchAndSaveRepo(fmt.Sprintf("https://github.com/%s/%s", owner, repo.Name), outputType, outputDir, outputName, includePatterns, excludePatterns)
				}
			} else {
				fetchAndSaveRepo(repoURL, outputType, outputDir, outputName, includePatterns, excludePatterns)
			}
		},
	}

	cmd.Flags().StringVarP(&repoURL, "url", "u", "", "GitHub repository URL")
	cmd.Flags().StringVarP(&outputType, "type", "t", "md", "Output type: json or md")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "d", ".", "Output directory")
	cmd.Flags().StringVarP(&outputName, "output-name", "n", "", "Output file name (without extension)")
	cmd.Flags().StringSliceVarP(&includePatterns, "include", "i", []string{"*"}, "Patterns to include files (space-separated)")
	cmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Patterns to exclude files (space-separated)")
	cmd.Flags().BoolVarP(&allRepos, "all-repos", "a", false, "Fetch all repositories for a user")

	return cmd
}

func fetchAndSaveRepo(repoURL, outputType, outputDir, outputName string, includePatterns, excludePatterns []string) {
	owner, repo, path, err := internal.ParseGitHubURL(repoURL)
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

	projectData, err := internal.FetchGithubRepo(owner, repo, path, includePatterns, excludePatterns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching GitHub repo: %v\n", err)
		return
	}

	outputPath := filepath.Join(outputDir, outputName+"."+outputType)

	if outputType == "json" {
		err = internal.SaveAsJSON(projectData, outputPath)
	} else if outputType == "md" {
		err = internal.SaveAsMarkdown(projectData, outputPath)
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
