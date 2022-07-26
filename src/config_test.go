package main

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestTest(t *testing.T) {
	metrics := `query.CASWriteRequestsPerSecond
query.CASReadRequestsPerSecond
query.viewWriteRequestsPerSecond
query.rangeSliceRequestsPerSecond
query.readRequestsPerSecond
query.writeRequestsPerSecond      
db.threadpool.requestCounterMutationStagePendingTasks
db.threadpool.requestViewMutationStagePendingTasks
db.threadpool.requestReadRepairStagePendingTasks
db.threadpool.requestReadStagePendingTasks
db.threadpool.requestRequestResponseStagePendingTasks
db.threadpool.requestMutationStagePendingTasks
db.threadpool.requestCounterMutationStageActiveTasks
db.threadpool.requestViewMutationStageActiveTasks
db.threadpool.requestReadRepairStageActiveTasks
db.threadpool.requestReadStageActiveTasks
db.threadpool.requestRequestResponseStageActiveTasks
db.threadpool.requestMutationStageActiveTasks
db.threadpool.requestReadStagePendingTasks 
db.threadpool.requestReadStageActiveTasks  
query.writeLatency50thPercentileMilliseconds
query.writeLatency75thPercentileMilliseconds
query.writeLatency95thPercentileMilliseconds
query.writeLatency98thPercentileMilliseconds
query.writeLatency99thPercentileMilliseconds  
db.threadpool.internalAntiEntropyStageActiveTasks
db.threadpool.internalCacheCleanupExecutorActiveTasks
db.threadpool.internalCompactionExecutorActiveTasks
db.threadpool.internalGossipStageActiveTasks
db.threadpool.internalHintsDispatcherActiveTasks
db.threadpool.internalInternalResponseStageActiveTasks
db.threadpool.internalMemtableFlushWriterActiveTasks
db.threadpool.internalMemtablePostFlushActiveTasks
db.threadpool.internalMemtableReclaimMemoryActiveTasks
db.threadpool.internalMigrationStageActiveTasks
db.threadpool.internalMiscStageActiveTasks
db.threadpool.internalPendingRangeCalculatorActiveTasks
db.threadpool.internalSamplerActiveTasks
db.threadpool.internalSecondaryIndexManagementActiveTasks
db.threadpool.internalValidationExecutorActiveTasks    
query.readLatency50thPercentileMilliseconds
query.readLatency75thPercentileMilliseconds
query.readLatency95thPercentileMilliseconds
query.readLatency98thPercentileMilliseconds
query.readLatency99thPercentileMilliseconds      
db.threadpool.internalAntiEntropyStagePendingTasks
db.threadpool.internalCacheCleanupExecutorActiveTasks
db.threadpool.internalCompactionExecutorActiveTasks
db.threadpool.internalGossipStageActiveTasks
db.threadpool.internalHintsDispatcherActiveTasks
db.threadpool.internalInternalResponseStageActiveTasks
db.threadpool.internalMemtableFlushWriterActiveTasks
db.threadpool.internalMemtablePostFlushActiveTasks
db.threadpool.internalMemtableReclaimMemoryActiveTasks
db.threadpool.internalMigrationStageActiveTasks
db.threadpool.internalMiscStageActiveTasks
db.threadpool.internalPendingRangeCalculatorActiveTasks
db.threadpool.internalSamplerActiveTasks
db.threadpool.internalSecondaryIndexManagementActiveTasks
db.threadpool.internalValidationExecutorActiveTasks
db.liveSSTableCount
db.keyspace
db.SSTablesPerRead50thPercentileMilliseconds
db.SSTablesPerRead75thPercentileMilliseconds
db.SSTablesPerRead95thPercentileMilliseconds
db.SSTablesPerRead98thPercentileMilliseconds
db.SSTablesPerRead99thPercentileMilliseconds
db.droppedBatchRemoveMessagesPerSecond
db.droppedBatchStoreMessagesPerSecond
db.droppedCounterMutationMessagesPerSecond
db.droppedHintMessagesPerSecond
db.droppedMutationMessagesPerSecond
db.droppedPagedRangeMessagesPerSecond
db.droppedRangeSliceMessagesPerSecond
db.droppedReadMessagesPerSecond
db.droppedReadRepairMessagesPerSecond
db.droppedRequestResponseMessagesPerSecond
db.droppedTraceMessagesPerSecond
db.allMemtablesOnHeapSizeBytes
db.allMemtablesOffHeapSizeBytes
db.pendingCompactions
db.columnFamily
db.keyspace
db.totalHintsInProgress`

	split := strings.Split(metrics, "\n")

	dedup := map[string]struct{}{}

	for _, s := range split {
		dedup[strings.TrimSpace(s)] = struct{}{}
	}

	column := []string{}
	metric := []string{}
	common := []string{}

	for k := range dedup {
		found := false
	loop:
		for _, def := range columnFamilyDefinitions {
			for _, attr := range def.Attributes {
				if attr.Alias == k {
					column = append(column, attr.Alias)
					found = true
					break loop
				}
			}
		}

		if found {
			continue
		}
	loop2:
		for _, def := range metricDefinitions {
			for _, attr := range def.Attributes {
				if attr.Alias == k {
					metric = append(metric, attr.Alias)
					found = true
					break loop2
				}
			}
		}

		if found {
			continue
		}

	loop3:
		for _, def := range commonDefinitions {
			for _, attr := range def.Attributes {
				if attr.Alias == k {
					common = append(common, attr.Alias)
					found = true
					break loop3
				}
			}
		}

		if !found {
			//fmt.Println("not found!!!", k)
		}

		for _, m := range metric {
			fmt.Println("-", m)
		}
	}

}

func TestExcludeWildcard(t *testing.T) {
	configYAML := `
exclude:
  metrics:
    - "*"
`
	f, err := ioutil.TempFile("", "cassandra_config.yml")
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()

	f.WriteString(configYAML)
	f.Close()

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
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()

	f.WriteString(configYAML)
	f.Close()

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
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, os.Remove(f.Name()))
	}()

	f.WriteString(configYAML)
	f.Close()

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
