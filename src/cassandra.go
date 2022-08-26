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
	LongRunning         bool   `default:"false" help:"In long-running mode integration process will be kept alive"`
	HeartbeatInterval   int    `default:"5" help:"Interval in seconds for submitting the heartbeat while in long-running mode"`
	Interval            int    `default:"30" help:"Interval in seconds for collecting data while while in long-running mode"`
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

	e, err := entity(i)
	fatalIfErr(err)

	jmxConfig := &gojmx.JMXConfig{
		Hostname:         args.Hostname,
		Port:             int32(args.Port),
		Username:         args.Username,
		Password:         args.Password,
		RequestTimeoutMs: int64(args.Timeout),
		Verbose:          args.Verbose,
	}

	if args.KeyStore != "" && args.KeyStorePassword != "" && args.TrustStore != "" && args.TrustStorePassword != "" {
		jmxConfig.KeyStore = args.KeyStore
		jmxConfig.KeyStorePassword = args.KeyStorePassword
		jmxConfig.TrustStore = args.TrustStore
		jmxConfig.TrustStorePassword = args.TrustStorePassword
	}

	hideSecrets := true
	formattedConfig := gojmx.FormatConfig(jmxConfig, hideSecrets)

	jmxClient := gojmx.NewClient(context.Background())
	_, err = jmxClient.Open(jmxConfig)
	log.Debug("nrjmx version: %s, config: %s", jmxClient.GetClientVersion(), formattedConfig)

	if err != nil {
		log.Error("Failed to open JMX connection, error: %v, Config: (%s)",
			err,
			formattedConfig,
		)
		os.Exit(1)
	}

	defer func() {
		if err := jmxClient.Close(); err != nil {
			log.Error(
				"Failed to close JMX connection: %s", err)
		}
	}()

	if args.HasMetrics() {
		err = runMetricCollection(i, e, jmxClient)
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

func runMetricCollection(i *integration.Integration, e *integration.Entity, jmxClient *gojmx.Client) error {
	if !args.LongRunning {
		collectMetrics(e, jmxClient)
		return nil
	}

	heartBeat := time.NewTicker(time.Duration(args.HeartbeatInterval) * time.Second)
	metricInterval := time.NewTicker(time.Millisecond)

	go func() {
		for range heartBeat.C {
			log.Debug("Sending heartBeat")
			// heartbeat signal for long-running integrations
			// https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/host-integrations-newer-configuration-format#timeout
			fmt.Println("{}")
		}
	}()

	// Force the first collection to happen immediately.
	first := true

	for range metricInterval.C {
		if first {
			metricInterval.Reset(time.Duration(args.Interval) * time.Second)
			first = false
		}

		// Check if the nrjmx java sub-process is still alive.
		if !jmxClient.IsRunning() {
			return errNRJMXNotRunning
		}

		e2, err := entity(i)
		if err != nil {
			log.Error("Failed to create entity: %v", err)
			continue
		}

		collectMetrics(e2, jmxClient)

		if err := i.Publish(); err != nil {
			log.Error("Failed to publish metrics, error: %v", err)
			continue
		}
	}

	return nil
}

func collectMetrics(entity *integration.Entity, jmxClient *gojmx.Client) {
	rawMetrics, allColumnFamilies, err := getMetrics(jmxClient)
	fatalIfErr(err)
	ms := metricSet(entity, "CassandraSample", args.Hostname, args.Port, args.RemoteMonitoring)
	populateMetrics(ms, rawMetrics, metricsDefinition)
	populateMetrics(ms, rawMetrics, commonDefinition)

	for _, columnFamilyMetrics := range allColumnFamilies {
		s := metricSet(entity, "CassandraColumnFamilySample", args.Hostname, args.Port, args.RemoteMonitoring)
		populateMetrics(s, columnFamilyMetrics, columnFamilyDefinition)
		populateMetrics(s, rawMetrics, commonDefinition)
	}
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
