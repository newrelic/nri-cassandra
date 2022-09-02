/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExcludeWildcard(t *testing.T) {
	configYAML := `
exclude:
  - "*"
`

	config, err := LoadFilteringConfig(configYAML)
	assert.NoError(t, err)

	definitions := NewDefinitions()
	definitions.Filter(config)

	expected := Definitions{}

	assert.Equal(t, expected, definitions)
}

func TestWithoutFilteringConfig(t *testing.T) {
	emptyCfg := ""
	config, err := LoadFilteringConfig(emptyCfg)
	assert.NoError(t, err)

	definitions := NewDefinitions()
	definitions.Filter(config)

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
  - "*"
include:
  - "*"
`

	config, err := LoadFilteringConfig(configYAML)
	assert.NoError(t, err)

	definitions := NewDefinitions()
	definitions.Filter(config)

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
  - "*"
include:
  - client.connectedNativeClients
  - db.droppedRangeSliceMessagesPerSecond
  - db.tombstoneScannedHistogram999thPercentile
`

	config, err := LoadFilteringConfig(configYAML)
	assert.NoError(t, err)

	definitions := NewDefinitions()
	definitions.Filter(config)

	expected := Definitions{
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
