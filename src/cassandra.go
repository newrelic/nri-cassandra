//go:generate goversioninfo
package main

import (
	"context"
	"fmt"
	"github.com/newrelic/nrjmx/gojmx"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

type argumentList struct {
	sdk_args.DefaultArgumentList

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
	EnableInternalStats bool   `default:"false" help:"Print nrjmx internal query stats"`
	LongRunning         bool   `default:"false" help:"Specify if this is a long running integration"`
	HeartbeatInterval   int    `default:"5" help:"Interval in seconds vor submitting the heartbeat while in long running mode"`
	Interval            int    `default:"30" help:"Interval in seconds for collecting data while in long running mode"`
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

	e, err := entity(i)
	fatalIfErr(err)

	if args.HasMetrics() {
		jmxClient, err := openJMXConnection()
		fatalIfErr(err)

		defer func() {
			if err := jmxClient.Close(); err != nil {
				log.Error(
					"Failed to close JMX connection: %s", err)
			}
		}()

		config, err := LoadConfig()
		if err != nil {
			// Not a fatal error.
			log.Error(
				"Failed to load configuration: %v", err)
		}

		definitions := GetDefinitions(config)

		err = runMetricCollection(i, e, jmxClient, definitions)
		fatalIfErr(err)
	}

	if args.HasInventory() {
		rawInventory, err := getInventory()
		fatalIfErr(err)
		err = populateInventory(e.Inventory, rawInventory)
		fatalIfErr(err)
	}

	fatalIfErr(i.Publish())
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
		if _, ok := gojmx.IsJMXClientError(err); ok || !args.LongRunning {
			return nil, fmt.Errorf("failed to open JMX connection, error: %w, Config: (%s)",
				err,
				formattedConfig,
			)
		}

		// We can recover in long-running mode.
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

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func logInternalStats(jmxClient *gojmx.Client) {
	internalStats, err := jmxClient.GetInternalStats()
	if err != nil {
		log.Error("Failed to collect nrjmx internal stats")
		return
	}

	calls := 0
	totalObjects := 0
	totalTimeMs := 0.0
	totalAttrs := 0
	successful := 0

	for _, stat := range internalStats {
		if stat.Successful {
			successful++
		}
		calls++
		totalObjects += int(stat.ResponseCount)
		totalTimeMs += stat.Milliseconds
		totalAttrs += len(stat.Attrs)
		log.Debug(fmt.Sprintf("%v", stat))
	}
	log.Debug(fmt.Sprintf("totalMs: '%.3f', totalObjects: %d, totalAttr: %d, totalCalls: %d, totalSuccessful: %d", totalTimeMs, totalObjects, totalAttrs, calls, successful))
}

func runMetricCollection(i *integration.Integration, e *integration.Entity, jmxClient *gojmx.Client, definitions Definitions) error {
	if !args.LongRunning {
		return collectMetrics(e, jmxClient, definitions)
	}

	heartBeat := time.NewTicker(time.Duration(args.HeartbeatInterval) * time.Second)
	metricInterval := time.NewTicker(time.Millisecond)

	first := true
	for {
		select {
		case <-heartBeat.C:
			log.Debug("Sending heartBeat")
			// heartbeat signal for long-running integrations
			// https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/host-integrations-newer-configuration-format#timeout
			fmt.Println("{}")
		case <-metricInterval.C:
			if !jmxClient.IsRunning() {
				log.Fatal(fmt.Errorf("jmx client not running"))
			}

			if first {
				metricInterval.Reset(time.Duration(args.Interval) * time.Second)
				first = false
			}

			e2, err := entity(i)
			if err != nil {
				log.Error("Failed to create entity: %v", err)
			}

			if err := collectMetrics(e2, jmxClient, definitions); err != nil {
				log.Error("Failed to collect metrics, error: %v", err)
			}

			if err := i.Publish(); err != nil {
				log.Error("Failed to publish metrics, error: %v", err)
			}
		}
	}
}

func collectMetrics(entity *integration.Entity, jmxClient *gojmx.Client, definitions Definitions) error {
	defer func() {
		if !args.EnableInternalStats {
			return
		}
		logInternalStats(jmxClient)
	}()

	rawMetrics, err := getMetrics(jmxClient, definitions.Metrics)
	if err != nil {
		return err
	}

	commonMetrics, err := getMetrics(jmxClient, definitions.Common)
	if err != nil {
		return err
	}

	ms := metricSet(entity, "CassandraSample", args.Hostname, args.Port, args.RemoteMonitoring)
	populateMetrics(ms, rawMetrics, definitions.Metrics)
	populateMetrics(ms, commonMetrics, definitions.Common)

	if args.ColumnFamiliesLimit > 0 {
		allColumnFamilies, err := getColumnFamilyMetrics(jmxClient, definitions.ColumnFamilyMetrics)
		if err != nil {
			return err
		}

		for _, columnFamilyMetrics := range allColumnFamilies {
			s := metricSet(entity, "CassandraColumnFamilySample", args.Hostname, args.Port, args.RemoteMonitoring)
			populateMetrics(s, commonMetrics, definitions.Common)
			populateMetrics(s, columnFamilyMetrics, definitions.ColumnFamilyMetrics)
			populateAttributes(s, columnFamilyMetrics, columnFamiliesSampleAttributes)
		}
	}
	return nil
}
