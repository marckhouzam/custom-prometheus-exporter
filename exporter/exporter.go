package exporter

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "docker"

// Types of metrics
const (
	containerStates = iota
	imageTypes
	totalMetricTypes
)

func init() {
	// Make sure the collector is able to run docker commands.
	// This is to quickly fail in case the docker-exporter was
	// started without access to /var/run/docker.sock
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	if err != nil {
		fmt.Println("The docker-exporter must be able to run docker commands.",
			"\nWhen running the docker-exporter in a docker container, you need to mount",
			"\n  /var/run/docker.sock")
		os.Exit(1)
	}
}

// MetricsCollector - An object to collect the metrics
type MetricsCollector struct {
	mutex     sync.RWMutex
	gaugeVecs []*prometheus.GaugeVec
}

// AddMetrics defines the metrics that will be provided
func (m *MetricsCollector) AddMetrics() {
	m.gaugeVecs = make([]*prometheus.GaugeVec, totalMetricTypes)

	m.gaugeVecs[containerStates] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "docker_container_states_containers",
			Help: "The count of containers in various states",
		},
		[]string{
			// What state is the container in (running, stopped, paused)
			"state",
		},
	)

	m.gaugeVecs[imageTypes] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "docker_image_types_images",
			Help: "The count of images of various types",
		},
		[]string{
			// What type of image (top-level, intermediate, dangling)
			"type",
		},
	)
}

func (m *MetricsCollector) getImageMetrics() {

	var err error
	var output []byte
	var countStr string

	// Get the total number of images
	output, err = exec.Command("sh", "-c", "docker images --all --quiet | wc -l").Output()
	if err != nil {
		log.Println("Got error when running docker command", err)
		return
	}

	countStr = strings.TrimSpace(string(output))
	totalImages, err := strconv.ParseFloat(countStr, 64)
	if err != nil {
		log.Println("Got error parsing float", err)
		return
	}

	// Get the number of top-level images
	output, err = exec.Command("sh", "-c", "docker images --quiet | wc -l").Output()
	if err != nil {
		log.Println("Got error when running docker command", err)
		return
	}

	countStr = strings.TrimSpace(string(output))
	topLevelImages, err := strconv.ParseFloat(countStr, 64)
	if err != nil {
		log.Println("Got error parsing float", err)
		return
	}

	// Get the number of dangling images
	output, err = exec.Command("sh", "-c", "docker images --quiet --filter dangling=true | wc -l").Output()
	if err != nil {
		log.Println("Got error when running docker command", err)
		return
	}

	countStr = strings.TrimSpace(string(output))
	danglingImages, err := strconv.ParseFloat(countStr, 64)
	if err != nil {
		log.Println("Got error parsing float", err)
		return
	}

	// Now set the metrics
	m.gaugeVecs[imageTypes].With(prometheus.Labels{"type": "top-level"}).Set(topLevelImages)
	m.gaugeVecs[imageTypes].With(prometheus.Labels{"type": "intermediate"}).Set(totalImages - topLevelImages)
	m.gaugeVecs[imageTypes].With(prometheus.Labels{"type": "dangling"}).Set(danglingImages)
}

func (m *MetricsCollector) getStateMetrics() {
	states := []string{"Running", "Stopped", "Paused"}

	for _, state := range states {
		formatStr := strings.Join([]string{"{{ .Containers", state, " }}"}, "")
		cmd := exec.Command("docker", "info", "--format", formatStr)
		output, err := cmd.Output()
		if err != nil {
			log.Println("Got error when running docker command", err)
		} else {
			out := strings.TrimSpace(string(output))
			count, err := strconv.ParseFloat(out, 64)
			if err != nil {
				log.Println("Got error parsing float", err)
			} else {
				m.gaugeVecs[containerStates].With(prometheus.Labels{"state": state}).Set(count)
			}
		}
	}
}

// Describe - Implements Collector.Describe
func (m *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range m.gaugeVecs {
		m.Describe(ch)
	}
}

// Collect - Implements Collector.Collect
func (m *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	m.mutex.Lock() // To protect metrics from concurrent collects.
	defer m.mutex.Unlock()

	m.getStateMetrics()
	m.getImageMetrics()

	for _, m := range m.gaugeVecs {
		m.Collect(ch)
	}
}
