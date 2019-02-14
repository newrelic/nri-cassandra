package main

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"strconv"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/jmx"
	"github.com/newrelic/infra-integrations-sdk/log"
	"os"
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
	integrationVersion = "2.1.0"

	entityRemoteType = "cassandra"
)

var (
	args argumentList
)

func main() {
	var i *integration.Integration
	var err error
	cachePath := os.Getenv("NRIA_CACHE_PATH")

	if cachePath == "" {
		i, err = integration.New(integrationName, integrationVersion, integration.Args(&args))
	} else {
		var storer persist.Storer

		logger := log.NewStdErr(args.Verbose)
		storer, err = persist.NewFileStore(cachePath, logger, persist.DefaultTTL)
		fatalIfErr(err)

		i, err = integration.New(integrationName, integrationVersion, integration.Args(&args),
			integration.Storer(storer), integration.Logger(logger))
	}

	fatalIfErr(err)
	log.SetupLogging(args.Verbose)

	e, err := entity(i)
	fatalIfErr(err)

	fatalIfErr(jmx.Open(args.Hostname, strconv.Itoa(args.Port), args.Username, args.Password))
	defer jmx.Close()

	if args.HasMetrics() {
		rawMetrics, allColumnFamilies, err := getMetrics()
		fatalIfErr(err)

		hostnameAttr := metric.Attr("hostname", args.Hostname)
		portAttr := metric.Attr("port", strconv.Itoa(args.Port))
		ms := e.NewMetricSet("CassandraSample", hostnameAttr, portAttr)
		populateMetrics(ms, rawMetrics, metricsDefinition)
		populateMetrics(ms, rawMetrics, commonDefinition)

		for _, columnFamilyMetrics := range allColumnFamilies {
			s := e.NewMetricSet("CassandraColumnFamilySample")
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

func entity(i *integration.Integration) (*integration.Entity, error) {
	if args.RemoteMonitoring {
		n := fmt.Sprintf("%s:%d", args.Hostname, args.Port)
		return i.Entity(n, entityRemoteType)
	}

	return i.LocalEntity(), nil
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
