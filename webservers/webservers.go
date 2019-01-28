package webservers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/marckhouzam/custom-prometheus-exporter/configparser"
	"github.com/marckhouzam/custom-prometheus-exporter/metricscollector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	reloadEndpoint   string = "/-/reload"
	validateEndpoint string = "/validate"
)

var (
	configuration configparser.Config
	mainServer    *http.Server
	webServers    []*http.Server
)

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
	newConfig := configparser.Config{
		MainPort:    configuration.MainPort,
		ConfigFiles: configuration.ConfigFiles,
	}

	err := newConfig.ParseConfig()
	if err != nil {
		errorMsg := fmt.Sprint("Reload failed! Error parsing new configuration:\n\t", err)
		log.Println(errorMsg)

		w.Write([]byte(errorMsg))
		return
	}

	// New configuration is valid, stop the web servers and restart them
	configuration = newConfig

	if err = shutdownWebServers(); err != nil {
		log.Fatal("Terminal error while stopping webservers for exporters: ", err)
	}
}

func handleMainRootEndpoint(w http.ResponseWriter, r *http.Request) {
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
	newConfig := configparser.Config{ConfigFiles: configuration.ConfigFiles}

	var msg string
	err := newConfig.ParseConfig()
	if err != nil {
		msg = fmt.Sprint("Error parsing new configuration:\n\t", err)
	} else {
		msg = fmt.Sprintln("Configuration is valid.  A reload will succeed.  Use the", reloadEndpoint, "endpoint.")
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

func handleExporterRootEndpoint(name string, endpoint string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
							<head><title>` + name + `</title></head>
							<body>
							   <h1>` + name + `</h1>
							   <p>This exporter was created in YAML using the <a href=https://github.com/marckhouzam/custom-prometheus-exporter>Custom Prometheus Exporter</a></p>
							   <p><a href='` + endpoint + `'>Metrics</a></p>
							   </body>
							</html>
						  `))
	}
}

// CreateExporters instantiates each exporter as requested
// in the configuration
func createExporters() {
	webServers = make([]*http.Server, 0, len(configuration.Exporters))

	for _, exporterCfg := range configuration.Exporters {
		metricsCollector := metricscollector.MetricsCollector{}
		metricsCollector.AddMetrics(exporterCfg.Metrics)

		// Don't use the default registry to avoid getting the go collector
		// and all its metrics
		registry := prometheus.NewRegistry()
		registry.MustRegister(&metricsCollector)
		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

		createExporterWebserver(&handler, exporterCfg)
	}
}

func createMainServer() {
	// Setup main server
	server := http.NewServeMux()

	server.HandleFunc("/", handleMainRootEndpoint)
	server.HandleFunc("/reload", handleWrongReloadEndpoint)
	server.HandleFunc(reloadEndpoint, handleReloadEndpoint)
	server.HandleFunc(validateEndpoint, handleValidateEndpoint)

	mainServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", configuration.MainPort),
		Handler: server,
	}
}

// CreateExporterWebserver -
func createExporterWebserver(handler *http.Handler, exporterCfg configparser.ExporterConfig) {
	if exporterCfg.Port == configuration.MainPort {
		// The exporter should re-use the main port
		server := mainServer.Handler.(*http.ServeMux)
		server.Handle(fmt.Sprintf("%s", exporterCfg.Endpoint), *handler)
		log.Println(exporterCfg.Name, "listening on port", configuration.MainPort, "and endpoint", exporterCfg.Endpoint)
	} else {
		// Need to call serveMux to be able to run multiple servers at the same time
		server := http.NewServeMux()
		// Handle the endpoint serving the metrics
		server.Handle(fmt.Sprintf("%s", exporterCfg.Endpoint), *handler)
		// Give some info on the root endpoint
		server.HandleFunc("/", handleExporterRootEndpoint(exporterCfg.Name, exporterCfg.Endpoint))

		// Create a server object which we can later Shutdown()
		newWebServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", exporterCfg.Port),
			Handler: server,
		}
		webServers = append(webServers, newWebServer)
		log.Println(exporterCfg.Name, "listening on port", exporterCfg.Port, "and endpoint", exporterCfg.Endpoint)
	}
}

// CreateListenAndServe creates then starts all webservers
// and blocks on the main one.
func CreateListenAndServe(config configparser.Config) {
	configuration = config

	// Infinite loop which allows us to shutdown all servers and re-create them
	// to support the reload functionality
	for {
		createMainServer()
		createExporters()

		for _, w := range webServers {
			// Use go routine so as to not block, since we run multiple servers.
			webServer := w
			go func() {
				webServer.ListenAndServe()
			}()
		}

		log.Println("Main server listening on port", configuration.MainPort)
		// Block on the main server
		mainServer.ListenAndServe()

		log.Println("Main server has shutdown")
	}
}

// ShutdownWebServers gracefully shutsdown every exporter, allowing a maximum
// 5 seconds for the shutdown to complete
func shutdownWebServers() error {
	// Do the shutdowns in parallel for efficiency
	var wg sync.WaitGroup
	var shutdownErr error

	for _, s := range webServers {
		server := s
		if server != nil {
			wg.Add(1)
			go func() {
				// Decrement the counter when the goroutine completes.
				defer wg.Done()

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				log.Println("About to shutdown exporter server")

				if err := server.Shutdown(ctx); err != nil {
					shutdownErr = err
				}
				log.Println("Exporter server has shutdown")
			}()
		}
	}
	// Wait for all shutdowns to complete
	wg.Wait()

	if shutdownErr != nil {
		log.Println("Shutdown error for exporter server", shutdownErr)
		return shutdownErr
	}

	// Now shutdown the main server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("About to shutdown main server")

	if err := mainServer.Shutdown(ctx); err != nil {
		log.Println("Shutdown error for main server", err)
		// Ignore Error.  We keep getting one but if we ignore things still work...
		log.Println("Ignoring error and continuing")
		// return err
	}

	return nil
}
