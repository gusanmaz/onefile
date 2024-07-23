package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gusanmaz/onefile/utils"
	"github.com/spf13/cobra"
)

func NewJSON2MDCmd() *cobra.Command {
	var jsonPath, outputPath string
	var includeGit, includeNonText bool
	var cmd = &cobra.Command{
		Use:   "json2md",
		Short: "Convert JSON to Markdown",
		Long:  `Convert a JSON file containing project structure to a Markdown file.`,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := ioutil.ReadFile(jsonPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
				return
			}

			var projectData utils.ProjectData
			err = json.Unmarshal(data, &projectData)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error unmarshaling JSON: %v\n", err)
				return
			}

			markdown := utils.GenerateMarkdown(projectData, includeGit, includeNonText)

			err = ioutil.WriteFile(outputPath, []byte(markdown), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing Markdown file: %v\n", err)
				return
			}

			fmt.Printf("Markdown generated at %s\n", outputPath)
		},
	}

	cmd.Flags().StringVarP(&jsonPath, "json", "j", "project_data.json", "Input JSON file")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "project_structure.md", "Output Markdown file")
	cmd.Flags().BoolVar(&includeGit, "include-git", false, "Include .git files and directories")
	cmd.Flags().BoolVar(&includeNonText, "include-non-text", false, "Include non-text files")

	return cmd
}
