/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"

	"github.com/newrelic/infra-integrations-sdk/log"
	"gopkg.in/yaml.v2"
)

const (
	// configPathEnv is the environment variable used by the infrastructure-agent to pass extra configuration
	// to integrations.
	configPathEnv = "CONFIG_PATH"
)

// Config is the extra configuration that can be received by nri-cassandra from infrastructure agent.
type Config struct {
	// Include will define which cassandra metric definitions will be included.
	Include QueryFilterConfig `yaml:"include"`
	// Exclude specifies which nri-cassandra metric definitions will be excluded from collection.
	Exclude QueryFilterConfig `yaml:"exclude"`
}

// QueryFilterConfig YAML Struct to unmarshal Query filter configuration.
type QueryFilterConfig struct {
	MetricNames []string `yaml:"metrics"`
}

// match returns true if the Attribute Alias matches the filter configuration.
func (f QueryFilterConfig) match(attribute Attribute) bool {
	for _, metricName := range f.MetricNames {
		if metricName == "*" || metricName == attribute.Alias {
			return true
		}
	}
	return false
}

// IsFiltered returns true if query is filtered by the configuration.
// Include filters have precedence over Exclude filters.
func (f Config) IsFiltered(attribute Attribute) bool {
	return f.Exclude.match(attribute) && !f.Include.match(attribute)
}

// LoadConfig will check if the configPathEnv file variable is set and will try to
// unmarshal it's content into a Config structure.
func LoadConfig() (Config, error) {
	result := Config{}

	// Extra configuration will be enabled just when "metrics" mode is
	// explicitly set to avoid colliding with inventory CONFIG_PATH.
	if !args.Metrics {
		return result, nil
	}

	configFile := os.Getenv(configPathEnv)
	if configFile == "" {
		log.Debug("No extra config provided")

		return result, nil
	}

	log.Debug("Loading config file from file: %s", configFile)

	content, err := os.ReadFile(configFile)
	if err != nil {
		return result, err
	}

	err = yaml.Unmarshal(content, &result)
	return result, err
}
