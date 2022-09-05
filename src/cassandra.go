//go:generate goversioninfo
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/newrelic/nrjmx/gojmx"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList

	Hostname            string `default:"localhost" help:"Hostname or IP where Cassandra is running."`
	Port                int    `default:"7199" help:"Port on which JMX server is listening."`
	Username            string `default:"" help:"Username for accessing JMX."`
	Password            string `default:"" help:"Password for the given user."`
	ConfigPath          string `default:"/etc/cassandra/cassandra.yaml" help:"Cassandra configuration file."`
	Timeout             int    `default:"2000" help:"Timeout in milliseconds per single JMX query."`
	ColumnFamiliesLimit int    `default:"20" help:"Limit on number of Cassandra Column Families."`
	RemoteMonitoring    bool   `default:"false" help:"Identifies the monitored entity as 'remote'. In doubt: set to true."`
	KeyStore            string `default:"" help:"The location for the keystore containing JMX Client's SSL certificate"`
	KeyStorePassword    string `default:"" help:"Password for the SSL Key Store"`
	TrustStore          string `default:"" help:"The location for the keystore containing JMX Server's SSL certificate"`
	TrustStorePassword  string `default:"" help:"Password for the SSL Trust Store"`
	ShowVersion         bool   `default:"false" help:"Print build information and exit"`
	LongRunning         bool   `default:"false" help:"BETA: In long-running mode integration process will be kept alive"`
	HeartbeatInterval   int    `default:"5" help:"BETA: Interval in seconds for submitting the heartbeat while in long-running mode"`
	Interval            int    `default:"30" help:"BETA: Interval in seconds for collecting data while while in long-running mode"`
	MetricsFilter       string `default:"" help:"BETA: Filtering rules for metrics collection"`
	EnableInternalStats bool   `default:"false" help:"Print nrjmx internal query stats for troubleshooting"`
}

const (
	integrationName  = "com.newrelic.cassandra"
	entityRemoteType = "node"
)

var (
	args               argumentList
	integrationVersion = "0.0.0"
	gitCommit          = ""
	buildDate          = ""

	errNRJMXNotRunning = errors.New("nrjmx client sub-process not running")
)

func main() {
	i, err := createIntegration()
	fatalIfErr(err)

	if args.ShowVersion {
		fmt.Printf(
			"New Relic %s integration Version: %s, Platform: %s, GoVersion: %s, GitCommit: %s, BuildDate: %s\n",
			strings.Title(strings.Replace(integrationName, "com.newrelic.", "", 1)),
			integrationVersion,
			fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			runtime.Version(),
			gitCommit,
			buildDate)
		os.Exit(0)
	}

	if args.HasMetrics() {
		jmxClient, conErr := openJMXConnection()
		fatalIfErr(conErr)

		defer func() {
			if err := jmxClient.Close(); err != nil {
				log.Error(
					"Failed to close JMX connection: %s", err)
			}
		}()

		err := runMetricCollection(i, jmxClient)
		fatalIfErr(err)
	}

	if args.HasInventory() {
		e, err := entity(i)
		fatalIfErr(err)

		rawInventory, err := getInventory()
		fatalIfErr(err)
		err = populateInventory(e.Inventory, rawInventory)
		fatalIfErr(err)
	}

	fatalIfErr(i.Publish())
}

// runMetricCollection will perform the metrics collection.
func runMetricCollection(i *integration.Integration, jmxClient *gojmx.Client) error {
	definitions := NewDefinitions()

	config, err := LoadFilteringConfig(args.MetricsFilter)
	if err != nil {
		return fmt.Errorf("failed to load metrics filtering configuration, error: %w", err)
	}
	definitions.Filter(config)

	if args.LongRunning {
		return collectMetricsEachInterval(i, jmxClient, definitions)
	}
	return collectMetrics(i, jmxClient, definitions)
}

// collectMetricsEachInterval will collect the metrics periodically when configured in long-running mode.
func collectMetricsEachInterval(i *integration.Integration, jmxClient *gojmx.Client, definitions Definitions) error {
	metricInterval := time.NewTicker(time.Duration(args.Interval) * time.Second)

	runHeartBeat()

	// do ... while.
	for ; true; <-metricInterval.C {
		// Check if the nrjmx java sub-process is still alive.
		if !jmxClient.IsRunning() {
			return errNRJMXNotRunning
		}

		if err := collectMetrics(i, jmxClient, definitions); err != nil {
			log.Error("Failed to collect metrics, error: %v", err)
			continue
		}

		if err := i.Publish(); err != nil {
			log.Error("Failed to publish metrics, error: %v", err)
			continue
		}
	}

	return nil
}

// collectMetrics will gather all the required metrics from the JMX endpoint and attach them the the sdk integration.
func collectMetrics(i *integration.Integration, jmxClient *gojmx.Client, definitions Definitions) error {
	// For troubleshooting purpose, if enabled, integration will log internal query stats.
	if args.EnableInternalStats {
		defer func() {
			logInternalStats(jmxClient)
		}()
	}

	e, err := entity(i)
	if err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	rawMetrics, err := getMetrics(jmxClient, definitions.Metrics)
	if err != nil {
		return err
	}

	commonMetrics, err := getMetrics(jmxClient, definitions.Common)
	if err != nil {
		return err
	}

	ms := metricSet(e, "CassandraSample", args.Hostname, args.Port, args.RemoteMonitoring)
	populateMetrics(ms, rawMetrics, definitions.Metrics)
	populateMetrics(ms, commonMetrics, definitions.Common)

	if args.ColumnFamiliesLimit > 0 {
		allColumnFamilies, err := getColumnFamilyMetrics(jmxClient, definitions.ColumnFamilyMetrics)
		if err != nil {
			return err
		}

		for _, columnFamilyMetrics := range allColumnFamilies {
			s := metricSet(e, "CassandraColumnFamilySample", args.Hostname, args.Port, args.RemoteMonitoring)
			populateMetrics(s, commonMetrics, definitions.Common)
			populateMetrics(s, columnFamilyMetrics, definitions.ColumnFamilyMetrics)
			populateAttributes(s, columnFamilyMetrics, columnFamiliesSampleAttributes)
		}
	}
	return nil
}

func metricSet(e *integration.Entity, eventType, hostname string, port int, remoteMonitoring bool) *metric.Set {
	if remoteMonitoring {
		return e.NewMetricSet(
			eventType,
			attribute.Attr("hostname", hostname),
			attribute.Attr("port", strconv.Itoa(port)),
		)
	}

	return e.NewMetricSet(
		eventType,
		attribute.Attr("port", strconv.Itoa(port)),
	)
}

func createIntegration() (*integration.Integration, error) {
	cachePath := os.Getenv("NRIA_CACHE_PATH")
	if cachePath == "" {
		return integration.New(integrationName, integrationVersion, integration.Args(&args))
	}

	l := log.NewStdErr(args.Verbose)
	s, err := persist.NewFileStore(cachePath, l, persist.DefaultTTL)
	if err != nil {
		return nil, err
	}

	return integration.New(integrationName, integrationVersion, integration.Args(&args), integration.Storer(s), integration.Logger(l))
}

// getJMXConfig will use the integration args to prepare the JMXConfig for the JMXClient.
func getJMXConfig() *gojmx.JMXConfig {
	jmxConfig := &gojmx.JMXConfig{
		Hostname:            args.Hostname,
		Port:                int32(args.Port),
		Username:            args.Username,
		Password:            args.Password,
		RequestTimeoutMs:    int64(args.Timeout),
		Verbose:             args.Verbose,
		EnableInternalStats: args.EnableInternalStats,
	}

	if args.KeyStore != "" && args.KeyStorePassword != "" && args.TrustStore != "" && args.TrustStorePassword != "" {
		jmxConfig.KeyStore = args.KeyStore
		jmxConfig.KeyStorePassword = args.KeyStorePassword
		jmxConfig.TrustStore = args.TrustStore
		jmxConfig.TrustStorePassword = args.TrustStorePassword
	}

	return jmxConfig
}

// openJMXConnection configures the JMX client and attempts to connect to the endpoint.
func openJMXConnection() (*gojmx.Client, error) {
	jmxConfig := getJMXConfig()

	hideSecrets := true
	formattedConfig := gojmx.FormatConfig(jmxConfig, hideSecrets)

	jmxClient := gojmx.NewClient(context.Background())
	_, err := jmxClient.Open(jmxConfig)

	log.Debug("nrjmx version: %s, config: %s", jmxClient.GetClientVersion(), formattedConfig)

	if err != nil {
		// When not in long-running mode, we cannot recover from any type of connection error.
		// However, in long-running mode, we can recover later from errors related with connection, except JMXClient error
		// which means that the nrjmx java sub-process was closed.
		if _, ok := gojmx.IsJMXClientError(err); ok || !args.LongRunning {
			return nil, fmt.Errorf("failed to open JMX connection, error: %w, Config: (%s)",
				err,
				formattedConfig,
			)
		}

		// In long-running mode just log the error.
		log.Error("Error while connecting to jmx connection, err: %v", err)
	}

	return jmxClient, nil
}

func entity(i *integration.Integration) (*integration.Entity, error) {
	if args.RemoteMonitoring {
		return i.Entity(args.Hostname, entityRemoteType)
	}

	return i.LocalEntity(), nil
}

// runHeartBeat is used in long-running mode to signal to the agent that the integration is alive.
func runHeartBeat() {
	heartBeat := time.NewTicker(time.Duration(args.HeartbeatInterval) * time.Second)

	go func() {
		for range heartBeat.C {
			log.Debug("Sending heartBeat")
			// heartbeat signal for long-running integrations
			// https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/host-integrations-newer-configuration-format#timeout
			fmt.Println("{}")
		}
	}()
}

// logInternalStats will print in verbose logs statistics gathered by nrjmx client
// that can be handy when troubleshooting performance issues.
func logInternalStats(jmxClient *gojmx.Client) {
	internalStats, err := jmxClient.GetInternalStats()
	if err != nil {
		log.Error("Failed to collect nrjmx internal stats, %v", err)
		return
	}

	for _, stat := range internalStats {
		log.Debug("%v", stat)
	}

	// Aggregated stats.
	log.Debug("%v", internalStats)
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
