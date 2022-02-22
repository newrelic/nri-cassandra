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
	HideSecrets         bool   `default:"true" help:"Set this to false if you want to see the secrets in the verbose logs."`
	KeyStore            string `default:"" help:"The location for the keystore containing JMX Client's SSL certificate"`
	KeyStorePassword    string `default:"" help:"Password for the SSL Key Store"`
	TrustStore          string `default:"" help:"The location for the keystore containing JMX Server's SSL certificate"`
	TrustStorePassword  string `default:"" help:"Password for the SSL Trust Store"`
	ShowVersion         bool   `default:"false" help:"Print build information and exit"`
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

	jmxConfig := &gojmx.JMXConfig{
		KeyStore:           args.KeyStore,
		KeyStorePassword:   args.KeyStorePassword,
		TrustStore:         args.TrustStore,
		TrustStorePassword: args.TrustStorePassword,
		Hostname:           args.Hostname,
		Port:               int32(args.Port),
		Username:           args.Username,
		Password:           args.Password,
		RequestTimeoutMs:   int64(args.Timeout),
		Verbose:            args.Verbose,
	}

	jmxClient := gojmx.NewClient(context.Background())
	_, err = jmxClient.Open(jmxConfig)
	log.Debug("nrjmx version: %s", jmxClient.GetClientVersion())

	if err != nil {
		log.Error("Failed to open JMX connection, error: %v, Config: (%s)",
			err,
			gojmx.FormatConfig(jmxConfig, args.HideSecrets),
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
		rawMetrics, allColumnFamilies, err := getMetrics(jmxClient)
		fatalIfErr(err)
		ms := metricSet(e, "CassandraSample", args.Hostname, args.Port, args.RemoteMonitoring)
		populateMetrics(ms, rawMetrics, metricsDefinition)
		populateMetrics(ms, rawMetrics, commonDefinition)

		for _, columnFamilyMetrics := range allColumnFamilies {
			s := metricSet(e, "CassandraColumnFamilySample", args.Hostname, args.Port, args.RemoteMonitoring)
			populateMetrics(s, columnFamilyMetrics, columnFamilyDefinition)
			populateMetrics(s, rawMetrics, commonDefinition)
		}
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
