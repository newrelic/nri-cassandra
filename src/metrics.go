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

// getMetrics will gather all node and keyspace level metrics and return them as two maps
// The main metrics map will contain all the keys got from JMX and the keyspace metrics map
// Will contain maps for each <keyspace>.<columnFamily> found while inspecting JMX metrics.
func getMetrics(client *gojmx.Client) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	for _, query := range jmxMetricsPatterns {
		results, err := client.QueryMBeanAttributes(query)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Error querying %s: %v", query, jmxErr)
				continue
			}
			return nil, fmt.Errorf("fatal jmx error while querying: %q: %w", query, err)
		}

		for _, jmxAttr := range results {
			if jmxAttr.ResponseType == gojmx.ResponseTypeErr {
				log.Debug("Failed to process attribute for query: %s status: %s", jmxAttr.Name, jmxAttr.StatusMsg)
				continue
			}

			metrics[jmxAttr.Name] = jmxAttr.GetValue()
		}
	}

	return metrics, nil
}

// getMetrics will gather all node and keyspace level metrics and return them as two maps
// The main metrics map will contain all the keys got from JMX and the keyspace metrics map
// Will contain maps for each <keyspace>.<columnFamily> found while inspecting JMX metrics.
func getColumnFamilyMetrics(client *gojmx.Client) (map[string]map[string]interface{}, error) {
	columnFamilyMetrics := make(map[string]map[string]interface{})

	mBeanNames, err := getColumnFamilyMBeanNames(client)
	if err != nil {
		return nil, fmt.Errorf("fatal jmx error while querying mBean names, error: %w", err)
	}

	for _, mBeanName := range mBeanNames {
		attrNames, err := client.GetMBeanAttributeNames(mBeanName)
		if err != nil {
			log.Debug("Error getting attribute names for mBeanName %s: %v", mBeanName, err)
			continue
		}

		results, err := client.GetMBeanAttributes(mBeanName, attrNames...)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Error getting attributes for mBeanName %s: %v", mBeanName, jmxErr)
				continue
			}
			return nil, fmt.Errorf("fatal jmx error while getting attributes for mBean: %q: %w", mBeanName, err)
		}

		for _, jmxAttr := range results {
			if jmxAttr.ResponseType == gojmx.ResponseTypeErr {
				log.Debug("Failed to process attribute for query: %s status: %s", jmxAttr.Name, jmxAttr.StatusMsg)
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

// getColumnFamilyMBeanNames will fetch all the mBean names for columnFamily metrics.
// Each MBean name will be used to query the metric values.
// QueryMBeanNames call is cheaper, we want to apply the filtering before querying metrics values.
func getColumnFamilyMBeanNames(client *gojmx.Client) ([]string, error) {
	var result []string
	visitedColumnFamilies := make(map[string]struct{})

	for _, pattern := range jmxColumnFamilyMetricsPatterns {
		mBeanNames, err := client.QueryMBeanNames(pattern)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Error querying mBean name %s: %v", pattern, jmxErr)
				continue
			}
			return nil, fmt.Errorf("fatal jmx error while querying: %q: %w", pattern, err)
		}

		for _, mBeanName := range mBeanNames {
			matches := columnFamilyRegex.FindStringSubmatch(mBeanName)

			columnFamily := matches[2]
			keyspace := matches[1]
			eventKey := keyspace + "." + columnFamily

			_, isFiltered := filteredKeyspace[keyspace]
			if isFiltered {
				continue
			}

			// limit to maximum args.ColumnFamiliesLimit.
			_, found := visitedColumnFamilies[eventKey]
			if !found {
				if len(visitedColumnFamilies) < args.ColumnFamiliesLimit {
					visitedColumnFamilies[eventKey] = struct{}{}
				} else {
					continue
				}
			}

			result = append(result, mBeanName)
		}
	}
	return result, nil
}

func populateMetrics(s *metric.Set, metrics map[string]interface{}, definition map[string][]interface{}) {
	notFoundMetrics := make([]string, 0)
	for metricName, metricConf := range definition {
		rawSource := metricConf[0]
		metricType := metricConf[1].(metric.SourceType)

		var rawMetric interface{}
		var ok bool

		switch source := rawSource.(type) {
		case string:
			rawMetric, ok = metrics[source]
			percentileRe, err := regexp.Compile("attr=.*Percentile")
			if err != nil {
				continue
			}
			if rawMetric != nil && percentileRe.MatchString(source) {
				// Convert percentiles from microseconds to milliseconds
				rawMetric = rawMetric.(float64) / 1000.0
			}
		case func(map[string]interface{}) (float64, bool):
			rawMetric, ok = source(metrics)
		default:
			log.Debug("Invalid raw source metric for %s", metricName)
			continue
		}

		if !ok {
			notFoundMetrics = append(notFoundMetrics, metricName)

			continue
		}

		err := s.SetMetric(metricName, rawMetric, metricType)
		if err != nil {
			log.Error("setting value: %s", err)
			continue
		}
	}
	if len(notFoundMetrics) > 0 {
		log.Debug("Can't find raw metrics in results for keys: %v", notFoundMetrics)
	}
}
