package main

import (
	"github.com/newrelic/infra-integrations-sdk/data/attribute"
	"os"
	"strconv"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/jmx"
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
}

const (
	integrationName    = "com.newrelic.cassandra"
	integrationVersion = "2.4.2"

	entityRemoteType = "node"
)

var (
	args argumentList
)

func main() {
	i, err := createIntegration()
	fatalIfErr(err)

	e, err := entity(i)
	fatalIfErr(err)

	var opts []jmx.Option
	if args.Verbose {
		opts = append(opts, jmx.WithVerbose())
	}

	fatalIfErr(jmx.Open(args.Hostname, strconv.Itoa(args.Port), args.Username, args.Password, opts...))
	defer jmx.Close()

	if args.HasMetrics() {
		rawMetrics, allColumnFamilies, err := getMetrics()
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
		populateInventory(e.Inventory, rawInventory)
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
