package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gusanmaz/onefile/internal"
	"github.com/spf13/cobra"
)

func NewReconstructCmd() *cobra.Command {
	var jsonPath, outputPath string
	var cmd = &cobra.Command{
		Use:   "reconstruct",
		Short: "Reconstruct a project from JSON",
		Long:  `Reconstruct a project structure and file contents from a JSON file.`,
		Run: func(cmd *cobra.Command, args []string) {
			data, err := ioutil.ReadFile(jsonPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
				return
			}

			var projectData internal.ProjectData
			err = json.Unmarshal(data, &projectData)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error unmarshaling JSON: %v\n", err)
				return
			}

			err = internal.ReconstructProject(projectData, outputPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reconstructing project: %v\n", err)
				return
			}

			fmt.Printf("Project reconstructed in %s\n", outputPath)
		},
	}

	cmd.Flags().StringVarP(&jsonPath, "json", "j", "project_data.json", "Input JSON file")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "reconstructed_project", "Output directory")

	return cmd
}
