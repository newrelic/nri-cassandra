package main

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

func TestPopulateMetrics(t *testing.T) {
	var rawMetrics = map[string]interface{}{
		"raw_metric_1": 1,
		"raw_metric_2": 2,
		"raw_metric_3": "foo",
	}

	functionSource := func(a map[string]interface{}) (float64, bool) {
		return float64(a["raw_metric_1"].(int) + a["raw_metric_2"].(int)), true
	}

	var metricDefinition = map[string][]interface{}{
		"rawMetric1":     {"raw_metric_1", metric.GAUGE},
		"rawMetric2":     {"raw_metric_2", metric.GAUGE},
		"rawMetric3":     {"raw_metric_3", metric.ATTRIBUTE},
		"unknownMetric":  {"raw_metric_4", metric.GAUGE},
		"badRawSource":   {10, metric.GAUGE},
		"functionSource": {functionSource, metric.GAUGE},
	}

	s := metric.NewSet("eventType", persist.NewInMemoryStore())
	populateMetrics(s, rawMetrics, metricDefinition)

	sample := s.Metrics

	assert.Equal(t, 1.0, sample["rawMetric1"])
	assert.Equal(t, 2.0, sample["rawMetric2"])
	assert.Equal(t, "foo", sample["rawMetric3"])
	assert.Nil(t, sample["unknownMetric"])
	assert.Nil(t, sample["badRawSource"])
	assert.Equal(t, 3.0, sample["functionSource"])
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
