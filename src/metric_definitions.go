/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"reflect"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
)

// Definitions struct will contain the metrics that have to be collected.
type Definitions struct {
	Common              []Query `yaml:"common"`
	Metrics             []Query `yaml:"metrics"`
	ColumnFamilyMetrics []Query `yaml:"column_family_metrics"`
}

// NewDefinitions returns the definitions of the metrics that have to be collected.
// If extra filtering configuration is provided by the agent, that will be applied to filter the result.
func NewDefinitions() Definitions {
	return Definitions{
		Common:              commonDefinitions,
		Metrics:             metricDefinitions,
		ColumnFamilyMetrics: columnFamilyDefinitions,
	}
}

// MetricNameList filters the definitions of the metrics that have to be collected based on received config.
func (d *Definitions) Filter(config FilteringConfig) {
	// Empty config, nothing to filter.
	if reflect.DeepEqual(config, FilteringConfig{}) {
		return
	}

	d.Common = filterQueries(d.Common, config)
	d.Metrics = filterQueries(d.Metrics, config)
	d.ColumnFamilyMetrics = filterQueries(d.ColumnFamilyMetrics, config)
}

func filterQueries(queries []Query, config FilteringConfig) []Query {
	var result []Query
	// MetricNameList Metric Definitions specified in config.
	for _, query := range queries {
		attributes := query.Attributes
		query.Attributes = []Attribute{}

		for _, attribute := range attributes {
			if config.IsFiltered(attribute) {
				continue
			}
			query.Attributes = append(query.Attributes, attribute)
		}

		if len(query.Attributes) > 0 {
			result = append(result, query)
		}
	}
	return result
}

// commonDefinitions are metric definitions that are common for both CassandraColumnFamilySample and CassandraSample.
var commonDefinitions = []Query{
	{
		MBean: "org.apache.cassandra.db:type=StorageService",
		Attributes: []Attribute{
			{MBeanAttribute: "ReleaseVersion", Alias: "software.version", MetricType: metric.ATTRIBUTE},
			{MBeanAttribute: "ClusterName", Alias: "cluster.name", MetricType: metric.ATTRIBUTE},
		},
	},
	{
		MBean: "org.apache.cassandra.db:type=EndpointSnitchInfo",
		Attributes: []Attribute{
			{MBeanAttribute: "Datacenter", Alias: "cluster.datacenter", MetricType: metric.ATTRIBUTE},
			{MBeanAttribute: "Rack", Alias: "cluster.rack", MetricType: metric.ATTRIBUTE},
		},
	},
}

// columnFamilyDefinitions are the CassandraColumnFamilySample metrics definition.
var columnFamilyDefinitions = []Query{
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=LiveSSTableCount",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.liveSSTableCount", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=SSTablesPerReadHistogram",
		Attributes: []Attribute{
			{MBeanAttribute: "75thPercentile", Alias: "db.SSTablesPerRead75thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "999thPercentile", Alias: "db.SSTablesPerRead999thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "50thPercentile", Alias: "db.SSTablesPerRead50thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "95thPercentile", Alias: "db.SSTablesPerRead95thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "98thPercentile", Alias: "db.SSTablesPerRead98thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "99thPercentile", Alias: "db.SSTablesPerRead99thPercentileMilliseconds", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=LiveDiskSpaceUsed",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.liveDiskSpaceUsedBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=ReadLatency",
		Attributes: []Attribute{
			{MBeanAttribute: "999thPercentile", Alias: "query.readLatency999thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "50thPercentile", Alias: "query.readLatency50thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "95thPercentile", Alias: "query.readLatency95thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "99thPercentile", Alias: "query.readLatency99thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "OneMinuteRate", Alias: "query.readRequestsPerSecond", MetricType: metric.GAUGE},
			{MBeanAttribute: "75thPercentile", Alias: "query.readLatency75thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "98thPercentile", Alias: "query.readLatency98thPercentileMilliseconds", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=WriteLatency",
		Attributes: []Attribute{
			{MBeanAttribute: "999thPercentile", Alias: "query.writeLatency999thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "98thPercentile", Alias: "query.writeLatency98thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "95thPercentile", Alias: "query.writeLatency95thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "OneMinuteRate", Alias: "query.writeRequestsPerSecond", MetricType: metric.GAUGE},
			{MBeanAttribute: "75thPercentile", Alias: "query.writeLatency75thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "99thPercentile", Alias: "query.writeLatency99thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "50thPercentile", Alias: "query.writeLatency50thPercentileMilliseconds", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=PendingCompactions",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.pendingCompactions", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=AllMemtablesHeapSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.allMemtablesOnHeapSizeBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=AllMemtablesOffHeapSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.allMemtablesOffHeapSizeBytes", MetricType: metric.GAUGE},
		},
	},

	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=TombstoneScannedHistogram",
		Attributes: []Attribute{
			{MBeanAttribute: "75thPercentile", Alias: "db.tombstoneScannedHistogram75thPercentile", MetricType: metric.GAUGE},
			{MBeanAttribute: "Count", Alias: "db.tombstoneScannedHistogramCount", MetricType: metric.GAUGE},
			{MBeanAttribute: "95thPercentile", Alias: "db.tombstoneScannedHistogram95thPercentile", MetricType: metric.GAUGE},
			{MBeanAttribute: "999thPercentile", Alias: "db.tombstoneScannedHistogram999thPercentile", MetricType: metric.GAUGE},
			{MBeanAttribute: "99thPercentile", Alias: "db.tombstoneScannedHistogram99thPercentile", MetricType: metric.GAUGE},
			{MBeanAttribute: "50thPercentile", Alias: "db.tombstoneScannedHistogram50thPercentile", MetricType: metric.GAUGE},
			{MBeanAttribute: "98thPercentile", Alias: "db.tombstoneScannedHistogram98thPercentile", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=SpeculativeRetries",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.speculativeRetries", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=BloomFilterFalseRatio",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.bloomFilterFalseRatio", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=MemtableLiveDataSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.memtableLiveDataSize", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ColumnFamily,keyspace=*,scope=*,name=MeanRowSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.meanRowSize", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ColumnFamily,keyspace=*,scope=*,name=MaxRowSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.maxRowSize", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ColumnFamily,keyspace=*,scope=*,name=MinRowSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.minRowSize", MetricType: metric.GAUGE},
		},
	},
}

// SampleAttribute is an attributes that make a NR metric-set unique.
type SampleAttribute struct {
	Key        string
	Alias      string
	MetricType metric.SourceType
}

// columnFamiliesSampleAttributes are NR extra attributes to make CassandraColumnFamilySample unique.
var columnFamiliesSampleAttributes = []SampleAttribute{
	{
		Key:        "keyspace",
		Alias:      "db.keyspace",
		MetricType: metric.ATTRIBUTE,
	},
	{
		Key:        "columnFamily",
		Alias:      "db.columnFamily",
		MetricType: metric.ATTRIBUTE,
	},
	{
		Key:        "keyspaceAndColumnFamily",
		Alias:      "db.keyspaceAndColumnFamily",
		MetricType: metric.ATTRIBUTE,
	},
}

// metricDefinitions are the metric definitions for the CassandraSample.
var metricDefinitions = []Query{
	{
		MBean: "org.apache.cassandra.metrics:type=Table,name=LiveSSTableCount",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.liveSSTableCount", MetricType: metric.GAUGE},
		},
	},
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
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtableReclaimMemory,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalMemtableReclaimMemoryCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=InternalResponseStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalInternalResponseStagePCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=Sampler,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalSamplerActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ReadRepairStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestReadRepairStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=KeyCache,name=Requests",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "db.keyCacheRequestsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=READ_REPAIR,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedReadRepairMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtableReclaimMemory,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtableReclaimMemoryActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=ValidationExecutor,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalValidationExecutorCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=HintsService,name=HintsFailed",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "db.hintsFailedPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Storage,name=TotalHintsInProgress",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.totalHintsInProgress", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=CompactionExecutor,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalCompactionExecutorCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=CacheCleanupExecutor,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalCacheCleanupExecutorCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=AntiEntropyStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalAntiEntropyStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=HintsDispatcher,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalHintsDispatcherPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=CacheCleanupExecutor,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalCacheCleanupExecutorCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=CounterMutationStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestCounterMutationStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=RangeSlice,name=Latency",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.rangeSliceRequestsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ReadStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestReadStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=CommitLog,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.commitLogCompletedTasksPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=ViewWrite,name=Latency",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.viewWriteRequestsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=CacheCleanupExecutor,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalCacheCleanupExecutorPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtablePostFlush,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtablePostFlushPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ViewMutationStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestViewMutationStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=InternalResponseStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalInternalResponseStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=CounterMutationStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestCounterMutationStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=AntiEntropyStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalAntiEntropyStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=RowCache,name=Requests",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "db.rowCacheRequestsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtableFlushWriter,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtableFlushWriterPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=GossipStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalGossipStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ReadStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.requestReadStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=RequestResponseStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.requestRequestResponseStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtableFlushWriter,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtableFlushWriterCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=Sampler,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalSamplerCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=CounterMutationStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.requestCounterMutationStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=KeyCache,name=Size",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.keyCacheSizeBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=CompactionExecutor,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalCompactionExecutorCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=GossipStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalGossipStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=SecondaryIndexManagement,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalSecondaryIndexManagementActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=RowCache,name=OneMinuteHitRate",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.rowCacheHitRate", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=KeyCache,name=Hits",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "db.keyCacheHitsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=HintsService,name=HintsSucceeded",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "db.hintsSucceededPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,name=AllMemtablesOffHeapSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.allMemtablesOffHeapSizeBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=HINT,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedHintMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=COUNTER_MUTATION,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedCounterMutationMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=HintsDispatcher,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalHintsDispatcherCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=Read,name=Latency",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.readRequestsPerSecond", MetricType: metric.GAUGE},
			{MBeanAttribute: "98thPercentile", Alias: "query.readLatency98thPercentileMilliseconds", MetricType: metric.GAUGE},

			{MBeanAttribute: "50thPercentile", Alias: "query.readLatency50thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "999thPercentile", Alias: "query.readLatency999thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "99thPercentile", Alias: "query.readLatency99thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "75thPercentile", Alias: "query.readLatency75thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "95thPercentile", Alias: "query.readLatency95thPercentileMilliseconds", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Table,name=AllMemtablesHeapSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.allMemtablesOnHeapSizeBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MiscStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMiscStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=ValidationExecutor,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalValidationExecutorCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=READ,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedReadMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=RowCache,name=Capacity",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.rowCacheCapacityBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=Write,name=Unavailables",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.writeUnavailablesPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=SecondaryIndexManagement,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalSecondaryIndexManagementPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtableFlushWriter,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalMemtableFlushWriterCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=AntiEntropyStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalAntiEntropyStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=ValidationExecutor,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalValidationExecutorPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=HintsDispatcher,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalHintsDispatcherActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=CacheCleanupExecutor,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalCacheCleanupExecutorActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=Read,name=Timeouts",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.readTimeoutsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ReadStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestReadStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=REQUEST_RESPONSE,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedRequestResponseMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtableReclaimMemory,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtableReclaimMemoryPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=Write,name=Timeouts",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.writeTimeoutsPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=CompactionExecutor,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalCompactionExecutorActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=MutationStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestMutationStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=RangeSlice,name=Unavailables",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.rangeSliceUnavailablesPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=MutationStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestMutationStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=Read,name=Unavailables",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.readUnavailablesPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ViewMutationStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestViewMutationStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=SecondaryIndexManagement,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalSecondaryIndexManagementCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=PendingRangeCalculator,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalPendingRangeCalculatorPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MiscStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMiscStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MiscStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMiscStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtablePostFlush,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtablePostFlushCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtableReclaimMemory,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtableReclaimMemoryCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=BATCH_REMOVE,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedBatchRemoveMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=CASRead,name=Latency",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.CASReadRequestsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=InternalResponseStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalInternalResponseStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=RowCache,name=Size",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.rowCacheSizeBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ViewMutationStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.requestViewMutationStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=SecondaryIndexManagement,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalSecondaryIndexManagementCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=Sampler,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalSamplerCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MigrationStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMigrationStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=PendingRangeCalculator,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalPendingRangeCalculatorCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=GossipStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalGossipStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=ValidationExecutor,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalValidationExecutorActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ReadRepairStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.requestReadRepairStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=AntiEntropyStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalAntiEntropyStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=KeyCache,name=OneMinuteHitRate",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.keyCacheHitRate", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ReadRepairStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestReadRepairStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=CASWrite,name=Latency",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.CASWriteRequestsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=RequestResponseStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestRequestResponseStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MigrationStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMigrationStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=RequestResponseStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestRequestResponseStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtablePostFlush,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalMemtablePostFlushCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=RowCache,name=Hits",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "db.rowCacheHitsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=PAGED_RANGE,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedPagedRangeMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=HintsService,name=HintsTimedOut",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "db.hintsTimedOutPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=CommitLog,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.commitLogPendindTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=CompactionExecutor,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalCompactionExecutorPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ReadRepairStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestReadRepairStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=MUTATION,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedMutationMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=CounterMutationStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestCounterMutationStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ReadStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestReadStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=ViewMutationStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestViewMutationStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Storage,name=Exceptions",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "storage.exceptionCount", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=PendingRangeCalculator,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalPendingRangeCalculatorCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=HintedHandOffManager,name=*",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.hintedHandoffManager", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtableFlushWriter,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtableFlushWriterActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=RequestResponseStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestRequestResponseStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MigrationStage,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMigrationStagePendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=InternalResponseStage,name=CompletedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalInternalResponseStageCompletedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=RangeSlice,name=Timeouts",
		Attributes: []Attribute{
			{MBeanAttribute: "OneMinuteRate", Alias: "query.rangeSliceTimeoutsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=MutationStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.requestMutationStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=BATCH_STORE,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedBatchStoreMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=HintsDispatcher,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalHintsDispatcherCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=GossipStage,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalGossipStageActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Storage,name=TotalHints",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.totalHintsPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Storage,name=Load",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.loadBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ClientRequest,scope=Write,name=Latency",
		Attributes: []Attribute{
			{MBeanAttribute: "999thPercentile", Alias: "query.writeLatency999thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "75thPercentile", Alias: "query.writeLatency75thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "95thPercentile", Alias: "query.writeLatency95thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "98thPercentile", Alias: "query.writeLatency98thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "50thPercentile", Alias: "query.writeLatency50thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "99thPercentile", Alias: "query.writeLatency99thPercentileMilliseconds", MetricType: metric.GAUGE},
			{MBeanAttribute: "OneMinuteRate", Alias: "query.writeRequestsPerSecond", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=Sampler,name=PendingTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalSamplerPendingTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=PendingRangeCalculator,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalPendingRangeCalculatorActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=CommitLog,name=TotalCommitLogSize",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.commitLogTotalSizeBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=request,scope=MutationStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.requestMutationStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=DroppedMessage,scope=_TRACE,name=Dropped",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.droppedTraceMessagesPerSecond", MetricType: metric.RATE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MemtablePostFlush,name=ActiveTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.threadpool.internalMemtablePostFlushActiveTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MigrationStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalMigrationStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=Cache,scope=KeyCache,name=Capacity",
		Attributes: []Attribute{
			{MBeanAttribute: "Value", Alias: "db.keyCacheCapacityBytes", MetricType: metric.GAUGE},
		},
	},
	{
		MBean: "org.apache.cassandra.metrics:type=ThreadPools,path=internal,scope=MiscStage,name=CurrentlyBlockedTasks",
		Attributes: []Attribute{
			{MBeanAttribute: "Count", Alias: "db.threadpool.internalMiscStageCurrentlyBlockedTasks", MetricType: metric.GAUGE},
		},
	},
}
