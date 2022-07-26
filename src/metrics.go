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

// getMetrics will gather all node and keyspace level metrics and return them as two maps
// The main metrics map will contain all the keys got from JMX and the keyspace metrics map
// Will contain maps for each <keyspace>.<columnFamily> found while inspecting JMX metrics.
func getMetrics(client *gojmx.Client, queryConfig []Query) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	for _, query := range queryConfig {
		results, err := client.QueryMBeanAttributes(query.MBean, query.GetAttributeNames()...)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Error querying %s: %v", query.MBean, jmxErr)
				continue
			}
			return nil, fmt.Errorf("fatal jmx error while querying: %q: %w", query.MBean, err)
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
func getColumnFamilyMetrics(client *gojmx.Client, queryConfig []Query) (map[string]map[string]interface{}, error) {
	columnFamilyMetrics := make(map[string]map[string]interface{})

	columnFamilyQueryConfig, err := getColumnFamilyQueries(client, queryConfig)
	if err != nil {
		return nil, fmt.Errorf("fatal jmx error while querying mBean names, error: %w", err)
	}

	for _, query := range columnFamilyQueryConfig {
		results, err := client.GetMBeanAttributes(query.MBean, query.GetAttributeNames()...)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Error getting attributes for mBeanName %s: %v", query.MBean, jmxErr)
				continue
			}
			return nil, fmt.Errorf("fatal jmx error while getting attributes for mBean: %q: %w", query.MBean, err)
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
func getColumnFamilyQueries(client *gojmx.Client, queryConfig []Query) ([]Query, error) {
	var result []Query
	visitedColumnFamilies := make(map[string]struct{})

	for _, query := range queryConfig {
		mBeanNames, err := client.QueryMBeanNames(query.MBean)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Error querying mBean name %s: %v", query.MBean, jmxErr)
				continue
			}
			return nil, fmt.Errorf("fatal jmx error while querying: %q: %w", query.MBean, err)
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

			query.MBean = mBeanName
			result = append(result, query)
		}
	}
	return result, nil
}

func populateMetrics(s *metric.Set, metrics map[string]interface{}, queryConfig []Query) {
	notFoundMetrics := make([]string, 0)

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
				log.Error("setting value: %s", err)
				continue
			}
		}
	}
	if len(notFoundMetrics) > 0 {
		log.Debug("Can't find raw metrics in results for keys: %v", notFoundMetrics)
	}
}

func populateAttributes(s *metric.Set, metrics map[string]interface{}, sampleAttributes []SampleAttribute) {
	for _, sampleAttr := range sampleAttributes {
		if rawMetric, found := metrics[sampleAttr.Key]; found {
			s.SetMetric(sampleAttr.Alias, rawMetric, sampleAttr.MetricType)
		}
	}
}
