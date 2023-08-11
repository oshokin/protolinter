package config

// Config represents the configuration read from the file.
type Config struct {
	// VerboseMode specifies whether to show verbose messages, such as when downloading dependencies.
	VerboseMode bool `mapstructure:"verbose_mode"`
	// OmitCoordinates specifies whether to omit source file coordinates from error messages.
	OmitCoordinates bool `mapstructure:"omit_coordinates"`
	// ExcludedChecks is a list of checks that should be excluded from analysis.
	ExcludedChecks []string `mapstructure:"excluded_checks"`
	// ExcludedDescriptors is a list of full protopaths that should be excluded from analysis.
	ExcludedDescriptors []string `mapstructure:"excluded_descriptors"`
	excludedChecksMap   map[string]struct{}
}
