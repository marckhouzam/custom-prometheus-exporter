package configparser

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const (
	defaultEndpoint      = "/metrics"
	defaultTimeout  uint = 1000
	defaultExecutionType = "bash"
)

// Config is the structure that holds the configuration of the custom-prometheus-exporter
type Config struct {
	// The port used by the main webserver and possibly by some exporters
	MainPort int

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

// MetricsConfig is the structure that contains the information about each metric
type MetricsConfig struct {
	// All fields below must be exported (start with a capital letter)
	// so that the yaml.UnmarshalStrict() method can set them.
	Name       string
	Help       string
	MetricType string `yaml:"type"`
	Executions []struct {
		ExecutionType string `yaml:"type"`
		Command       string
		Timeout       *uint // A pointer so we can check for nil (missing)
		Labels        map[string]string
	}
}

func (c *Config) verifyExporterConfig(exporter *ExporterConfig) error {
	// Make sure 'name' is present
	if exporter.Name == "" {
		return errors.New("Missing field 'name' in top configuration")
	}

	// If 'port' is absent, use the MainPort
	if exporter.Port == 0 {
		exporter.Port = c.MainPort
	}

	// If 'endpoint' is absent, use the the default endpoint
	if exporter.Endpoint == "" {
		exporter.Endpoint = defaultEndpoint
		return nil
	}

	// Add '/' at the start of 'endpoint' if it is missing
	if exporter.Endpoint[0] != '/' {
		exporter.Endpoint = strings.Join([]string{"/", exporter.Endpoint}, "")
	}

	// Make sure 'metrics' is present
	if len(exporter.Metrics) == 0 {
		return errors.New("Missing field 'metrics' in top configuration")
	}

	for i, metric := range exporter.Metrics {
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
			// ExecutionType defaults to the bash shell
			if execution.ExecutionType == "" {
				exporter.Metrics[i].Executions[j].ExecutionType = defaultExecutionType
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

			// If 'timeout' was omitted use the default timeout
			if execution.Timeout == nil {
				defaultT := defaultTimeout
				// Cannot use execution.Timeout as it is a copy of the actual config (since we use range in the for loop)
				exporter.Metrics[i].Executions[j].Timeout = &defaultT
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

// ParseConfig parses the YAML config files which provide
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
		if err = c.verifyExporterConfig(&newExporter); err != nil {
			return err
		}

		// Add the new exporter to the final array of exporters
		c.Exporters = append(c.Exporters, newExporter)
	}

	return nil
}
