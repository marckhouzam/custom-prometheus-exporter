package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/marckhouzam/custom-prometheus-exporter/configparser"
	"github.com/marckhouzam/custom-prometheus-exporter/exporter"
)

const (
	reloadEndpoint   string = "/-/reload"
	validateEndpoint string = "/validate"
)

var (
	port          int    // The port for the main http server
	configDirPath string // The path to the configuration directory
)

func parseFlags() {
	flag.IntVar(&port, "port", 9555, "The main http port for the global custom-prometheus-exporter")
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

func handleWrongReloadEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`The reload endpoint is ` + reloadEndpoint + ` (and requires a POST)`))
}

func handleReloadEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Write([]byte(`<html>
				        <head><title>Custom Prometheus Exporter</title></head>
					    <body>
					       <h1>Custom Prometheus Exporter</h1>
					       <p>For more information, visit <a href=https://github.org/marckhouzam/custom-prometheus-exporter>GitHub</a></p>
					       <p><h3>Only a POST can be used to reload the Custom Prometheus Exporter</h3></p>
						</body>
					    </html>
				  `))
		return
	}

	// POST method requesting a reload

	// Parse the new configuration, if it is not valid, ignore it and give an error message.
	config := configparser.Config{ConfigDir: configDirPath}

	err := config.ParseConfig()
	if err != nil {
		errorMsg := fmt.Sprint("Reload failed! Error parsing new configuration within directory ", configDirPath, ":\n\t", err)
		log.Println(errorMsg)

		w.Write([]byte(errorMsg))
		return
	}

	// Valid configuration: stop all exporters and restart new ones
	w.Write([]byte(`The reload endpoint is NOT SUPPORTED yet`))

}

func handleRootEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
					<head><title>Custom Prometheus Exporter</title></head>
					<body>
					   <h1>Custom Prometheus Exporter</h1>
					   <p>For more information, visit <a href=https://github.org/marckhouzam/custom-prometheus-exporter>GitHub</a></p>
					   <p><a href='hello'>LIST ALL EXISTING EXPORTERS AND THEIR ENDPOINTS</a></p>
					   </body>
					</html>
				  `))
}

func handleValidateEndpoint(w http.ResponseWriter, r *http.Request) {
	// Parse the new configuration and let the user know if it is valid.
	log.Println(validateEndpoint, "has been called")
	config := configparser.Config{ConfigDir: configDirPath}

	msg := fmt.Sprintln("Configuration is valid.  A reload will succeed.  Use the", reloadEndpoint, "endpoint.")

	err := config.ParseConfig()
	if err != nil {
		msg = fmt.Sprint("Error parsing new configuration within directory ", configDirPath, ":\n\t", err)
	}

	log.Println(msg)
	w.Write([]byte(`<html>
					<head><title>Custom Prometheus Exporter</title></head>
					<body>
                        <h1>Custom Prometheus Exporter</h1>
                        <p>For more information, visit <a href=https://github.org/marckhouzam/custom-prometheus-exporter>GitHub</a></p>
                        <p><h3>` + msg + `</h3></p>
                        </body>
                        </html>
                    `))
}

func setupMainServer() {
	http.HandleFunc("/", handleRootEndpoint)
	http.HandleFunc(reloadEndpoint, handleReloadEndpoint)
	http.HandleFunc("/reload", handleWrongReloadEndpoint)
	http.HandleFunc(validateEndpoint, handleValidateEndpoint)

	log.Println("Main server listening on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func main() {
	parseFlags()

	config := configparser.Config{ConfigDir: configDirPath}

	err := config.ParseConfig()
	if err != nil {
		log.Fatal("Error parsing configuration within directory ", configDirPath, ": ", err)
	}

	go setupMainServer()

	exporter.CreateExporters(config.Exporters)
}
