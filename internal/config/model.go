package config

// Config represents the configuration read from the file.
type Config struct {
	// VerboseMode specifies whether to show verbose messages, such as when downloading dependencies.
	VerboseMode bool `mapstructure:"verbose_mode" yaml:"verbose_mode"`
	// ExcludedChecks is a list of checks that should be excluded from analysis.
	ExcludedChecks []string `mapstructure:"excluded_checks" yaml:"excluded_checks"`
	// ExcludedDescriptors is a list of full protopaths that should be excluded from analysis.
	ExcludedDescriptors []string `mapstructure:"excluded_descriptors" yaml:"excluded_descriptors"`

	// The name of the Go module extracted from go.mod file.
	moduleName string
	// The custom file repository URL for replacing github.com links.
	githubURL string
	// Indicates whether to include all descriptors, not just erroneous ones.
	printAllDescriptors bool
	// A map for efficient verification of excluded checks.
	excludedChecksMap map[string]struct{}
}
