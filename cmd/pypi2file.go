package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gusanmaz/onefile/utils"
	"github.com/spf13/cobra"
)

func NewPyPI2FileCmd() *cobra.Command {
	var packageName, outputType, outputDir, outputName string
	var excludePatterns []string
	var includeGit, includeNonText, showExcluded bool
	var cmd = &cobra.Command{
		Use:   "pypi2file",
		Short: "Fetch a PyPI package and save as JSON or Markdown",
		Long:  `Fetch a PyPI package and save its structure and contents as JSON or Markdown.`,
		Run: func(cmd *cobra.Command, args []string) {
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

			projectData, err := utils.FetchPyPIPackage(packageName, gitIgnore, includeGit, includeNonText)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching PyPI package: %v\n", err)
				return
			}

			if outputName == "" {
				outputName = packageName
			}

			outputPath := filepath.Join(outputDir, outputName+"."+outputType)

			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
				return
			}

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
		},
	}

	cmd.Flags().StringVarP(&packageName, "package", "p", "", "PyPI package name")
	cmd.Flags().StringVarP(&outputType, "type", "t", "md", "Output type: json or md")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "d", ".", "Output directory")
	cmd.Flags().StringVarP(&outputName, "output-name", "n", "", "Output file name (without extension)")
	cmd.Flags().StringArrayVarP(&excludePatterns, "exclude", "e", []string{}, "Patterns to exclude files (Use @ for file-based patterns, e.g., @.gitignore)")
	cmd.Flags().BoolVar(&includeGit, "include-git", false, "Include .git files and directories")
	cmd.Flags().BoolVar(&includeNonText, "include-non-text", false, "Include non-text files")
	cmd.Flags().BoolVar(&showExcluded, "show-excluded", false, "Show excluded files in project structure and shell commands")

	cmd.MarkFlagRequired("package")

	return cmd
}
