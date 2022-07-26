package main

import (
	"github.com/newrelic/infra-integrations-sdk/log"
	"gopkg.in/yaml.v2"
	"os"
)

const (
	configPathEnv = "CONFIG_PATH"
)

type Config struct {
	Include QueryFilterConfig `yaml:"include"`
	Exclude QueryFilterConfig `yaml:"exclude"`
}

func LoadConfig() (Config, error) {
	configFile := os.Getenv(configPathEnv)
	if configFile == "" {
		log.Debug("No extra config provided")

		return Config{}, nil
	}

	log.Debug("Loading config file from: %s", configFile)

	content, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	config := Config{}

	err = yaml.Unmarshal(content, &config)
	return config, err
}

// Bypass returns true if query is not filtered by the configuration.
func (f Config) BypassFiltering(attribute Attribute) bool {
	return !f.Exclude.match(attribute) || f.Include.match(attribute)
}

// FilterConfig YAML Struct
type QueryFilterConfig struct {
	MetricNames []string `yaml:"metrics"`
}

// match returns true if the entry contains the fields specified by the filter configuration.
func (f QueryFilterConfig) match(attribute Attribute) bool {
	for _, metricName := range f.MetricNames {
		if metricName == "*" || metricName == attribute.Alias {
			return true
		}
	}
	return false
}
