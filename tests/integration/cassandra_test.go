package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-cassandra/tests/integration/helpers"
	"github.com/newrelic/nri-cassandra/tests/integration/jsonschema"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"os/exec"
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

	defaultBinPath  = "/nr-cassandra"
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

	require.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	if stderr != "" {
		t.Fatalf("Unexpected stderr output: %s", stderr)
	}

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema.json")

	err = jsonschema.Validate(schemaPath, stdout)

	require.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_OnlyMetrics(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"METRICS=true",
		fmt.Sprintf("TIMEOUT=%v", timeout),
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	require.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	if stderr != "" {
		t.Fatalf("Unexpected stderr output: %s", stderr)
	}

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	require.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_OnlyInventory(t *testing.T) {
	testName := helpers.GetTestName(t)
	stdout, stderr, err := runIntegration(t,
		"CONFIG_PATH=/etc/cassandra/cassandra.yaml",
		"INVENTORY=true",
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	require.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	if stderr != "" {
		t.Fatalf("Unexpected stderr output: %s", stderr)
	}

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-inventory.json")

	err = jsonschema.Validate(schemaPath, stdout)
	require.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_Error_NoUserCredentials(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)
	cmd := exec.Command(*(binPath))

	cmd.Env = []string{
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}
	expectedErrorMessage := "Access denied"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err == nil {
		t.Fatal("Error not returned")
	}
	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestCassandraIntegration_NoError_NoUserCredentials_OnlyInventory(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t,
		"INVENTORY=true",
		"CONFIG_PATH=/etc/cassandra/cassandra.yaml",
		fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName),
	)

	require.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	if stderr != "" {
		t.Fatalf("Unexpected stderr output: %s", stderr)
	}

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-inventory.json")

	err = jsonschema.Validate(schemaPath, stdout)
	require.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
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
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)
	}
	if stdout != "" {
		t.Fatalf("Unexpected output: %s", stdout)
	}
}

func TestCassandraIntegration_Error_InvalidPassword(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)
	cmd := exec.Command(*binPath)

	cmd.Env = []string{
		"USERNAME=monitorRole",
		"PASSWORD=invalid_password",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}
	expectedErrorMessage := "Access denied"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err == nil {
		t.Fatal("Error not returned")
	}
	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestCassandraIntegration_Error_InvalidUsername(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)
	cmd := exec.Command(*binPath)

	cmd.Env = []string{
		"USERNAME=invalid_username",
		"PASSWORD=monitorPwd",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}
	expectedErrorMessage := "Access denied"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err == nil {
		t.Fatal("Error not returned")
	}
	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestCassandraIntegration_Error_InvalidHostname(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)
	cmd := exec.Command(*binPath)

	cmd.Env = []string{
		"HOSTNAME=nonExistingHost",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}
	expectedErrorMessage := "no such host"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err == nil {
		t.Fatal("Error not returned")
	}
	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestCassandraIntegration_Error_InvalidPort(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	testName := helpers.GetTestName(t)
	cmd := exec.Command(*binPath)

	cmd.Env = []string{
		"USERNAME=monitorRole",
		"PASSWORD=monitorPwd",
		"PORT=1",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}
	expectedErrorMessage := "Failed to connect, invalid port"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err == nil {
		t.Fatal("Error not returned")
	}
	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
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
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)
	}

	if stdout != "" {
		t.Fatalf("Unexpected output: %s", stdout)
	}
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

	require.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	if stderr != "" {
		t.Fatalf("Unexpected stderr output: %s", stderr)
	}

	// Cassandra 3.7 does not report DataCenter and Racks names through JMX
	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	require.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")

	err = jsonschema.Validate(schemaPath, stdout)
	require.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func TestCassandraIntegration_Error_InvalidConfigPath_ExistingFile(t *testing.T) {
	t.Skip("Skipping test - not correct message return - it will be fixed in JIRA ticket IHOST-176")
	tmpfile, err := ioutil.TempFile("", "empty.yaml")
	if err != nil {
		t.Fatalf("Cannot create a new temporary file, got error: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	testName := helpers.GetTestName(t)
	cmd := exec.Command(*binPath)
	cmd.Env = []string{
		fmt.Sprintf("CONFIG_PATH=%s", tmpfile.Name()),
		"INVENTORY=true",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}

	expectedErrorMessage := "Config path not correctly set, cannot fetch inventory data"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err = cmd.Run()
	if err != nil {
		t.Fatalf("It isn't possible to execute Cassandra integration binary. Err: %s -- %s", err, errbuf.String())
	}

	schemaPath := filepath.Join(schemaFolder, "cassandra-schema-empty.json")

	err = jsonschema.Validate(schemaPath, outbuf.String())
	if err != nil {
		t.Fatalf("The output of Cassandra integration doesn't have expected format. Err: %s", err)
	}

	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if !errMatch {
		t.Fatalf("Expected warning message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
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
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)
	}

	if stdout != "" {
		t.Fatalf("Unexpected output: %s", stdout)
	}
}
