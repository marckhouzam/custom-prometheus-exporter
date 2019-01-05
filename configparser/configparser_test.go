package configparser

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func createFile(data string) (string, error) {
	filename := "/tmp/customPromExporterTest.data"
	err := ioutil.WriteFile(filename, []byte(data), 0644)
	return filename, err
}

func TestSingleConfigFile(t *testing.T) {
	c := Config{ConfigFiles: []string{"../example-configurations/test-exporter.yaml"}}
	if err := c.ParseConfig(); err != nil {
		t.Error(err)
	}
}

func TestTwoConfigFiles(t *testing.T) {
	c := Config{ConfigFiles: []string{"../example-configurations/test-exporter.yaml", "../example-configurations/docker-exporter.yaml"}}
	if err := c.ParseConfig(); err != nil {
		t.Error(err)
	}
}

func TestMissingSingleConfigFile(t *testing.T) {
	c := Config{ConfigFiles: []string{"missing.yaml"}}
	if err := c.ParseConfig(); err == nil {
		t.Error("Did not detect missing config file")
	}
}

func TestMissingFirstConfigFile(t *testing.T) {
	c := Config{ConfigFiles: []string{"missing.yaml", "../example-configurations/docker-exporter.yaml"}}
	if err := c.ParseConfig(); err == nil {
		t.Error("Did not detect missing config file")
	}
}

func TestMissingSecondConfigFile(t *testing.T) {
	c := Config{ConfigFiles: []string{"../example-configurations/test-exporter.yaml", "missing.yaml"}}
	if err := c.ParseConfig(); err == nil {
		t.Error("Did not detect missing config file")
	}
}
func TestInvalidConfigFile(t *testing.T) {
	errorTag := "unexpectedTag"

	filename, err := createFile(`
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
`)
	if err != nil {
		t.Error(err)
		return
	}

	defer os.Remove(filename)

	c := Config{ConfigFiles: []string{filename}}
	if err := c.ParseConfig(); err == nil || !strings.Contains(err.Error(), errorTag+" not found") {
		if err == nil {
			t.Error("Did not detect bad config file")
			return
		} else {
			t.Error("Did not detect bad config file, but got error:\n\t" + err.Error())
			return
		}
	}
}
