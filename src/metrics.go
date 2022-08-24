package main

import (
	"fmt"
	"regexp"

	"github.com/newrelic/nrjmx/gojmx"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/log"
)

// getMetrics will gather all node and keyspace level metrics and return them as two maps
// The main metrics map will contain all the keys got from JMX and the keyspace metrics map
// Will contain maps for each <keyspace>.<columnFamily> found while inspecting JMX metrics.
func getMetrics(client *gojmx.Client) (map[string]interface{}, map[string]map[string]interface{}, error) {
	internalKeyspaces := map[string]struct{}{
		"OpsCenter":          {},
		"system":             {},
		"system_auth":        {},
		"system_distributed": {},
		"system_schema":      {},
		"system_traces":      {},
	}
	metrics := make(map[string]interface{})
	columnFamilyMetrics := make(map[string]map[string]interface{})
	visitedColumnFamilies := make(map[string]struct{})

	re, err := regexp.Compile("keyspace=(.*),scope=(.*?),")
	if err != nil {
		return nil, nil, err
	}

	for _, query := range jmxPatterns {
		results, err := client.QueryMBeanAttributes(query)
		if err != nil {
			if jmxErr, ok := gojmx.IsJMXError(err); ok {
				log.Debug("Error querying %s: %v", query, jmxErr)
				continue
			}
			return nil, nil, fmt.Errorf("fatal jmx error while querying: %q: %w", query, err)
		}

		for _, jmxAttr := range results {
			if jmxAttr.ResponseType == gojmx.ResponseTypeErr {
				log.Debug("Failed to process attribute for query: %s status: %s", jmxAttr.Name, jmxAttr.StatusMsg)
				continue
			}
			matches := re.FindStringSubmatch(jmxAttr.Name)
			key := re.ReplaceAllString(jmxAttr.Name, "")

			if len(matches) != 3 {
				metrics[key] = jmxAttr.GetValue()
			} else {
				columnfamily := matches[2]
				keyspace := matches[1]
				eventkey := keyspace + "." + columnfamily

				_, found := internalKeyspaces[keyspace]
				if !found {
					_, found := visitedColumnFamilies[eventkey]
					if !found {
						if len(visitedColumnFamilies) < args.ColumnFamiliesLimit {
							visitedColumnFamilies[eventkey] = struct{}{}
						} else {
							log.Warn("Skipping column family %s due to limit reached. Current limit set to %d",
								columnfamily, args.ColumnFamiliesLimit)
							continue
						}
					}

					_, ok := columnFamilyMetrics[eventkey]
					if !ok {
						columnFamilyMetrics[eventkey] = make(map[string]interface{})
						columnFamilyMetrics[eventkey]["keyspace"] = keyspace
						columnFamilyMetrics[eventkey]["columnFamily"] = columnfamily
						columnFamilyMetrics[eventkey]["keyspaceAndColumnFamily"] = eventkey
					}
					columnFamilyMetrics[eventkey][key] = jmxAttr.GetValue()
				}

			}
		}
	}

	return metrics, columnFamilyMetrics, nil
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
