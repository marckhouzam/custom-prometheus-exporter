# A YAML-defined Custom Prometheus exporter

A YAML-defined Custom Prometheus exporter which allows you to generate and publish metrics in the Prometheus format.  The goal of this project is to allow to easily create a new Prometheus exporter without having to know how to write a native exporter.

Using a short YAML configuration file, you can define your own metrics and make them available for Prometheus to scrape.

## Configuration

The Custom Prometheus Exporter take one or more YAML-configuration files, which specify the metrics that are to be collected and how to collect them.  Example configurations can be found in the directory ```example-configurations```.

Here is a sample configuration.  This configuration will create a "docker-exporter", which can be scraped on port ```9550``` and endpoint ```/metrics```.  This exporter generates a single metric (named: ```docker_container_states_containers```) that provides the count of containers in their three possible states (Running, Stopped, Paused). This metric is collected on each call to the /metrics endpoint using the three sh-shell commands specified in ```executions```.

```
- name: docker-exporter
  port: 9550
  endpoint: /metrics
  metrics:
  - name: docker_container_states_containers
    help: The count of containers in various states
    type: gauge
    executions:
    - type: sh
      command: docker info --format '{{ .ContainersRunning }}'
      labels:
        state: Running
    - type: sh
      command: docker info --format '{{ .ContainersStopped }}'
      labels:
        state: Stopped
    - type: sh
      command: docker info --format '{{ .ContainersPaused }}'
      labels:
        state: Paused
```

The generated metric looks like this:

```
# HELP docker_container_states_containers The count of containers in various states
# TYPE docker_container_states_containers gauge
docker_container_states_containers{state="Paused"} 0
docker_container_states_containers{state="Running"} 0
docker_container_states_containers{state="Stopped"} 4
```
### Configuration API

The format of the YAML configuration is the following:

```
- name: string          # A name for the exporter
  port: int             # The TCP port serving the metrics
  endpoint: string      # The endpoint serving the metrics
  metrics:              # An array of metrics to be generated
  - name: string        # The published name of the metric
    help: string        # The published help message of the metric
    type: gauge         # Only Prometheus "gauge" is currently supported
    executions:         # An array of executions to generate the metric
    - type: sh          # Only sh is currently supported
      command: string   # An sh command that will be run exactly as-specified
                        #   Shell pipes (|) are allowed.
                        #   The result of the command must be the single
                        #      integer to be used in the metric
      labels: map(string, string)
                        # A map of label to value which qualifies an instance
                        #   of the metric
```

### Backwards-compatibility considerations

Once your YAML-defined exporter is being used, you should be careful when making modifications to its YAML-definition.  It may seem harmless to change the configuration, but changes to some fields could cause consumers to break (such as Prometheus alerts, or Grafana dashboards).

Below are the fields that you should treat as API:
```
- name: string
  port: int             # API - Changes affect Prometheus configuration
  endpoint: string      # API - Changes affect Prometheus configuration
  metrics:
  - name: string        # API - Changes affect consumers of metrics
    help: string
    type: gauge
    executions:
    - type: sh
      command: string
      labels: map(string, string)
                        # API - Changes affect consumers of metrics
```

## Contributing

Contributions are welcomed!  You can submit code or documentation Pull Requests, or open issues with ideas or problems you found.

### Building the project

The code is written in [Go](https://tour.golang.org).  If you don't already have your Go development environment setup, start [here](https://golang.org/doc/install
).

To compile the custom-prometheus-exporter you will also need to install the [go yaml package](https://github.com/go-yaml/yaml):

```
go get gopkg.in/yaml.v2
```

Then get and compile the code:
```
git clone https://github.com/marckhouzam/custom-prometheus-exporter.git
cd custom-prometheus-exporter
go build
```
### Running

Natively, after you've compiled it:
```
./custom-prometheus-exporter -f yamlConfigFile1 [-f yamlConfigFile2] ...
```

By default, the metrics are exposed on port 9555 under the /metrics endpoint.  The port can be changed using the ```-p``` flag.

### Docker
You an also use Docker:
```
docker build -t custom-prometheus-exporter .
docker run --rm -d \
    --name custom-prometheus-exporter -p <port>:<yourExporterPort> \
    -v <yourExporterConfigFile.yaml>:/tmp/exporter.yaml \
    custom-prometheus-exporter -f /tmp/exporterConfig.yaml
```

### Automated Tests

Still on my TODO list, but I know Go provides a framework to make that easy.
