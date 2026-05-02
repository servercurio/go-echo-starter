package main

import (
	"os"
	"testing"

	"github.com/servercurio/go-echo-starter/internal/env"
	asrt "github.com/stretchr/testify/assert"
)

const envPrefix = "CRUSNIK"

func TestNoPanicsInMain(t *testing.T) {
	assert := asrt.New(t)
	// Suppress the log output to avoid cluttering the test output
	assert.NoError(os.Setenv(env.AddPrefix(envPrefix, "daemon_log_enabled"), "false"), "failed to set CRUSNIK_DAEMON_LOG_ENABLED")
	assert.NoError(os.Setenv(env.AddPrefix(envPrefix, "http_access_log_enabled"), "false"), "failed to set CRUSNIK_HTTP_ACCESS_LOG_ENABLED")
	assert.NotPanics(main, "main() should not panic")
}
