package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gusanmaz/onefile/internal"
	"github.com/spf13/cobra"
)

func NewDumpCmd() *cobra.Command {
	var rootPath, outputPath, outputType string
	var includePatterns, excludePatterns []string
	var includeGit, includeNonText bool
	var cmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump a local project to JSON or Markdown",
		Run: func(cmd *cobra.Command, args []string) {
			// Ensure include patterns have a leading dot for file extensions
			for i, pattern := range includePatterns {
				if pattern[0] != '.' && pattern[0] != '*' {
					includePatterns[i] = "." + pattern
				}
			}

			projectData, err := internal.DumpProject(rootPath, includePatterns, excludePatterns, includeGit, includeNonText)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error dumping project: %v\n", err)
				return
			}

			// Set default output name if not provided
			if outputPath == "" {
				outputPath = "project_data"
			}

			// Add file extension if not present
			if filepath.Ext(outputPath) == "" {
				outputPath += "." + outputType
			}

			// Create output directory if it doesn't exist
			outputDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
				return
			}
			fmt.Printf("Created output directory: %s\n", outputDir)

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

			fmt.Printf("Project dumped to %s\n", outputPath)
			fmt.Println("Shell commands to recreate the project structure:")
			fmt.Println(internal.GenerateShellCommands(projectData))
		},
	}

	cmd.Flags().StringVarP(&rootPath, "path", "p", ".", "Root path of the project")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file name (without extension)")
	cmd.Flags().StringVarP(&outputType, "type", "t", "json", "Output type: json or md")
	cmd.Flags().StringSliceVarP(&includePatterns, "include", "i", []string{"*"}, "Patterns to include files (space-separated)")
	cmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Patterns to exclude files (space-separated)")
	cmd.Flags().BoolVar(&includeGit, "include-git", false, "Include .git files and directories")
	cmd.Flags().BoolVar(&includeNonText, "include-non-text", false, "Include non-text files")

	return cmd
}
