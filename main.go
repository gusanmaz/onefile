package main

import (
	"fmt"
	"os"

	"github.com/gusanmaz/onefile/cmd"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "onefile",
		Short: "A tool for project file management and repository fetching",
		Long: `onefile is a versatile command-line tool that allows you to:
- Dump local project structures to JSON or Markdown
- Reconstruct projects from JSON
- Convert JSON project representations to Markdown
- Fetch GitHub repositories and save them as JSON or Markdown
- Fetch PyPI packages and save them as JSON or Markdown`,
	}

	rootCmd.AddCommand(
		cmd.NewDumpCmd(),
		cmd.NewReconstructCmd(),
		cmd.NewJSON2MDCmd(),
		cmd.NewGitHub2FileCmd(),
		cmd.NewPyPI2FileCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
