package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gusanmaz/onefile/internal"
	"github.com/spf13/cobra"
)

func NewPyPI2FileCmd() *cobra.Command {
	var packageName, outputType, outputDir, outputName string
	var cmd = &cobra.Command{
		Use:   "pypi2file",
		Short: "Fetch a PyPI package and save as JSON or Markdown",
		Long:  `Fetch a PyPI package and save its structure and contents as JSON or Markdown.`,
		Run: func(cmd *cobra.Command, args []string) {
			projectData, err := internal.FetchPyPIPackage(packageName)
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
		},
	}

	cmd.Flags().StringVarP(&packageName, "package", "p", "", "PyPI package name")
	cmd.Flags().StringVarP(&outputType, "type", "t", "md", "Output type: json or md")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "d", ".", "Output directory")
	cmd.Flags().StringVarP(&outputName, "output-name", "n", "", "Output file name (without extension)")

	cmd.MarkFlagRequired("package")

	return cmd
}
