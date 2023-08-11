package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "protolinter",
	Short: "A tool to lint and analyze Protocol Buffer files",
	Long: `The 'protolinter' command-line tool empowers developers to lint and analyze
Protocol Buffer files for compliance with coding conventions, best practices,
and standards. It assists in ensuring that your Protocol Buffer files are well-formed,
consistent, and follow recommended guidelines.
Configuration:
  The tool supports customization through a '.protolinter.yaml' configuration file.
  This YAML file can be used to define excluded checks and descriptors, allowing you
  to fine-tune the analysis to your project's needs.
Example '.protolinter.yaml' configuration can be found in .protolinter.example.yaml`,
	Version: "1.0.0",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
