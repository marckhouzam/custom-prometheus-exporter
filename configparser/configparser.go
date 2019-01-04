package configparser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type executionType int

const (
	sh executionType = iota
	numExecutionTypes
)

type executionConfig struct {
	execType executionType
	command  string
	labels   map[string]string
}

type metricType int

const (
	gauge metricType = iota
	counter
	numMetricType
)

type metricsConfig struct {
	name       string
	help       string
	metricType metricType
	execution  []executionConfig
}
type exporterConfig struct {
	port     int
	endpoint string
	metrics  []metricsConfig
}

// Config is the structure that holds the configuration of the custom-prometheus-exporter
// as defined by the user.
type Config struct {
	ConfigDir string
	exporters map[string]exporterConfig
}

func (c *Config) parseMetricsConfig(infoMap map[interface{}]interface{}) error {
	return nil
}

func (c *Config) parseExporterConfig(name string, infoMap map[interface{}]interface{}) error {
	log.Println("Got exporter", name, "with info", infoMap)

	c.exporters = make(map[string]exporterConfig)
	c.exporters[name] = exporterConfig{}

	var port int
	var endpoint string
	var metrics []metricsConfig
	var ok bool

	for k, v := range infoMap {

		key := k.(string)

		switch key {
		case "port":
			port, ok = v.(int)
			if !ok {
				return fmt.Errorf("The value for the 'port' key must be an int, got a %v instead", v)
			}
		case "endpoint":
			endpoint, ok = v.(string)
			if !ok {
				return fmt.Errorf("The value for the 'endpoint' key must be a string, got a %v instead", v)
			}
		case "metrics":
			// value, ok := v.([]map[interface{}]interface{})
			// if !ok {
			// 	return errors.New(`The value of the "metrics" key must be an array of maps`)
			// }

			// for _, m := range value {

			// }
			// metrics, err = c.parseMetricsConfig(value)
			// if !ok {
			// 	return fmt.Errorf("The value for the 'endpoint' key must be a string, got a %v instead", v)
			// }
		default:
			return errors.New(`Found invalid key (` + key + `). Key must be one of: port, endpoint, metrics`)
		}
	}

	c.exporters[name] = exporterConfig{port: port, endpoint: endpoint, metrics: metrics}

	return nil
}

func (c *Config) parseConfigData(genericMap map[interface{}]interface{}) error {
	for k, v := range genericMap {
		key, ok := k.(string)

		if !ok || (key != "exporters" && key != "exporter") {
			return errors.New(`Found invalid key (` + key + `). Key must be "exporters"`)
		}

		value, ok := v.(map[interface{}]interface{})
		if !ok {
			return errors.New(`The value of the "exporters" key must be a map`)
		}

		for k, v = range value {
			key, ok = k.(string)
			if !ok {
				return errors.New(
					"Found invalid key (" + key + "). Key must be a string name for the exporter")
			}

			value, ok := v.(map[interface{}]interface{})
			if !ok {
				return fmt.Errorf(
					"The value of the exporter name key must be a map. Got %v instead", v)
			}

			err := c.parseExporterConfig(key, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ParseConfig parses the YAML files present in configDir which provide
// the definition and configuration of the exporters
func (c *Config) ParseConfig() error {
	// Check if the directory exists
	if _, err := os.Stat(c.ConfigDir); os.IsNotExist(err) {
		return err
	}

	// List all files in the directory
	f, err := os.Open(c.ConfigDir)
	if err != nil {
		return err
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return err
	}

	// Now parse the content of each file to populate our configuration
	for _, file := range files {
		// First extract the data out of the file
		data, err := ioutil.ReadFile(c.ConfigDir + "/" + file.Name())
		if err != nil {
			return err
		}

		// Now parse the yaml into a generic map
		configMap := make(map[interface{}]interface{})
		err = yaml.Unmarshal(data, &configMap)
		if err != nil {
			return err
		}

		// Finally parse the generic map into our data structure
		err = c.parseConfigData(configMap)
		if err != nil {
			return errors.New("Error in " + file.Name() + ": " + err.Error())
		}
	}

	fmt.Println("The final config is", c.exporters)
	return nil
}
