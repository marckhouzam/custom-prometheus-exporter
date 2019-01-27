package exporter

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/marckhouzam/custom-prometheus-exporter/configparser"
	"github.com/marckhouzam/custom-prometheus-exporter/webservers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metricsCollector struct {
	mutex         sync.RWMutex
	metricsConfig []configparser.MetricsConfig
	gaugeVecs     []*prometheus.GaugeVec
}

func getKeys(mymap map[string]string) []string {
	i := 0
	keys := make([]string, len(mymap))
	for k := range mymap {
		keys[i] = k
		i++
	}
	return keys
}

// CreateExporters instantiates each exporter as requested
// in the configuration
func CreateExporters(config configparser.Config) {

	for _, exporterCfg := range config.Exporters {
		metricsCollector := metricsCollector{}
		metricsCollector.addMetrics(exporterCfg.Metrics)

		// Don't use the default registry to avoid getting the go collector
		// and all its metrics
		registry := prometheus.NewRegistry()
		registry.MustRegister(&metricsCollector)
		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

		webservers.CreateExporterWebserver(&handler, &exporterCfg)
	}
}

func (m *metricsCollector) addMetrics(metrics []configparser.MetricsConfig) {
	m.metricsConfig = metrics
	m.gaugeVecs = make([]*prometheus.GaugeVec, len(metrics))

	for i, metric := range m.metricsConfig {
		if metric.MetricType == "gauge" {
			m.gaugeVecs[i] = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: metric.Name,
					Help: metric.Help,
				},
				getKeys(metric.Executions[0].Labels),
			)
		} else {
			panic("Only the gauge metric is supported at the moment")
		}
	}
}

func (m *metricsCollector) getMetrics() {
	for i, metric := range m.metricsConfig {
		for _, execution := range metric.Executions {
			cmd := exec.Command(execution.ExecutionType, "-c", execution.Command)

			var timedout bool
			timeout := *execution.Timeout
			if timeout != 0 {
				timer := time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
					timedout = true
					cmd.Process.Kill()
				})
				defer timer.Stop()
			}
			output, err := cmd.Output()
			if timedout {
				log.Println("Timeout when running:", execution.Command)
				continue
			}

			if err != nil {
				log.Println("Got error when running:", execution.Command+":", err)
				continue
			}

			countStr := strings.TrimSpace(string(output))
			count, err := strconv.ParseFloat(countStr, 64)
			if err != nil {
				log.Printf(
					"Got error when parsing result of: "+execution.Command+
						". Expecting integer result but got %v and error "+err.Error(), countStr)
				continue
			}

			// Now set the metrics
			m.gaugeVecs[i].With(execution.Labels).Set(count)
		}
	}
}

// Describe - Implements Collector.Describe
func (m *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range m.gaugeVecs {
		m.Describe(ch)
	}
}

// Collect - Implements Collector.Collect
func (m *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	m.mutex.Lock() // To protect metrics from concurrent collects.
	defer m.mutex.Unlock()

	m.getMetrics()

	for _, m := range m.gaugeVecs {
		m.Collect(ch)
	}
}
