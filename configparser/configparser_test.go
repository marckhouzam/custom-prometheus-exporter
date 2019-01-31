package configparser

import (
	"io/ioutil"
	"os"
	"testing"

	"gotest.tools/assert"
)

func createFile(t *testing.T, data string) string {
	filename := "/tmp/customPromExporterTest.data"
	assert.NilError(t, ioutil.WriteFile(filename, []byte(data), 0644))
	return filename
}

func removeFile(name string) {
	os.Remove(name)
}

func TestSingleConfigFile(t *testing.T) {
	c := Config{ConfigFiles: []string{"../example-configurations/test-exporter.yaml"}}
	assert.NilError(t, c.ParseConfig())
}

func TestTwoConfigFiles(t *testing.T) {
	c := Config{ConfigFiles: []string{"../example-configurations/test-exporter.yaml", "../example-configurations/docker-exporter.yaml"}}
	assert.NilError(t, c.ParseConfig())
}

func TestMissingSingleConfigFile(t *testing.T) {
	c := Config{ConfigFiles: []string{"missing.yaml"}}
	assert.ErrorContains(t, c.ParseConfig(), "missing.yaml: no such file or directory")
}

func TestMissingFirstConfigFile(t *testing.T) {
	c := Config{ConfigFiles: []string{"missing.yaml", "../example-configurations/docker-exporter.yaml"}}
	assert.ErrorContains(t, c.ParseConfig(), "missing.yaml: no such file or directory")
}

func TestMissingSecondConfigFile(t *testing.T) {
	c := Config{ConfigFiles: []string{"../example-configurations/test-exporter.yaml", "missing.yaml"}}
	assert.ErrorContains(t, c.ParseConfig(), "missing.yaml: no such file or directory")

}
func TestInvalidConfigFile(t *testing.T) {
	errorTag := "unexpectedTag"

	data := `
name: test-exporter
port: 12345
endpoint: /test
` + errorTag + `: true    # This is not valid and should cause a parse failure
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "field "+errorTag+" not found")
}

func TestMissingName(t *testing.T) {
	data := `
#name: test-exporter       # Missing field should cause error
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`

	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'name' in top configuration")
}

func TestMissingPort(t *testing.T) {
	data := `
name: test-exporter
#port: 12345          # Missing port should default to port of main program
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`

	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.NilError(t, c.ParseConfig())
	// The port should remain unset as an indicator that the main port should be used
	assert.Equal(t, c.Exporters[0].Port, 0)
	// Check that some other field of the config are correct
	assert.Equal(t, *c.Exporters[0].Metrics[0].Executions[0].Timeout, defaultTimeout)
}

func TestMissingEndpoint(t *testing.T) {
	data := `
name: test-exporter
port: 12345
#endpoint: /test         # Missing field should default to /metrics
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.NilError(t, c.ParseConfig())
	assert.Equal(t, c.Exporters[0].Endpoint, "/metrics")
}

func TestIncompleteEndpoint(t *testing.T) {
	endpoint := "test" // Endpoint not starting with /

	data := `
name: test-exporter
port: 12345
endpoint: ` + endpoint + `         # Missing / should be added automatically
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.NilError(t, c.ParseConfig())
	// Make sure / was added
	assert.Equal(t, c.Exporters[0].Endpoint, "/"+endpoint)
	// Check that some other field of the config are correct
	assert.Equal(t, *c.Exporters[0].Metrics[0].Executions[0].Timeout, defaultTimeout)
}

func TestMissingMetrics(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
#metrics:            # Missing field should cause error
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'metrics' in top configuration")
}

func TestEmptyMetrics(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:            # Empty field should cause error
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'metrics' in top configuration")
}

func TestMissingMetricName(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- help: Some values
  # name: test_gauge_values       # Missing field should cause error
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'name' in 'metrics' configuration")
}

func TestMissingMetricHelp(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
# help: Some values            # Missing field should cause error
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'help' in 'metrics' configuration")
}

func TestWrongMetricType(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: wrong                  # Wrong field should cause error
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Wrong value for field 'type' in 'metrics' configuration")
}

func TestMissingMetricType(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
#  type: gauge                  # Missing field should cause error
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'type' in 'metrics' configuration")
}

func TestMissingMetricExecutions(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
# executions:              # Missing field should cause error
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'executions' in 'metrics' configuration")
}

func TestEmptyMetricExecutions(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:              # Empty field should cause error
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'executions' in 'metrics' configuration")
}

func TestMissingMetricExecutionType(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - command: expr 111
#    type: sh                  # Missing field should trigger a default
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.NilError(t, c.ParseConfig())
	assert.Equal(t, *c.Exporters[0].Metrics[0].Executions[0].ExecutionType, defaultExecutionType)
}

func TestWrongMetricExecutionType(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - command: expr 111
    type: wrong                  # Wrong value for field should cause error
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Wrong value for field 'type' in 'executions' configuration")
}

func TestMissingMetricExecutionCommand(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
#   command: expr 111           # Missing field should cause error
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'command' in 'executions' configuration")
}

func TestEmptyMetricExecutionCommand(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: ""           # Empty field should cause error
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'command' in 'executions' configuration")
}

func TestMissingMetricExecutionTimeout(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: sleep 5
#    timeout: 3                # Missing field should default to defaultTimeout
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.NilError(t, c.ParseConfig())
	assert.Equal(t, *c.Exporters[0].Metrics[0].Executions[0].Timeout, defaultTimeout)
}

func TestNegativeMetricExecutionTimeout(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: sleep 5
    timeout: -3                # Negative timeout should cause an error
    labels:
      order: first
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "unmarshal error")
}

func TestMissingMetricExecutionLabelsForMoreThanOneExec(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
  - type: sh
    command: expr 222
#    labels:              # Missing field cause an error
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'labels' in 'executions' configuration")
}

func TestEmptyMetricExecutionLabelsMoreThanOneExec(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:
      order: first
  - type: sh
    command: expr 222
    labels:                   # Empty field should cause an error
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.ErrorContains(t, c.ParseConfig(), "Missing field 'labels' in 'executions' configuration")
}

func TestMissingMetricExecutionLabelsForOneExec(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
#    labels:              # Missing field should be accepted
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.NilError(t, c.ParseConfig())
	assert.Equal(t, len(c.Exporters[0].Metrics[0].Executions[0].Labels), 0, "Labels should be empty based on config")
}

func TestEmptyMetricExecutionLabelsForOneExec(t *testing.T) {
	data := `
name: test-exporter
port: 12345
endpoint: /test
metrics:
- name: test_gauge_values
  help: Some values
  type: gauge
  executions:
  - type: sh
    command: expr 111
    labels:                   # Empty field should be accepted
`
	filename := createFile(t, data)
	defer removeFile(filename)

	c := Config{ConfigFiles: []string{filename}}
	assert.NilError(t, c.ParseConfig())
	assert.Equal(t, len(c.Exporters[0].Metrics[0].Executions[0].Labels), 0, "Labels should be empty based on config")
}
