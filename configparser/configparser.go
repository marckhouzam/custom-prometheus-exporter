package configparser

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const (
	defaultEndpoint = "/metrics"
)

// Config is the structure that holds the configuration of the custom-prometheus-exporter
// as defined by the user.
type Config struct {
	// The path of each configuration file defining the exporters
	ConfigFiles []string
	// The result of parsing the configuration files, which provides
	// all necessary details to create the exporters
	Exporters []ExporterConfig
}

// ExporterConfig is the structure that contains the information about each defined
// exporter that will be instantiated
type ExporterConfig struct {
	// All fields below must be exported (start with a capital letter)
	// so that the yaml.UnmarshalStrict() method can set them.
	Name     string
	Port     int
	Endpoint string
	Metrics  []MetricsConfig
}

// MetricsConfig -
type MetricsConfig struct {
	// All fields below must be exported (start with a capital letter)
	// so that the yaml.UnmarshalStrict() method can set them.
	Name       string
	Help       string
	MetricType string `yaml:"type"`
	Executions []struct {
		ExecutionType string `yaml:"type"`
		Command       string
		Labels        map[string]string
	}
}

func verifyExporterConfig(config *ExporterConfig) error {
	// Make sure 'name' is present
	if config.Name == "" {
		return errors.New("Missing field 'name' in top configuration")
	}

	// Make sure 'port' is present
	if config.Port == 0 {
		return errors.New("Missing field 'port' in top configuration")
	}

	// If 'endpoint' is absent, use the the default endpoint
	if config.Endpoint == "" {
		config.Endpoint = defaultEndpoint
		return nil
	}

	// Add '/' at the start of 'endpoint' if it is missing
	if config.Endpoint[0] != '/' {
		config.Endpoint = strings.Join([]string{"/", config.Endpoint}, "")
		return nil
	}

	// Make sure 'metrics' is present
	if len(config.Metrics) == 0 {
		return errors.New("Missing field 'metrics' in top configuration")
	}

	for i, metric := range config.Metrics {
		if metric.Name == "" {
			return errors.New("Missing field 'name' in 'metrics' configuration of metric " + string(i))
		}

		if metric.Help == "" {
			return errors.New("Missing field 'help' in 'metrics' configuration of metric " + string(i))
		}

		if metric.MetricType == "" {
			return errors.New("Missing field 'type' in 'metrics' configuration of metric " + string(i))
		}

		if metric.MetricType != "gauge" {
			return errors.New("Wrong value for field 'type' in 'metrics' configuration of metric " + string(i) +
				". Supported values are: gauge")
		}

		// Make sure 'executions' is present
		if len(metric.Executions) == 0 {
			return errors.New("Missing field 'executions' in 'metrics' configuration of metric " + string(i))
		}

		for j, execution := range metric.Executions {
			if execution.ExecutionType == "" {
				return errors.New("Missing field 'type' in 'executions' configuration of metric " + string(i) +
					" and execution " + string(j))
			}

			if execution.ExecutionType != "sh" && execution.ExecutionType != "bash" &&
				execution.ExecutionType != "tcsh" && execution.ExecutionType != "zsh" {
				return errors.New("Wrong value for field 'type' in 'executions' configuration of metric " + string(i) +
					" and execution " + string(j) + ". Supported values are: sh, bash, tcsh or zsh")
			}

			if execution.Command == "" {
				return errors.New("Missing field 'command' in 'executions' configuration of metric " + string(i) +
					" and execution " + string(j))
			}

			// Check 'labels'. Can be omitted only if there is a single element
			// in the 'executions' array, for this metric
			if len(metric.Executions) > 1 && len(execution.Labels) == 0 {
				return errors.New("Missing field 'labels' in 'executions' configuration of metric " + string(i) +
					" and execution " + string(j))
			}
		}
	}
	return nil
}

// ParseConfig parses the YAML files present in configDir which provide
// the definition and configuration of the exporters
func (c *Config) ParseConfig() error {
	// Check if all files exist
	for _, file := range c.ConfigFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return err
		}
	}

	// Now parse the content of each file to populate our configuration
	for _, file := range c.ConfigFiles {
		// First extract the data out of the file
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		// Now parse the yaml directly into our data structure
		newExporter := ExporterConfig{}
		if err = yaml.UnmarshalStrict(data, &newExporter); err != nil {
			return err
		}

		// Do some sanity checks on the configuration
		if err = verifyExporterConfig(&newExporter); err != nil {
			return err
		}

		// Add the new exporter to the final array of exporters
		c.Exporters = append(c.Exporters, newExporter)
	}

	return nil
}
