/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newrelic/nri-cassandra/tests/integration/jsonschema"
	"github.com/newrelic/nri-cassandra/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

const (
	envCassandraVersion = "3.11.0"
	timeout             = "5000"

	integrationBinPath       = "/nri-cassandra"
	integrationContainerName = "integration_nri-cassandra_1"
	cassandraContainerName   = "integration_cassandra_1"
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

func (s *CassandraTestSuite) TestCassandraIntegration_Success() {
	t := s.T()

	testName := t.Name()

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	testCases := []struct {
		name       string
		config     map[string]string
		schemaFile string
	}{
		{
			name: "ValidArguments",
			config: map[string]string{
				"CONFIG_PATH":     "/etc/cassandra/cassandra.yaml",
				"HOSTNAME":        testutils.Hostname,
				"TIMEOUT":         timeout,
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			schemaFile: "cassandra-schema.json",
		},
		{
			name: "OnlyMetrics",
			config: map[string]string{
				"METRICS":         "true",
				"CONFIG_PATH":     "/etc/cassandra/cassandra.yaml",
				"HOSTNAME":        testutils.Hostname,
				"TIMEOUT":         timeout,
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			schemaFile: "cassandra-schema-metrics.json",
		},
		{
			name: "OnlyInventory",
			config: map[string]string{
				"INVENTORY":       "true",
				"CONFIG_PATH":     "/etc/cassandra/cassandra.yaml",
				"HOSTNAME":        testutils.Hostname,
				"TIMEOUT":         timeout,
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			schemaFile: "cassandra-schema-inventory.json",
		},
		{
			name: "InvalidConfigPath_NonExistingFile_OnlyMetrics",
			config: map[string]string{
				"CONFIG_PATH":     "/nonExisting.yaml",
				"METRICS":         "true",
				"HOSTNAME":        testutils.Hostname,
				"TIMEOUT":         timeout,
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			schemaFile: "cassandra-schema-metrics.json",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, testCase.config)
			assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

			assert.Empty(t, testutils.FilterStderr(stderr), "Unexpected stderr")

			schemaDir := fmt.Sprintf("json-schema-files-%s", envCassandraVersion)
			schemaPath := filepath.Join(schemaDir, testCase.schemaFile)

			err = jsonschema.Validate(schemaPath, stdout)
			assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
		})
	}
}

func (s *CassandraTestSuite) TestCassandraIntegration_WrongConfig() {
	t := s.T()

	testName := t.Name()

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	// Create an empty file in the container, required for some tests.
	path := "/tmp/empty.yaml"
	_, _, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{"touch", path}, nil)
	assert.NoError(t, err)

	testCases := []struct {
		name          string
		config        map[string]string
		expectedError string
	}{
		{
			name: "InvalidPort",
			config: map[string]string{
				"METRICS":         "true",
				"HOSTNAME":        "wrong_hostname",
				"TIMEOUT":         timeout,
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			expectedError: "Unknown host",
		},
		{
			name: "InvalidHostname",
			config: map[string]string{
				"METRICS":         "true",
				"HOSTNAME":        testutils.Hostname,
				"PORT":            "1",
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			expectedError: "Connection refused",
		},
		{
			name: "InvalidConfigPath_NonExistingFile",
			config: map[string]string{
				"CONFIG_PATH":     "/nonExisting.yaml",
				"INVENTORY":       "true",
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			expectedError: "no such file or directory",
		},
		{
			name: "InvalidConfigPath_ExistingDirectory",
			config: map[string]string{
				"CONFIG_PATH":     "/etc/cassandra/",
				"INVENTORY":       "true",
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			expectedError: "is a directory",
		},
		{
			name: "InvalidConfigPath_ExistingFile",
			config: map[string]string{
				"CONFIG_PATH":     path,
				"INVENTORY":       "true",
				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			expectedError: "config path not correctly set, cannot fetch inventory data",
		},
	}

	for _, testCase := range testCases {
		stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, testCase.config)
		assert.Error(t, err, "Expected error")

		assert.Empty(t, stdout, "Unexpected stdout content")
		testutils.AssertReceivedErrors(t, testCase.expectedError, strings.Split(stderr, "\n")...)
	}
}
