//go:build integration
// +build integration

/*
 * Copyright 2021 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testutils

import (
	"bufio"
	"bytes"
	"context"
	"fmt"

	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/newrelic/nri-cassandra/tests/integration/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

const (
	// Hostname for the Cassandra service. (Will be the cassandra service inside the docker-compose file).
	Hostname           = "cassandra"
	JMXUsername        = "cassandra"
	JMXPassword        = "cassandra"
	KeystorePassword   = "cassandra"
	KeystoreFile       = "/certs/cassandra.keystore"
	TruststorePassword = "cassandra"
	TruststoreFile     = "/certs/cassandra.truststore"
)

// GetIntegrationTestsPath return the absolute path to this project's integration tests.
func GetIntegrationTestsPath() (testsPath string) {
	var err error
	testsPath, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	return
}

// GetPrjDir returns the main directory from this project. (where go.mod file is located.)
func GetPrjDir() string {
	testsPath := GetIntegrationTestsPath()
	// Configure tests to point to the project's tests directory.
	return filepath.Join(testsPath, "../..")
}

// RunDockerExecCommand executes the given command inside the specified container.
func RunDockerExecCommand(ctx context.Context, t *testing.T, containerName string, args []string, envVars map[string]string) (stdout string, stderr string, err error) {
	cmd := NewDockerExecCommand(ctx, t, containerName, args, envVars)

	t.Logf("executing: docker %v", cmd)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	return
}

// NewDockerExecCommand returns a configured un-started exec.Cmd for a docker exec command.
func NewDockerExecCommand(ctx context.Context, t *testing.T, containerName string, args []string, envVars map[string]string) *exec.Cmd {
	cmdLine := []string{
		"exec",
		"-i",
	}

	for key, val := range envVars {
		cmdLine = append(cmdLine, "-e", fmt.Sprintf("%s=%s", key, val))
	}

	cmdLine = append(cmdLine, containerName)
	cmdLine = append(cmdLine, args...)

	t.Logf("executing: docker %s", strings.Join(cmdLine, " "))

	return exec.CommandContext(ctx, "docker", cmdLine...)
}

// Output for a long-running docker exec command.
type Output struct {
	StdoutCh chan string
	StderrCh chan string
}

// NewOutput returns a new Output object.
func NewOutput() *Output {
	size := 1000
	return &Output{
		StdoutCh: make(chan string, size),
		StderrCh: make(chan string, size),
	}
}

// Flush will empty the Output channels and returns the content.
func (o *Output) Flush(t *testing.T) (stdout []string, stderr []string) {
	for {
		select {
		case line := <-o.StdoutCh:
			t.Logf("Flushing stdout line: %s", line)
			stdout = append(stdout, line)
		case line := <-o.StderrCh:
			t.Logf("Flushing stderr line: %s", line)
			stderr = append(stderr, line)
		default:
			return
		}
	}
}

// StartLongRunningProcess will execute a command and will pipe the stdout & stderr into and Output object.
func StartLongRunningProcess(ctx context.Context, t *testing.T, cmd *exec.Cmd) (*Output, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	copyToChan := func(ctx context.Context, reader io.Reader, outputC chan string) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() && ctx.Err() == nil {
			outputC <- scanner.Text()
		}

		if err := scanner.Err(); ctx.Err() == nil && err != nil {
			t.Logf("Error while reading the pipe, %v", err)
			return
		}

		t.Log("Finished reading the pipe")
	}

	output := NewOutput()

	go copyToChan(ctx, stdout, output.StdoutCh)
	go copyToChan(ctx, stderr, output.StderrCh)

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return output, nil
}

// RunDockerCommandForContainer will execute a docker command for the specified containerName.
func RunDockerCommandForContainer(t *testing.T, command, containerName string) error {
	t.Logf("running docker %s container %s", command, containerName)

	cmd := exec.Command("docker", command, containerName)

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("errror while %s the container '%s', error: %v, stderr: %s", command, containerName, err, errBuf.String())
	}

	return nil
}

// RunDockerCompose runs the docker-compose object.
func RunDockerCompose(compose *testcontainers.LocalDockerCompose) error {
	execError := compose.Invoke()
	err := execError.Error
	if err != nil {
		return fmt.Errorf("failed to run docker-compose: error: %w", err)
	}
	return nil
}

// ConfigureCassandraDockerCompose prepares the Cassandra integration test docker-compose.
func ConfigureCassandraDockerCompose() *testcontainers.LocalDockerCompose {
	identifier := strings.ToLower(uuid.New().String())

	composeFilePaths := []string{filepath.Join(GetIntegrationTestsPath(), "docker-compose.yml")}

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)

	compose.WithCommand([]string{"up", "-d", "--build"}).
		WithEnv(map[string]string{
			"EXTRA_JVM_OPTS": "-Dcom.sun.management.jmxremote.authenticate=false " +
				fmt.Sprintf("-Djava.rmi.server.hostname=%s ", Hostname) +
				"-Dcom.sun.management.jmxremote=true ",
		})

	return compose
}

// ConfigureSSLCassandraDockerCompose prepares the Cassandra integration test docker-compose run with SSL JMX.
func ConfigureSSLCassandraDockerCompose() *testcontainers.LocalDockerCompose {
	identifier := strings.ToLower(uuid.New().String())

	composeFilePaths := []string{filepath.Join(GetIntegrationTestsPath(), "docker-compose.yml")}

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)

	compose.WithCommand([]string{"up", "-d", "--build"}).
		WithEnv(map[string]string{
			"EXTRA_JVM_OPTS": "-Dcom.sun.management.jmxremote.authenticate=true " +
				fmt.Sprintf("-Djava.rmi.server.hostname=%s ", Hostname) +
				"-Dcom.sun.management.jmxremote.ssl=true " +
				"-Dcom.sun.management.jmxremote.ssl.need.client.auth=true " +
				"-Dcom.sun.management.jmxremote.registry.ssl=true " +
				"-Dcom.sun.management.jmxremote=true " +
				"-Djavax.net.ssl.keyStore=/opt/cassandra/conf/certs/cassandra.keystore  " +
				fmt.Sprintf("-Djavax.net.ssl.keyStorePassword=%s ", KeystorePassword) +
				"-Djavax.net.ssl.trustStore=/opt/cassandra/conf/certs/cassandra.truststore " +
				fmt.Sprintf("-Djavax.net.ssl.trustStorePassword=%s ", TruststorePassword),
		})

	return compose
}

// AssertReceivedErrors check if at least one the log lines provided contains the given message.
func AssertReceivedErrors(t *testing.T, msg string, errLog ...string) {
	assert.GreaterOrEqual(t, len(errLog), 1)

	for _, line := range errLog {
		if strings.Contains(line, msg) {
			return
		}
	}

	assert.Failf(t, fmt.Sprintf("Expected to find the following error message: %s", msg), "but got %s", errLog)
}

// AssertReceivedPayloadsMatchSchema will check if payloads inside Output object matches the give JSON schema.
func AssertReceivedPayloadsMatchSchema(t *testing.T, ctx context.Context, output *Output, schemaPath string, timeout time.Duration) {
	var cancelFn context.CancelFunc

	ctx, cancelFn = context.WithTimeout(ctx, timeout)
	defer cancelFn()

	validPayloads := 0
	validHeartbeats := 0

	for {
		if validPayloads >= 3 && validHeartbeats >= 3 {
			break
		}

		select {
		case stdoutLine := <-output.StdoutCh:
			if stdoutLine == "{}" {
				t.Log("Received heartbeat")
				validHeartbeats++
			} else {
				t.Logf("Received payload: %s", stdoutLine)

				err := jsonschema.Validate(schemaPath, stdoutLine)
				if err == nil {
					validPayloads++
				}
				assert.NoError(t, err)
			}

		case stderrLine := <-output.StderrCh:
			t.Logf("Received stderr: %s", stderrLine)

			assert.Empty(t, FilterStderr(stderrLine))
		case <-ctx.Done():
			assert.FailNow(t, "didn't received output in time")
		}
	}
}

// FilterStderr is handy to filter some log lines that are expected.
func FilterStderr(content string) string {
	return FilterLines(content, ExpectedErrMessagesFilter)
}

func FilterLines(content string, filter func(line string) bool) string {
	if content == "" {
		return content
	}
	var result []string
	for _, line := range strings.Split(content, "\n") {
		if !filter(line) {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func ExpectedErrMessagesFilter(line string) bool {
	wordsToIgnoreLines := []string{
		"[INFO]",
		"[DEBUG]",
		"non-numeric value for gauge metric",
	}
	for _, chunk := range wordsToIgnoreLines {
		if strings.Contains(line, chunk) {
			return true
		}
	}
	return false
}
