package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
)

func runCrashingTest(t *testing.T, testCode func()) (outStr, errStr string) {
	crashEnvVarName := "RUN_CRASHING_CODE"
	crashEnvVarValue := "1"

	if os.Getenv(crashEnvVarName) == crashEnvVarValue {
		testCode()

		// Cancel the second-level test so that we don't run
		// the code after the call to this method more than once
		t.SkipNow()

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), crashEnvVarName+"="+crashEnvVarValue)

	var timedout bool
	timer := time.AfterFunc(500*time.Millisecond, func() {
		timedout = true
		cmd.Process.Kill()
	})
	defer timer.Stop()

	bytes, err := cmd.Output()

	if timedout {
		t.Fatal("Timeout")
		return
	}

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return string(bytes), string(e.Stderr)
	}

	t.Fatal("process ran without an error, while we expected an error code")

	return
}

func TestFlagsEmpty(t *testing.T) {
	out, _ := runCrashingTest(t, func() {
		os.Args = []string{"."}
		main()
	})
	assert.Assert(t, strings.Contains(out, "must specify at least one configuration file"), "Output: "+out)
}

func TestFlagsPortOnly(t *testing.T) {
	out, _ := runCrashingTest(t, func() {
		os.Args = []string{".", "-p", "12345"}
		main()
	})
	assert.Assert(t, strings.Contains(out, "must specify at least one configuration file"), "Output: "+out)
}

func TestFlagsInvalidPort(t *testing.T) {
	invalidPort := "abcde"
	_, err := runCrashingTest(t, func() {
		os.Args = []string{".", "-p", invalidPort}
		main()
	})
	assert.Assert(t, strings.Contains(err, "invalid value \""+invalidPort+"\" for flag -p"), "Error: "+err)
}

func TestFlagsValid(t *testing.T) {
	validPortStr := "12345"
	configFile := "example-configurations/test-exporter.yaml"

	validPort, _ := strconv.Atoi(validPortStr)

	os.Args = []string{".", "-p", validPortStr, "-f", configFile}
	port, configFiles := parseFlags()

	assert.Equal(t, port, validPort)
	assert.Equal(t, len(configFiles), 1)
	assert.Equal(t, configFiles[0], configFile)
}

func TestFlagsDefaultPort(t *testing.T) {
	configFile := "example-configurations/test-exporter.yaml"

	os.Args = []string{".", "-f", configFile}
	port, configFiles := parseFlags()

	assert.Equal(t, port, defaultMainPort)
	assert.Equal(t, len(configFiles), 1)
	assert.Equal(t, configFiles[0], configFile)
}
