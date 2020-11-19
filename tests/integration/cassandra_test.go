// +build integration

package integration

import (
	"flag"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-cassandra/tests/integration/helpers"
	"github.com/newrelic/nri-cassandra/tests/integration/jsonschema"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

const (
	envCassandraVersion = "3.11.0"
	timeout             = "30000"
)

var (
	iName = "cassandra"

	defaultContainer = "integration_nri-cassandra_1"

	iVersion = "1.1.0"

	defaultBinPath  = "/nri-cassandra"
	defaultHostname = "cassandra"

	schemaFolder = fmt.Sprintf("json-schema-files-%s", envCassandraVersion)

	// cli flags
	container = flag.String("container", defaultContainer, "container where the integration is installed")
	binPath   = flag.String("bin", defaultBinPath, "Integration binary path")

	hostname = flag.String("hostname", defaultHostname, "cassandra hostname")
)

// Returns the standard output, or fails testing if the command returned an error
func runIntegration(t *testing.T, envVars ...string) (string, string, error) {
	t.Helper()

	command := make([]string, 0)
	command = append(command, *binPath)

	var found bool
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, "HOSTNAME") {
			found = true
			break
		}
	}

	if !found && hostname != nil {
		command = append(command, "--hostname", *hostname)
	}

	stdout, stderr, err := helpers.ExecInContainer(*container, command, envVars...)

	if stderr != "" {
		log.Debug("Integration command Standard Error: ", stderr)
	}

	return stdout, stderr, err
}

func TestMain(m *testing.M) {
	fmt.Println("Wait for cassandra to initialize...")
	time.Sleep(30 * time.Second)

	flag.Parse()

	result := m.Run()

	os.Exit(result)
}

func TestCassandraIntegration_ValidArguments(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"CONFIG_PATH=/etc/cassandra/cassandra.yaml",
		fmt.Sprintf("TIMEOUT=%v", timeout),
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.NotNil(t, stderr, "unexpected stderr")

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema.json")

	err = jsonschema.Validate(schemaPath, stdout)

	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_OnlyMetrics(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"METRICS=true",
		fmt.Sprintf("TIMEOUT=%v", timeout),
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.NotNil(t, stderr, "unexpected stderr")

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_OnlyInventory(t *testing.T) {
	testName := helpers.GetTestName(t)
	stdout, stderr, err := runIntegration(t,
		"CONFIG_PATH=/etc/cassandra/cassandra.yaml",
		"INVENTORY=true",
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.NotNil(t, stderr, "unexpected stderr")

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-inventory.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_Error_NoUserCredentials(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	)

	expectedErrorMessage := "Access denied"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestCassandraIntegration_NoError_NoUserCredentials_OnlyInventory(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"INVENTORY=true",
		"CONFIG_PATH=/etc/cassandra/cassandra.yaml",
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.NotNil(t, stderr, "unexpected stderr")

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-inventory.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_Error_NoPassword(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"USERNAME=monitorRole",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	)
	expectedErrorMessage := "Access denied"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestCassandraIntegration_Error_InvalidPassword(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"USERNAME=monitorRole",
		"PASSWORD=invalid_password",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	)

	expectedErrorMessage := "Access denied"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestCassandraIntegration_Error_InvalidUsername(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"USERNAME=invalid_username",
		"PASSWORD=monitorPwd",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	)

	expectedErrorMessage := "Access denied"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestCassandraIntegration_Error_InvalidHostname(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)
	stdout, stderr, err := runIntegration(t,
		"HOSTNAME=nonExistingHost",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	)

	expectedErrorMessage := "no such host"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestCassandraIntegration_Error_InvalidPort(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)
	stdout, stderr, err := runIntegration(t,
		"USERNAME=monitorRole",
		"PASSWORD=monitorPwd",
		"PORT=1",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	)

	expectedErrorMessage := "Failed to connect, invalid port"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestCassandraIntegration_Error_InvalidConfigPath_NonExistingFile(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"CONFIG_PATH=/nonExisting.yaml",
		"INVENTORY=true",
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	expectedErrorMessage := "no such file or directory"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestCassandraIntegration_NoError_InvalidConfigPath_NonExistingFile_OnlyMetrics(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"CONFIG_PATH=/nonExisting.yaml",
		"USERNAME=monitorRole",
		"PASSWORD=monitorPwd",
		"METRICS=true",
		fmt.Sprintf("TIMEOUT=%v", timeout),
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.NotNil(t, stderr, "unexpected stderr")

	// Cassandra 3.7 does not report DataCenter and Racks names through JMX
	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_Error_InvalidConfigPath_ExistingFile(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	tmpfile, err := ioutil.TempFile("", "empty.yaml")
	if err != nil {
		t.Fatalf("Cannot create a new temporary file, got error: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	testName := helpers.GetTestName(t)
	stdout, stderr, err := runIntegration(t,
		fmt.Sprintf("CONFIG_PATH=%s", tmpfile.Name()),
		"INVENTORY=true",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	)

	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	expectedErrorMessage := "Config path not correctly set, cannot fetch inventory data"

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-empty.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)
}

func TestCassandraIntegration_Error_InvalidConfigPath_ExistingDirectory(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		fmt.Sprintf("CONFIG_PATH=%s", "/etc/cassandra/"),
		"INVENTORY=true",
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	expectedErrorMessage := "is a directory"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestCassandraIntegration_NoError_IncompleteSSLConfig(t *testing.T) {
	// Enabling SSL requires setting all 4 keystore/truststore options.
	//
	// In this test, the SSL configuration is incomplete, because
	// TRUST_STORE_PASSWORD is not defined.
	//
	// nri-cassandra ignores incomplete SSL config, so it calls nrjmx
	// without SSL options. Since the Cassandra container is not encrypting
	// JMX connections, the call succeeds.

	stdout, stderr, err := runIntegration(t,
		"CONFIG_PATH=/etc/cassandra/cassandra.yaml",
		"KEY_STORE=/etc/cassandra/keystore.p12",
		"KEY_STORE_PASSWORD=keystorePassword",
		"TRUST_STORE=/etc/cassandra/truststore.p12",
	)

	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.NotNil(t, stderr, "unexpected stderr")

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_Error_InvalidSSLConfig(t *testing.T) {
	// Setting all 4 keystore/truststore options causes nri-cassandra to
	// call nrjmx with SSL options.
	//
	// Because the nri-cassandra command fails, the test confirms that the
	// SSL options are being passed through to nrjmx. Otherwise, the call
	// would succeed, as it does in the NoError_IncompleteSSLConfig test.
	//
	// Call fails because Cassandra is not encrypting JMX connections, and
	// because Keystore and Truststore paths don't exist.

	stdout, stderr, err := runIntegration(t,
		"CONFIG_PATH=/etc/cassandra/cassandra.yaml",
		"KEY_STORE=/etc/cassandra/keystore.p12",
		"KEY_STORE_PASSWORD=keystorePassword",
		"TRUST_STORE=/etc/cassandra/truststore.p12",
		"TRUST_STORE_PASSWORD=trustStorePassword",
	)

	assert.NotNil(t, stdout, "Unexpected stdout")
	assert.NotNil(t, stderr, "Unexpected stderr")
	assert.Error(t, err, "Expected error")
}
