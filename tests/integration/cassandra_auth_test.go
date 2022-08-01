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
	"github.com/newrelic/infra-integrations-sdk/log"
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
	log.Info("Wait for cassandra to initialize...")
	time.Sleep(30 * time.Second)
}

func (s *CassandraSSLTestSuite) TearDownSuite() {
	s.compose.Down()
}

func (s *CassandraSSLTestSuite) TestCassandraIntegration_SSL() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"METRICS": "true",
		"TIMEOUT": SSLConnectionTimeout,

		"HOSTNAME": testutils.Hostname,
		"USERNAME": testutils.JMXUsername,
		"PASSWORD": testutils.JMXPassword,

		"TRUST_STORE":          "/certs/cassandra.truststore",
		"TRUST_STORE_PASSWORD": testutils.TruststorePassword,
		"KEY_STORE":            "/certs/cassandra.keystore",
		"KEY_STORE_PASSWORD":   testutils.KeystorePassword,

		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, integrationContainerName, []string{integrationBinPath}, env)
	assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

	assert.Empty(t, testutils.FilterStderr(stderr))

	schemaPath := filepath.Join(schemaDir, "cassandra-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
}

func (s *CassandraSSLTestSuite) TestCassandraIntegration_WrongPassword() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"METRICS": "true",
		"TIMEOUT": SSLConnectionTimeout,

		"HOSTNAME": testutils.Hostname,
		"USERNAME": "wrong",
		"PASSWORD": "wrong",

		"TRUST_STORE":          "/certs/cassandra.truststore",
		"TRUST_STORE_PASSWORD": testutils.TruststorePassword,
		"KEY_STORE":            "/certs/cassandra.keystore",
		"KEY_STORE_PASSWORD":   testutils.KeystorePassword,

		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, integrationContainerName, []string{integrationBinPath}, env)
	assert.Error(t, err)
	assert.Empty(t, stdout)

	testutils.AssertReceivedErrors(t, "Authentication failed! Invalid username or password", strings.Split(stderr, "\n")...)
}

func (s *CassandraSSLTestSuite) TestCassandraIntegration_WrongKeyStorePassword() {
	t := s.T()

	testName := testutils.GetTestName(t)

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	env := map[string]string{
		"METRICS": "true",
		"TIMEOUT": SSLConnectionTimeout,

		"HOSTNAME": testutils.Hostname,
		"USERNAME": testutils.JMXUsername,
		"PASSWORD": testutils.JMXPassword,

		"TRUST_STORE":          "/certs/cassandra.truststore",
		"TRUST_STORE_PASSWORD": "wrong",
		"KEY_STORE":            "/certs/cassandra.keystore",
		"KEY_STORE_PASSWORD":   "wrong",

		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
	}

	stdout, stderr, err := testutils.RunDockerExecCommand(ctx, integrationContainerName, []string{integrationBinPath}, env)
	assert.Error(t, err)
	assert.Empty(t, stdout)

	testutils.AssertReceivedErrors(t, "java.security.NoSuchAlgorithmException", strings.Split(stderr, "\n")...)
}
