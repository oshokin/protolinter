// Package config provides functionality for loading and working with
// the configuration settings.
package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigName - default configuration file name.
	DefaultConfigName = ".protolinter.yaml"

	goModPath                    = "go.mod"
	moduleNamePatternGroupsCount = 2
)

var moduleNamePattern = regexp.MustCompile(`\s*module\s+(?P<module>\S+)`)

// LoadConfig loads the configuration from the specified file using Viper.
// If the filename is empty, it loads the default configuration file.
func LoadConfig(filename, githubURL string, printAllDescriptors bool) (*Config, error) {
	if filename == "" {
		filename = DefaultConfigName
	}

	var (
		configFileNotFound bool
		result             Config
	)

	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		configFileNotFound = os.IsNotExist(err) || errors.As(err, &viper.ConfigFileNotFoundError{})
		if !configFileNotFound {
			return nil, err
		}
	}

	if !configFileNotFound {
		if err := viper.Unmarshal(&result); err != nil {
			return nil, err
		}
	}

	moduleName, err := getModuleName()
	if err != nil {
		return nil, err
	}

	result.moduleName = moduleName
	result.githubURL = githubURL
	result.printAllDescriptors = printAllDescriptors

	result.fillInnerData()

	return &result, nil
}

// MarshalConfig converts the given Config struct to its YAML representation.
func MarshalConfig(v *Config) (string, error) {
	if v == nil {
		return "", nil
	}

	result, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// GetVerboseMode returns the value of VerboseMode from the Config struct.
// If the Config is nil or VerboseMode is not set, it returns false.
func (cfg *Config) GetVerboseMode() bool {
	return cfg != nil && cfg.VerboseMode
}

// GetExcludedChecks returns the list of excluded checks from the Config struct.
// If the Config is nil or ExcludedChecks is not set, it returns an empty slice.
func (cfg *Config) GetExcludedChecks() []string {
	if cfg != nil {
		return cfg.ExcludedChecks
	}

	return nil
}

// GetExcludedDescriptors returns the list of excluded descriptors from the Config struct.
// If the Config is nil or ExcludedDescriptors is not set, it returns an empty slice.
func (cfg *Config) GetExcludedDescriptors() []string {
	if cfg != nil {
		return cfg.ExcludedDescriptors
	}

	return nil
}

// GetModuleName returns the module name extracted from the go.mod file.
// If the Config is nil, it returns an empty string.
func (cfg *Config) GetModuleName() string {
	if cfg != nil {
		return cfg.moduleName
	}

	return ""
}

// GetGitHubURL returns the custom file repository URL set in the configuration.
// If the Config is nil or GitHubURL is not set, it returns an empty string.
func (cfg *Config) GetGitHubURL() string {
	if cfg != nil {
		return cfg.githubURL
	}

	return ""
}

// GetPrintAllDescriptors returns the value of printAllDescriptors from the Config struct.
// If the Config is nil, it returns false.
func (cfg *Config) GetPrintAllDescriptors() bool {
	if cfg != nil {
		return cfg.printAllDescriptors
	}

	return false
}

// IsCheckExcluded checks if a specific check is excluded based on the configuration.
func (cfg *Config) IsCheckExcluded(name string) bool {
	if cfg == nil {
		return false
	}

	_, isExcluded := cfg.excludedChecksMap[name]

	return isExcluded
}

func (cfg *Config) fillInnerData() {
	if cfg == nil {
		return
	}

	checks := cfg.GetExcludedChecks()
	if len(checks) == 0 {
		return
	}

	checksMap := make(map[string]struct{}, len(checks))
	for _, v := range checks {
		checksMap[v] = struct{}{}
	}

	cfg.excludedChecksMap = checksMap
}

func getModuleName() (string, error) {
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return "", nil
	}

	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s file: %w", goModPath, err)
	}

	match := moduleNamePattern.FindSubmatch(goModContent)
	if len(match) < moduleNamePatternGroupsCount {
		return "", fmt.Errorf("module name not found in %s", goModPath)
	}

	return string(match[1]), nil
}
