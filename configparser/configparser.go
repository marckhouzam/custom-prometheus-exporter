package configparser

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
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
		err = yaml.UnmarshalStrict(data, &newExporter)
		if err != nil {
			return err
		}
		// Add the new exporter to the final array of exporters
		c.Exporters = append(c.Exporters, newExporter)
	}

	return nil
}
