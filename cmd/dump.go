package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gusanmaz/onefile/internal"
	"github.com/spf13/cobra"
)

func NewDumpCmd() *cobra.Command {
	var rootPath, outputPath, outputType string
	var excludePatterns []string
	var includeGit, includeNonText bool
	var cmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump a local project to JSON or Markdown",
		Long: `Dump a local project to JSON or Markdown.
Exclude patterns can be specified directly or by referencing a file with @.
Example: -e "*.go @.gitignore" -e "internal/extension_language_map.json go.mod go.sum"`,
		Run: func(cmd *cobra.Command, args []string) {
			// Process exclude patterns
			var processedPatterns []string
			for _, pattern := range excludePatterns {
				patterns := strings.Fields(pattern)
				processedPatterns = append(processedPatterns, patterns...)
			}

			parsedExcludePatterns, err := internal.ParsePatterns(processedPatterns)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing exclude patterns: %v\n", err)
				return
			}

			gitIgnore := internal.CreateGitIgnoreMatcher(parsedExcludePatterns)

			projectData, err := internal.DumpProject(rootPath, gitIgnore, includeGit, includeNonText)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error dumping project: %v\n", err)
				return
			}

			if outputPath == "" {
				outputPath = "project_data"
			}

			if outputType == "json" {
				err = internal.SaveAsJSON(projectData, outputPath+".json", includeGit, includeNonText)
			} else if outputType == "md" {
				err = internal.SaveAsMarkdown(projectData, outputPath+".md", includeGit, includeNonText)
			} else {
				fmt.Fprintf(os.Stderr, "Invalid output type. Use 'json' or 'md'\n")
				return
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error saving output: %v\n", err)
				return
			}

			fmt.Printf("Project dumped to %s.%s\n", outputPath, outputType)
		},
	}

	cmd.Flags().StringVarP(&rootPath, "path", "p", ".", "Root path of the project")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file name (without extension)")
	cmd.Flags().StringVarP(&outputType, "type", "t", "json", "Output type: json or md")
	cmd.Flags().StringArrayVarP(&excludePatterns, "exclude", "e", []string{}, "Patterns to exclude files (Use @ for file-based patterns, e.g., @.gitignore)")
	cmd.Flags().BoolVar(&includeGit, "include-git", false, "Include .git files and directories")
	cmd.Flags().BoolVar(&includeNonText, "include-non-text", false, "Include non-text files")

	return cmd
}
