/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"regexp"

	"github.com/newrelic/nrjmx/gojmx"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/log"
)

var (
	// columnFamilyRegex matches the keyspace name and the scope.
	columnFamilyRegex = regexp.MustCompile("keyspace=(.*),scope=(.*?),")

	percentileRegex = regexp.MustCompile("attr=.*Percentile")

	// filteredKeyspace set used to match internal keyspace that should not be reported.
	filteredKeyspace = map[string]struct{}{
		"OpsCenter":          {},
		"system":             {},
		"system_auth":        {},
		"system_distributed": {},
		"system_schema":      {},
		"system_traces":      {},
	}
)

// getMetrics will gather all node level metrics and return them as a map.
func getMetrics(client *gojmx.Client, queryConfig []Query) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	for _, query := range queryConfig {
		attrNames := query.GetAttributeNames()

		results, err := client.QueryMBeanAttributes(query.MBean, attrNames...)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Failed to get attributes: %s, for mBeanName %s: %v", attrNames, query.MBean, jmxErr)
				continue
			}
			return nil, fmt.Errorf("failed to perform query: %q: for attributes: %s error: %w", query.MBean, attrNames, err)
		}

		for _, jmxAttr := range results {
			if jmxAttr.ResponseType == gojmx.ResponseTypeErr {
				log.Debug("Failed to retrieve attribute for query: %s status: %s", jmxAttr.Name, jmxAttr.StatusMsg)
				continue
			}

			metrics[jmxAttr.Name] = jmxAttr.GetValue()
		}
	}

	return metrics, nil
}

// getMetrics will gather all keyspace level metrics and return them as a map that
// will contain maps for each <keyspace>.<columnFamily> found while inspecting JMX metrics.
func getColumnFamilyMetrics(client *gojmx.Client, queryConfig []Query) (map[string]map[string]interface{}, error) {
	columnFamilyMetrics := make(map[string]map[string]interface{})

	columnFamilyQueryConfig, err := getColumnFamilyQueries(client, queryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch 'column-family' metrics: %w", err)
	}

	for _, query := range columnFamilyQueryConfig {
		attrNames := query.GetAttributeNames()

		results, err := client.GetMBeanAttributes(query.MBean, attrNames...)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Failed to get 'column-family' attributes: %s, for mBeanName %s: %v", attrNames, query.MBean, jmxErr)
				continue
			}
			return nil, fmt.Errorf("failed to fetch 'column-family' metrics, query: %q: attributes: %s error: %w", query.MBean, attrNames, err)
		}

		for _, jmxAttr := range results {
			if jmxAttr.ResponseType == gojmx.ResponseTypeErr {
				log.Debug("Failed to retrieve 'column-family' attribute for query: %s status: %s", jmxAttr.Name, jmxAttr.StatusMsg)
				continue
			}

			matches := columnFamilyRegex.FindStringSubmatch(jmxAttr.Name)
			key := columnFamilyRegex.ReplaceAllString(jmxAttr.Name, "")

			columnFamily := matches[2]
			keyspace := matches[1]
			eventKey := keyspace + "." + columnFamily

			_, ok := columnFamilyMetrics[eventKey]
			if !ok {
				columnFamilyMetrics[eventKey] = make(map[string]interface{})
				columnFamilyMetrics[eventKey]["keyspace"] = keyspace
				columnFamilyMetrics[eventKey]["columnFamily"] = columnFamily
				columnFamilyMetrics[eventKey]["keyspaceAndColumnFamily"] = eventKey
			}
			columnFamilyMetrics[eventKey][key] = jmxAttr.GetValue()
		}
	}

	return columnFamilyMetrics, nil
}

// getColumnFamilyMBeanNames will query the MBeanNames patterns to expand the wildcards ('*').
// gojmx.QueryMBeanNames call is cheaper than fetching altogether the MBeanAttributes values.
// This way we can apply filtering (e.g. column_families_limit) before querying the actual attribute values.
func getColumnFamilyQueries(client *gojmx.Client, queryConfig []Query) ([]Query, error) {
	var result []Query
	visitedColumnFamilies := make(map[string]struct{})

	for _, query := range queryConfig {
		mBeanNames, err := client.QueryMBeanNames(query.MBean)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Failed to querying mBeanNames %s: %v", query.MBean, jmxErr)
				continue
			}
			return nil, fmt.Errorf("cannot retrieve mBeanNames for query: %q, error: %w", query.MBean, err)
		}

		for _, mBeanName := range mBeanNames {
			matches := columnFamilyRegex.FindStringSubmatch(mBeanName)

			keyspace, columnFamily := matches[1], matches[2]

			eventKey := keyspace + "." + columnFamily

			// Discard internal keyspaces.
			_, isFiltered := filteredKeyspace[keyspace]
			if isFiltered {
				continue
			}

			// Limit to maximum args.ColumnFamiliesLimit.
			_, found := visitedColumnFamilies[eventKey]
			if !found {
				if len(visitedColumnFamilies) < args.ColumnFamiliesLimit {
					visitedColumnFamilies[eventKey] = struct{}{}
				} else {
					log.Warn("Skipping column family %s due to limit reached. Current limit set to %d",
						columnFamily, args.ColumnFamiliesLimit)
					continue
				}
			}

			query.MBean = mBeanName
			result = append(result, query)
		}
	}
	return result, nil
}

// populateMetrics will use the rawMetrics received from the JMXClient and store them into a nr-infra-sdk metric object.
func populateMetrics(s *metric.Set, metrics map[string]interface{}, queryConfig []Query) {
	var notFoundMetrics []string

	for _, query := range queryConfig {
		for _, attr := range query.Attributes {
			rawSource := fmt.Sprintf("%s,attr=%s", query.MBean, attr.MBeanAttribute)
			rawSource = columnFamilyRegex.ReplaceAllString(rawSource, "")

			metricType := attr.MetricType

			var rawMetric interface{}
			var ok bool

			rawMetric, ok = metrics[rawSource]

			if rawMetric != nil && percentileRegex.MatchString(rawSource) {
				// Convert percentiles from microseconds to milliseconds
				rawMetric = rawMetric.(float64) / 1000.0
			}

			if !ok {
				notFoundMetrics = append(notFoundMetrics, attr.Alias)

				continue
			}

			err := s.SetMetric(attr.Alias, rawMetric, metricType)
			if err != nil {
				log.Debug("Failed to set metric value: %v", err)

				continue
			}
		}
	}
	if len(notFoundMetrics) > 0 {
		log.Debug("Can't find raw metrics in results for keys: %v", notFoundMetrics)
	}
}

// populateAttributes will use the rawMetrics received from the JMXClient to get the attributes that will make
// the metrics unique.
func populateAttributes(s *metric.Set, metrics map[string]interface{}, sampleAttributes []SampleAttribute) {
	for _, sampleAttr := range sampleAttributes {
		rawMetric, found := metrics[sampleAttr.Key]
		if found {
			log.Debug("Can't find raw metrics in results for attribute: %s", sampleAttr.Key)
			continue
		}

		err := s.SetMetric(sampleAttr.Alias, rawMetric, sampleAttr.MetricType)
		if err != nil {
			log.Debug("Failed to set attribute: %s: %v", sampleAttr.Key, err)
			continue
		}
	}
}
