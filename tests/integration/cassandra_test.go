///go:build integration
//go:build integration
// +build integration

/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integration

import (
	"context"
	"fmt"
	"github.com/newrelic/nri-cassandra/tests/integration/jsonschema"
	"github.com/newrelic/nri-cassandra/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	envCassandraVersion = "3.11.0"
	timeout             = "5000"
)

var (
	iName = "cassandra"

	iVersion = "1.1.0"

	integrationBinPath       = "/nri-cassandra"
	integrationContainerName = "integration_nri-cassandra_1"
	cassandraContainerName   = "integration_cassandra_1"

	schemaDir = fmt.Sprintf("json-schema-files-%s", envCassandraVersion)
)

type CassandraTestSuite struct {
	suite.Suite
	compose *testcontainers.LocalDockerCompose
}

func TestCassandraTestSuite(t *testing.T) {
	suite.Run(t, new(CassandraTestSuite))
}

func (s *CassandraTestSuite) SetupSuite() {
	s.compose = testutils.ConfigureCassandraDockerCompose()

	err := testutils.RunDockerCompose(s.compose)
	require.NoError(s.T(), err)

	// Could not rely on testcontainers wait strategies here, as the server might be up but not reporting all mbeans.
	s.T().Log("Wait for cassandra to initialize...")
	time.Sleep(30 * time.Second)
}

func (s *CassandraTestSuite) TearDownSuite() {
	s.compose.Down()
}

func (s *CassandraTestSuite) TestCassandraIntegration_ValidArguments() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"CONFIG_PATH":     "/etc/cassandra/cassandra.yaml",
		"HOSTNAME":        testutils.Hostname,
		"TIMEOUT":         timeout,
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.Empty(t, testutils.FilterStderr(stderr), "Unexpected stderr")

	schemaPath := filepath.Join(schemaDir, "cassandra-schema.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func (s *CassandraTestSuite) TestCassandraIntegration_OnlyMetrics() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"METRICS":         "true",
		"CONFIG_PATH":     "/etc/cassandra/cassandra.yaml",
		"HOSTNAME":        testutils.Hostname,
		"TIMEOUT":         timeout,
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.Empty(t, testutils.FilterStderr(stderr), "Unexpected stderr")

	schemaPath := filepath.Join(schemaDir, "cassandra-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func (s *CassandraTestSuite) TestCassandraIntegration_OnlyInventory() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"INVENTORY":       "true",
		"CONFIG_PATH":     "/etc/cassandra/cassandra.yaml",
		"HOSTNAME":        testutils.Hostname,
		"TIMEOUT":         timeout,
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.Empty(t, testutils.FilterStderr(stderr), "Unexpected stderr")

	schemaPath := filepath.Join(schemaDir, "cassandra-schema-inventory.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func (s *CassandraTestSuite) TestCassandraIntegration_Error_InvalidHostname() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"METRICS":         "true",
		"HOSTNAME":        "wrong_hostname",
		"TIMEOUT":         timeout,
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	expectedErrorMessage := "Unknown host"

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.Error(t, err, "Expected error")

	assert.Empty(t, stdout)
	testutils.AssertReceivedErrors(t, expectedErrorMessage, strings.Split(stderr, "\n")...)
}

func (s *CassandraTestSuite) TestCassandraIntegration_Error_InvalidPort() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"METRICS":         "true",
		"HOSTNAME":        testutils.Hostname,
		"PORT":            "1",
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	expectedErrorMessage := "Connection refused"

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.Error(t, err, "Expected error")

	assert.Empty(t, stdout, "Unexpected stdout content")
	testutils.AssertReceivedErrors(t, expectedErrorMessage, strings.Split(stderr, "\n")...)
}

func (s *CassandraTestSuite) TestCassandraIntegration_Error_InvalidConfigPath_NonExistingFile() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"CONFIG_PATH":     "/nonExisting.yaml",
		"INVENTORY":       "true",
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	expectedErrorMessage := "no such file or directory"

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.Error(t, err, "Expected error")

	assert.Empty(t, stdout, "Unexpected stdout content")
	testutils.AssertReceivedErrors(t, expectedErrorMessage, strings.Split(stderr, "\n")...)
}

func (s *CassandraTestSuite) TestCassandraIntegration_NoError_InvalidConfigPath_NonExistingFile_OnlyMetrics() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"CONFIG_PATH":     "/nonExisting.yaml",
		"METRICS":         "true",
		"HOSTNAME":        testutils.Hostname,
		"TIMEOUT":         timeout,
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.Empty(t, testutils.FilterStderr(stderr), "Unexpected stderr")

	// Cassandra 3.7 does not report DataCenter and Racks names through JMX
	schemaPath := filepath.Join(schemaDir, "cassandra-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func (s *CassandraTestSuite) TestCassandraIntegration_Error_InvalidConfigPath_ExistingFile() {
	t := s.T()

	testName := testutils.GetTestName(t)

	path := "/tmp/empty.yaml"

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	_, _, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{"touch", path}, nil)
	assert.NoError(t, err)

	env := map[string]string{
		"CONFIG_PATH":     path,
		"INVENTORY":       "true",
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.Error(t, err)

	assert.Empty(t, stdout, "Unexpected stdout output")

	expectedErrorMessage := "config path not correctly set, cannot fetch inventory data"
	testutils.AssertReceivedErrors(t, expectedErrorMessage, strings.Split(stderr, "\n")...)
}

func (s *CassandraTestSuite) TestCassandraIntegration_Error_InvalidConfigPath_ExistingDirectory() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"CONFIG_PATH":     "/etc/cassandra/",
		"INVENTORY":       "true",
		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.Error(t, err)

	assert.Empty(t, stdout, "Unexpected stdout output")

	expectedErrorMessage := "is a directory"
	testutils.AssertReceivedErrors(t, expectedErrorMessage, strings.Split(stderr, "\n")...)
}
