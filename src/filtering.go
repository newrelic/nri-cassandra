/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import "gopkg.in/yaml.v3"

// FilteringConfig specifies which metrics should be included/excluded from reporting.
type FilteringConfig struct {
	// Include will define which cassandra metric definitions will be included.
	Include MetricNameList `yaml:"include"`
	// Exclude specifies which nri-cassandra metric definitions will be excluded from collection.
	Exclude MetricNameList `yaml:"exclude"`
}

// IsFiltered returns true if query is filtered by the configuration.
// Include filters have precedence over Exclude filters.
func (f FilteringConfig) IsFiltered(attribute Attribute) bool {
	return f.Exclude.contains(attribute) && !f.Include.contains(attribute)
}

// LoadFilteringConfig unmarshal the YAML format filtering configuration.
func LoadFilteringConfig(metricsCfg string) (FilteringConfig, error) {
	result := FilteringConfig{}

	if metricsCfg == "" {
		return result, nil
	}

	err := yaml.Unmarshal([]byte(metricsCfg), &result)
	return result, err
}

// MetricNameList contains a list of metric names.
type MetricNameList []string

// contains returns true if the Attribute Alias is found in the MetricNameList.
func (f MetricNameList) contains(attribute Attribute) bool {
	for _, metricName := range f {
		if metricName == "*" || metricName == attribute.Alias {
			return true
		}
	}
	return false
}
