package config

import (
	"os"

	"github.com/spf13/viper"
)

// DefaultConfigName - default configuration file name.
const DefaultConfigName = ".protolinter.yaml"

// LoadConfig loads the configuration from the specified file using Viper.
// If the filename is empty, it loads the default configuration file.
func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		filename = DefaultConfigName
	}

	if _, err := os.Open(filename); os.IsNotExist(err) {
		return nil, nil
	}

	viper.SetConfigName(filename)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var container Config

	err = viper.Unmarshal(&container)
	if err != nil {
		return nil, err
	}

	result := &container
	result.fillInnerData()

	return result, nil
}

// GetVerboseMode returns the value of VerboseMode from the Config struct.
// If the Config is nil or VerboseMode is not set, it returns false.
func (cfg *Config) GetVerboseMode() bool {
	if cfg != nil {
		return cfg.VerboseMode
	}

	return false
}

// GetOmitCoordinates returns the value of OmitCoordinates from the Config struct.
// If the Config is nil or OmitCoordinates is not set, it returns false.
func (cfg *Config) GetOmitCoordinates() bool {
	if cfg != nil {
		return cfg.OmitCoordinates
	}

	return false
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
