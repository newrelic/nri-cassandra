/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExcludeWildcard(t *testing.T) {
	configYAML := `
exclude:
  metrics:
    - "*"
`
	f, err := ioutil.TempFile("", "cassandra_config.yml")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()

	_, err = f.WriteString(configYAML)
	require.NoError(t, err)

	err = f.Close()
	assert.NoError(t, err)

	os.Setenv(configPathEnv, f.Name())
	defer os.Unsetenv(configPathEnv)

	config, err := LoadConfig()
	assert.NoError(t, err)

	definitions := GetDefinitions(config)

	expected := Definitions{
		Common: commonDefinitions,
	}

	assert.Equal(t, expected, definitions)
}

func TestWithoutFilteringConfig(t *testing.T) {
	config, err := LoadConfig()
	assert.NoError(t, err)

	definitions := GetDefinitions(config)

	expected := Definitions{
		Common:              commonDefinitions,
		Metrics:             metricDefinitions,
		ColumnFamilyMetrics: columnFamilyDefinitions,
	}
	assert.Equal(t, expected, definitions)
}

func TestIncludeWildcard(t *testing.T) {
	configYAML := `
exclude:
  metrics:
    - "*"
include:
  metrics:
    - "*"
`
	f, err := ioutil.TempFile("", "cassandra_config.yml")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()

	_, err = f.WriteString(configYAML)
	require.NoError(t, err)

	err = f.Close()
	assert.NoError(t, err)

	os.Setenv(configPathEnv, f.Name())
	defer os.Unsetenv(configPathEnv)

	config, err := LoadConfig()
	assert.NoError(t, err)

	definitions := GetDefinitions(config)

	expected := Definitions{
		Common:              commonDefinitions,
		Metrics:             metricDefinitions,
		ColumnFamilyMetrics: columnFamilyDefinitions,
	}
	assert.Equal(t, expected, definitions)
}

func TestIncludeMetric(t *testing.T) {
	configYAML := `
exclude:
  metrics:
    - "*"
include:
  metrics:
    - client.connectedNativeClients
    - db.droppedRangeSliceMessagesPerSecond
    - db.tombstoneScannedHistogram999thPercentile
`

	f, err := ioutil.TempFile("", "cassandra_config.yml")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()

	_, err = f.WriteString(configYAML)
	require.NoError(t, err)
	assert.NoError(t, f.Close())

	os.Setenv(configPathEnv, f.Name())
	defer os.Unsetenv(configPathEnv)

	config, err := LoadConfig()
	assert.NoError(t, err)

	definitions := GetDefinitions(config)

	expected := Definitions{
		Common: commonDefinitions,
		Metrics: []Query{
			{
				MBean: "org.apache.cassandra.metrics:type=Client,name=connectedNativeClients",
				Attributes: []Attribute{
					{MBeanAttribute: "Value", Alias: "client.connectedNativeClients", MetricType: metric.GAUGE},
				},
			},
			{
				MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=RANGE_SLICE,name=Dropped",
				Attributes: []Attribute{
					{MBeanAttribute: "Count", Alias: "db.droppedRangeSliceMessagesPerSecond", MetricType: metric.RATE},
				},
			},
		},
		ColumnFamilyMetrics: []Query{
			{
				MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=TombstoneScannedHistogram",
				Attributes: []Attribute{
					{MBeanAttribute: "999thPercentile", Alias: "db.tombstoneScannedHistogram999thPercentile", MetricType: metric.GAUGE},
				},
			},
		},
	}
	assert.Equal(t, expected, definitions)
}
