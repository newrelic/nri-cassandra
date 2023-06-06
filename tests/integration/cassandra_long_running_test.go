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
	"path/filepath"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-cassandra/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CassandraLongRunningTestSuite struct {
	suite.Suite
	cancelComposeCtx context.CancelFunc
}

func TestCassandraLongRunningTestSuite(t *testing.T) {
	suite.Run(t, new(CassandraLongRunningTestSuite))
}

func (s *CassandraLongRunningTestSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelComposeCtx = cancel
	err := testutils.ConfigureCassandraDockerCompose(ctx)
	require.NoError(s.T(), err)

	// Containers are running, but we want to wait that all mBeans are ready.
	log.Info("Wait for cassandra to initialize...")
	time.Sleep(60 * time.Second)
}

func (s *CassandraLongRunningTestSuite) TearDownSuite() {
	s.cancelComposeCtx()
}

func (s *CassandraLongRunningTestSuite) TestCassandraIntegration_LongRunningIntegration() {
	t := s.T()

	testName := t.Name()

	ctx, cancelFn := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancelFn()

	env := map[string]string{
		"METRICS":  "true",
		"HOSTNAME": testutils.Hostname,
		"TIMEOUT":  timeout,

		"LONG_RUNNING":       "true",
		"INTERVAL":           "2",
		"HEARTBEAT_INTERVAL": "2",

		"NRIA_CACHE_PATH": fmt.Sprintf("/tmp/%v.json", testName),

		// Uncomment those for troubleshooting.
		// "VERBOSE":               "true",
		// "ENABLE_INTERNAL_STATS": "true",
	}

	cmd := testutils.NewDockerExecCommand(ctx, t, integrationContainerName, []string{integrationBinPath}, env)

	output, err := testutils.StartLongRunningProcess(ctx, t, cmd)
	assert.NoError(t, err)

	go func() {
		err = cmd.Wait()

		// Avoid failing the test when we cancel the context at the end. (This is a long-running integration)
		if ctx.Err() == nil {
			assert.NoError(t, err)
		}
	}()

	schemaDir := fmt.Sprintf("json-schema-files-%s", envCassandraVersion)
	schemaFile := filepath.Join(schemaDir, "cassandra-schema-metrics.json")
	testutils.AssertReceivedPayloadsMatchSchema(t, ctx, output, schemaFile, 10*time.Second)

	err = testutils.RunDockerCommandForContainer(t, "stop", cassandraContainerName)
	require.NoError(t, err)

	// Wait for the jmx connection to fail. We need to give it time as it might
	// take time to timeout. The assumption is that after 60 seconds even if the jmx connection hangs,
	// when we restart the container again it will fail because of a new server listening on jmx port.
	log.Info("Waiting for jmx connection to fail")
	time.Sleep(60 * time.Second)

	err = testutils.RunDockerCommandForContainer(t, "start", cassandraContainerName)
	require.NoError(t, err)

	log.Info("Waiting for cassandra server to be up again")
	time.Sleep(30 * time.Second)

	_, stderr := output.Flush(t)

	testutils.AssertReceivedErrors(t, "connection error", stderr...)

	testutils.AssertReceivedPayloadsMatchSchema(t, ctx, output, schemaFile, 10*time.Second)
}
