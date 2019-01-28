package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/marckhouzam/custom-prometheus-exporter/configparser"
	"github.com/marckhouzam/custom-prometheus-exporter/webservers"
)

const defaultMainPort int = 9530 // Reserved at https://github.com/prometheus/prometheus/wiki/Default-port-allocations

// Support for flags that fill an array, this allows to pass the same
// flag multiple times at the command line, for example to specify
// multiple configuration files
type arrayFlag []string

func (a *arrayFlag) String() string {
	return fmt.Sprintf("%v", []string(*a))
}

func (a *arrayFlag) Set(str string) error {
	*a = append(*a, str)
	return nil
}

// End arrayFlag

func parseFlags() (int, []string) {
	// Use a new flag set to allow tests to call this method more than once
	var f = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var port int
	var configFiles = arrayFlag{}

	f.IntVar(&port, "p", defaultMainPort, "The main http port for the global custom-prometheus-exporter")
	f.Var(&configFiles, "f", "A configuration file defining some exporters.\n"+
		"This flag can be used multiple times to include multiple files.")

	f.Parse(os.Args[1:])

	if len(configFiles) == 0 {
		fmt.Println("You must specify at least one configuration file.")
		fmt.Println()
		f.Usage()
		os.Exit(1)
	}

	return port, configFiles
}

func main() {
	port, files := parseFlags()

	config := configparser.Config{
		MainPort:    port,
		ConfigFiles: files,
	}

	if err := config.ParseConfig(); err != nil {
		log.Fatal("Error parsing configuration: ", err)
	}

	// Blocks forever
	webservers.CreateListenAndServe(config)
}
