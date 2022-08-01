/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import "github.com/newrelic/infra-integrations-sdk/data/metric"

// Query defines a JMX query that has to be performed. Multiple JMX attributes can be received through a single Query.
// Each JMX Attribute maps to a single NR metric. We set an Alias to attribute to define the name of the metric in NR.
type Query struct {
	MBean      string      `yaml:"mbean"`
	Attributes []Attribute `yaml:"attributes"`
}

// GetAttributeNames will iterate over the attributes to retrieve a slice with only the attribute names.
// This is handy when performing the JMX query.
func (q *Query) GetAttributeNames() []string {
	attrs := make([]string, len(q.Attributes))

	for i := range q.Attributes {
		attrs[i] = q.Attributes[i].MBeanAttribute
	}
	return attrs
}

// Attribute maps the JMX Attribute to the NR metric. Alias defines the name of the metric in NR.
type Attribute struct {
	MBeanAttribute string            `yaml:"mbean_attribute"`
	Alias          string            `yaml:"alias"`
	MetricType     metric.SourceType `yaml:"metric_type"`
}

func (a *Attribute) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var details map[string]interface{}
	if err := unmarshal(&details); err != nil {
		return err
	}
	sourceType := metric.GAUGE
	if metricTypeField, ok := details["metric_type"]; ok {
		metricTypeName, ok := metricTypeField.(string)
		if ok {
			metricType, found := metric.SourcesNameToType[metricTypeName]
			if found {
				sourceType = metricType
			}
		}
	}
	a.MetricType = sourceType
	return nil
}
