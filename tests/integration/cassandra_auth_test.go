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
	SSLConnectionTimeout = "10000"
)

type CassandraSSLTestSuite struct {
	suite.Suite
	compose *testcontainers.LocalDockerCompose
}

func TestCassandraSSLTestSuite(t *testing.T) {
	suite.Run(t, new(CassandraSSLTestSuite))
}

func (s *CassandraSSLTestSuite) SetupSuite() {
	s.compose = testutils.ConfigureSSLCassandraDockerCompose()

	err := testutils.RunDockerCompose(s.compose)
	require.NoError(s.T(), err)

	// Could not rely on testcontainers wait strategies here, as the server might be up but not reporting all mbeans.
	s.T().Log("Wait for cassandra to initialize...")
	time.Sleep(30 * time.Second)
}

func (s *CassandraSSLTestSuite) TearDownSuite() {
	s.compose.Down()
}

func (s *CassandraSSLTestSuite) TestCassandraIntegration_SSL() {
	t := s.T()

	testName := t.Name()

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"METRICS": "true",
		"TIMEOUT": SSLConnectionTimeout,

		"HOSTNAME": testutils.Hostname,
		"USERNAME": testutils.JMXUsername,
		"PASSWORD": testutils.JMXPassword,

		"TRUST_STORE":          testutils.TruststoreFile,
		"TRUST_STORE_PASSWORD": testutils.TruststorePassword,
		"KEY_STORE":            testutils.KeystoreFile,
		"KEY_STORE_PASSWORD":   testutils.KeystorePassword,

		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)
	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.Empty(t, testutils.FilterStderr(stderr))

	schemaPath := filepath.Join(schemaDir, "cassandra-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func (s *CassandraSSLTestSuite) TestCassandraIntegration_WrongConfig() {
	t := s.T()

	testName := t.Name()

	testCases := []struct {
		name          string
		config        map[string]string
		expectedError string
	}{
		{
			name: "WrongPassword",
			config: map[string]string{
				"METRICS": "true",
				"TIMEOUT": SSLConnectionTimeout,

				"HOSTNAME": testutils.Hostname,
				"USERNAME": "wrong",
				"PASSWORD": "wrong",

				"TRUST_STORE":          testutils.TruststoreFile,
				"TRUST_STORE_PASSWORD": testutils.TruststorePassword,
				"KEY_STORE":            testutils.KeystoreFile,
				"KEY_STORE_PASSWORD":   testutils.KeystorePassword,

				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			expectedError: "Authentication failed! Invalid username or password",
		},
		{
			name: "WrongKeyStorePassword",
			config: map[string]string{
				"METRICS": "true",
				"TIMEOUT": SSLConnectionTimeout,

				"HOSTNAME": testutils.Hostname,
				"USERNAME": testutils.JMXUsername,
				"PASSWORD": testutils.JMXPassword,

				"TRUST_STORE":          testutils.TruststoreFile,
				"TRUST_STORE_PASSWORD": "wrong",
				"KEY_STORE":            testutils.KeystoreFile,
				"KEY_STORE_PASSWORD":   "wrong",

				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			expectedError: "Authentication failed! Invalid username or password",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancelFn()

			stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, testCase.config)
			assert.Error(t, err)
			assert.Empty(t, stdout)

			testutils.AssertReceivedErrors(t, testCase.expectedError, strings.Split(stderr, "\n")...)
		})
	}
}
