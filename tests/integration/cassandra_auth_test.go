//go:build integration

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
)

const (
	SSLConnectionTimeout = "10000"
)

type CassandraSSLTestSuite struct {
	suite.Suite
	cancelComposeCtx context.CancelFunc
}

func TestCassandraSSLTestSuite(t *testing.T) {
	suite.Run(t, new(CassandraSSLTestSuite))
}

func (s *CassandraSSLTestSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelComposeCtx = cancel
	err := testutils.ConfigureSSLCassandraDockerCompose(ctx)
	require.NoError(s.T(), err)

	// Containers are running, but we want to wait that all mBeans are ready.
	s.T().Log("Wait for cassandra to initialize...")
	time.Sleep(60 * time.Second)
}

func (s *CassandraSSLTestSuite) TearDownSuite() {
	s.cancelComposeCtx()
}

func testCassandraIntegration_SSL(t *testing.T, ctx context.Context, cassandraConfig testutils.CassandraConfig) {
	testName := t.Name()
	testCase := struct {
		name string
		env  map[string]string
	}{
		name: fmt.Sprintf("CassandraIntegrationSSL - %s", cassandraConfig.Version),
		env: map[string]string{
			"METRICS": "true",
			"TIMEOUT": SSLConnectionTimeout,

			"HOSTNAME": cassandraConfig.Hostname,
			"USERNAME": testutils.JMXUsername,
			"PASSWORD": testutils.JMXPassword,

			"TRUST_STORE":          testutils.TruststoreFile,
			"TRUST_STORE_PASSWORD": testutils.TruststorePassword,
			"KEY_STORE":            testutils.KeystoreFile,
			"KEY_STORE_PASSWORD":   testutils.KeystorePassword,

			"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
		},
	}
	t.Run(testCase.name, func(t *testing.T) {
		stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, testCase.env)
		assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

		assert.Empty(t, testutils.FilterStderr(stderr))

		schemaDir := fmt.Sprintf("json-schema-files-%s", cassandraConfig.Version)
		schemaPath := filepath.Join(schemaDir, "cassandra-schema-metrics.json")

		err = jsonschema.Validate(schemaPath, stdout)
		assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
	})
}

func (s *CassandraSSLTestSuite) TestCassandraIntegration_SSL() {
	t := s.T()

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	for _, cassandraConfig := range testutils.CassandraConfigs {
		testCassandraIntegration_SSL(t, ctx, cassandraConfig)
	}
}

func testCassandraIntegration_WrongConfig(t *testing.T, cassandraConfig testutils.CassandraConfig) {
	testName := t.Name()
	testCases := []struct {
		name          string
		config        map[string]string
		expectedError string
	}{
		{
			name: fmt.Sprintf("WrongPasword - %s", cassandraConfig.Version),
			config: map[string]string{
				"METRICS": "true",
				"TIMEOUT": SSLConnectionTimeout,

				"HOSTNAME": cassandraConfig.Hostname,
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
			name: fmt.Sprintf("WrongKeyStorePassword - %s", cassandraConfig.Version),
			config: map[string]string{
				"METRICS": "true",
				"TIMEOUT": SSLConnectionTimeout,

				"HOSTNAME": cassandraConfig.Hostname,
				"USERNAME": testutils.JMXUsername,
				"PASSWORD": testutils.JMXPassword,

				"TRUST_STORE":          testutils.TruststoreFile,
				"TRUST_STORE_PASSWORD": "wrong",
				"KEY_STORE":            testutils.KeystoreFile,
				"KEY_STORE_PASSWORD":   "wrong",

				"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),
			},
			expectedError: "java.security.NoSuchAlgorithmException",
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

func (s *CassandraSSLTestSuite) TestCassandraIntegration_WrongConfig() {
	t := s.T()

	for _, cassandraConfig := range testutils.CassandraConfigs {
		testCassandraIntegration_WrongConfig(t, cassandraConfig)
	}
}
