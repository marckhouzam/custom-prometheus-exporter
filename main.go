package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/marckhouzam/custom-prometheus-exporter/configparser"
)

var (
	configDirPath string // The path to the configuration directory
)

func parseFlags() {
	flag.StringVar(&configDirPath, "configDirPath", "example-configurations",
		"The path to the directory where the YAML files defining the exporters reside.")

	flag.Parse()

	if configDirPath == "" {
		fmt.Println("You cannot specify an empty configuration directory")
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	parseFlags()

	config := configparser.Config{ConfigDir: configDirPath}

	err := config.ParseConfig()
	if err != nil {
		log.Fatal("Error parsing configuration within ", configDirPath, ": ", err)
	}
}
