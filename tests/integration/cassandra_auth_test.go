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
	// suite.Run(t, new(CassandraSSLTestSuite))
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

func (s *CassandraSSLTestSuite) TestCassandraIntegration_SSL() {
	t := s.T()

	testName := t.Name()

	ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	testCases := []struct {
		env              map[string]string
		cassandraVersion string
	}{}
	for _, cassandraConfig := range testutils.CassandraConfigs {
		testCases = append(testCases, struct {
			env              map[string]string
			cassandraVersion string
		}{
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
			cassandraVersion: cassandraConfig.Version,
		})
	}

	for _, testCase := range testCases {
		stdout, stderr, err := testutils.RunDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, testCase.env)
		assert.NoError(t, err, "It isn't possible to execute Cassandra integration binary.")

		assert.Empty(t, testutils.FilterStderr(stderr))

		schemaDir := fmt.Sprintf("json-schema-files-%s", testCase.cassandraVersion)
		schemaPath := filepath.Join(schemaDir, "cassandra-schema-metrics.json")

		err = jsonschema.Validate(schemaPath, stdout)
		assert.NoError(t, err, "The output of Cassandra integration doesn't have expected format.")
	}
}

func (s *CassandraSSLTestSuite) TestCassandraIntegration_WrongConfig() {
	t := s.T()

	testName := t.Name()

	testCases := []struct {
		name          string
		config        map[string]string
		expectedError string
	}{}

	for _, cassandraConfig := range testutils.CassandraConfigs {
		testCases = append(testCases, struct {
			name          string
			config        map[string]string
			expectedError string
		}{
			name: "WrongPassword",
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
		})
		testCases = append(testCases, struct {
			name          string
			config        map[string]string
			expectedError string
		}{
			name: "WrongKeyStorePassword",
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
		})
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
