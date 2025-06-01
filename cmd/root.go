package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "protolinter",
	Short: "Lint and analyze Protocol Buffer files",
	Long: `The 'protolinter' command-line tool helps developers review and analyze
Protocol Buffer files for coding conventions, best practices, and standards.
It ensures your Protocol Buffer files are well-structured, consistent, and adhere
to recommended guidelines.
Configuration:
  Customize the tool using the configuration file (default filename: '.protolinter.yaml').
  This YAML configuration lets you exclude checks and descriptors,
  tailoring the analysis to match your project's needs.
See '.protolinter.example.yaml' for a sample configuration.`,
	Version: "1.1.6",
}

// Execute runs the root command.
// Everything starts here.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
