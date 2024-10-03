/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/persist"
	"github.com/stretchr/testify/assert"
)

func TestPopulateMetrics(t *testing.T) {
	var rawMetrics = map[string]interface{}{
		"raw_metric_1,attr=Value": 1,
		"raw_metric_2,attr=Value": 2,
		"raw_metric_3,attr=Value": "foo",
	}

	var metricDefinition = []Query{
		{
			MBean: "raw_metric_1",
			Attributes: []Attribute{
				{MBeanAttribute: "Value", Alias: "rawMetric1", MetricType: metric.GAUGE},
			},
		},
		{
			MBean: "raw_metric_2",
			Attributes: []Attribute{
				{MBeanAttribute: "Value", Alias: "rawMetric2", MetricType: metric.GAUGE},
			},
		},
		{
			MBean: "raw_metric_3",
			Attributes: []Attribute{
				{MBeanAttribute: "Value", Alias: "rawMetric3", MetricType: metric.ATTRIBUTE},
			},
		},
		{
			MBean: "raw_metric_4",
			Attributes: []Attribute{
				{MBeanAttribute: "Value", Alias: "unknownMetric", MetricType: metric.GAUGE},
			},
		},
	}

	s := metric.NewSet("eventType", persist.NewInMemoryStore())
	populateMetrics(s, rawMetrics, metricDefinition)

	sample := s.Metrics

	assert.Equal(t, 1.0, sample["rawMetric1"])
	assert.Equal(t, 2.0, sample["rawMetric2"])
	assert.Equal(t, "foo", sample["rawMetric3"])
	assert.Nil(t, sample["unknownMetric"])
}

func TestPopulateInventory(t *testing.T) {
	var rawInventory = inventory.Item{
		"key_1":                 1,
		"key_2":                 2,
		"key_3":                 "foo",
		"key_4":                 map[interface{}]interface{}{"test": 2},
		"my_important_password": "12345",
		"key_6":                 map[interface{}]interface{}{"otherImportantPassword": 54321},
	}

	i := inventory.New()
	assert.NoError(t, populateInventory(i, rawInventory))

	expectedItems := inventory.Items{
		"key_1":                 {"value": 1},
		"key_2":                 {"value": 2},
		"key_3":                 {"value": "foo"},
		"key_4":                 {"test": 2},
		"my_important_password": {"value": "(omitted value)"},
		"key_6":                 {"otherImportantPassword": "(omitted value)"},
	}

	assert.Equal(t, expectedItems, i.Items())
}
