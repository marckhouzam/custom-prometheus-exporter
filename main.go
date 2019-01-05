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
	defaultMainPort  int    = 9530 // Reserved at https://github.com/prometheus/prometheus/wiki/Default-port-allocations
)

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

var (
	port        int       // The port for the main http server
	configFiles arrayFlag // The list of input configuration files
)

func parseFlags() {
	flag.IntVar(&port, "p", defaultMainPort, "The main http port for the global custom-prometheus-exporter")
	flag.Var(&configFiles, "f", "A configuration file defining some exporters.\n"+
		"This flag can be used multiple times to include multiple files.")

	flag.Parse()

	if len(configFiles) == 0 {
		fmt.Println("You must specify at least one configuration file.")
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
	config := configparser.Config{ConfigFiles: configFiles}

	err := config.ParseConfig()
	if err != nil {
		errorMsg := fmt.Sprint("Reload failed! Error parsing new configuration:\n\t", err)
		log.Println(errorMsg)

		w.Write([]byte(errorMsg))
		return
	}

	// TODO Valid configuration: stop all exporters and restart new ones
	w.Write([]byte(`The reload endpoint is NOT SUPPORTED yet`))

}

func handleRootEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
	    <html>
	    <head><title>Custom Prometheus Exporter</title></head>
	    <body>
	        <h1>Custom Prometheus Exporter</h1>
	        <p>For more information, visit <a href=https://github.org/marckhouzam/custom-prometheus-exporter>GitHub</a></p>
	        <p>Available main endpoints:</p>
	        <ul>
	            <li><a href=/validate>/validate</a> - Validate if modifications to the configuration files are valid.
	            <li><a href=/-/reload>/-/reload</a> - POST only. Reload the configuration files (assuming they have changed) and restart all exporters
	        </ul>
	    </body>
	    </html>
	`))
}

func handleValidateEndpoint(w http.ResponseWriter, r *http.Request) {
	// Parse the new configuration and let the user know if it is valid.
	log.Println(validateEndpoint, "has been called")
	config := configparser.Config{ConfigFiles: configFiles}

	msg := fmt.Sprintln("Configuration is valid.  A reload will succeed.  Use the", reloadEndpoint, "endpoint.")

	err := config.ParseConfig()
	if err != nil {
		msg = fmt.Sprint("Error parsing new configuration:\n\t", err)
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

func main() {
	parseFlags()

	config := configparser.Config{ConfigFiles: configFiles}

	err := config.ParseConfig()
	if err != nil {
		log.Fatal("Error parsing configuration: ", err)
	}

	// Setup main server
	go func() {
		http.HandleFunc("/", handleRootEndpoint)
		http.HandleFunc(reloadEndpoint, handleReloadEndpoint)
		http.HandleFunc("/reload", handleWrongReloadEndpoint)
		http.HandleFunc(validateEndpoint, handleValidateEndpoint)

		log.Println("Main server listening on port", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()

	exporter.CreateExporters(config.Exporters)
}
