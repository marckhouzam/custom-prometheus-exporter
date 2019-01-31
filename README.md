# A YAML-defined Custom Prometheus exporter

Create your own Prometheus exporters using simple YAML.

This project allows you to create your own Prometheus exporter which will generate and publish the metrics of your choice in the Prometheus format.  All this through a simple YAML configuration file.  The goal of this project is to allow you to quickly create and easily augment a new Prometheus exporter without having to know how to write a native exporter.

Using a short YAML configuration file, you can define your own metrics and make them available for Prometheus to scrape.

## Using the Custom Prometheus Exporter

The simplest way to try out the Custom Prometheus Exporter is through docker:

```
docker run --rm \
    --name custom-prometheus-exporter -p 12345:12345 \
    -v $(pwd)/example-configurations/test-exporter.yaml:/tmp/test-exporter.yaml \
    marckhouzam/custom-prometheus-exporter -f /tmp/test-exporter.yaml
```
Then you can see the metrics using:
```
curl localhost:12345/test
```

## Configuration

The Custom Prometheus Exporter takes one or more YAML-configuration files, which specify the metrics that are to be collected and how to collect them.  Example configurations can be found in the directory ```example-configurations```.

Here is a sample configuration.  This configuration will create a "docker-exporter", which can be scraped on port ```9550``` and endpoint ```/metrics```.  This exporter generates different metrics, for example the ```docker_container_states_containers``` metric which provides the count of containers in their three possible states (Running, Stopped, Paused). The metrics are collected on each call to the ```/metrics``` endpoint using the three sh-shell commands specified in ```executions```.

```
name: docker-exporter
port: 9550
endpoint: /metrics
metrics:
- name: docker_container_states_containers
  help: The count of containers in various states
  type: gauge
  executions:
  - type: sh
    command: docker info --format '{{ .ContainersRunning }}'
    timeout: 500
    labels:
      state: Running
  - type: sh
    command: docker info --format '{{ .ContainersStopped }}'
    timeout: 500
    labels:
      state: Stopped
  - type: sh
    command: docker info --format '{{ .ContainersPaused }}'
    timeout: 500
    labels:
      state: Paused
```

The generated metric looks like this:

```
# HELP docker_container_states_containers The count of containers in various states
# TYPE docker_container_states_containers gauge
docker_container_states_containers{state="Paused"} 0
docker_container_states_containers{state="Running"} 10
docker_container_states_containers{state="Stopped"} 4
```

### Multiple exporters

The Custom Prometheus Exporter allows you to define many exporters at once. Each exporter **must** be in its own YAML configuration file. All defined exporters will be run concurrently and be accessible using their own configuration-specified endpoint and optionally using their own specific port.  If you want to create metrics that are logically different, it is recommended to use multiple exporters instead of a single exporter lumping all the unrelated metrics together.  Besides cleanly separating the definition of each logical exporter, the separation also allows each exporter to be scraped at different intervals.

You may instead choose to run the Custom Prometheus Exporter multiple times, one for each exporter you want to create.  However, having a single central Custom Prometheus Exporter provides a single set of HTTP endpoints to access information about the different custom exporters that have been instantiated (see [this section](#main-custom-prometheus-exporter-endpoints)).

Here is how to run both example exporters together, using Docker:

```
docker run --rm \
    --name custom-prometheus-exporter -p 12345:12345 -p 9550:9550 \
    -v $(pwd)/example-configurations/test-exporter.yaml:/tmp/test-exporter.yaml \
    -v $(pwd)/example-configurations/docker-exporter.yaml:/tmp/docker-exporter.yaml \
    -v /var/run/docker.sock:/var/run/docker.sock \
    marckhouzam/custom-prometheus-exporter -f /tmp/test-exporter.yaml -f /tmp/docker-exporter.yaml
```
Then you can see the metrics using:
```
curl localhost:9550/metrics
curl localhost:12345/test
```
Note that the example ```docker-exporter.yaml``` uses docker commands.  To be able to run docker commands inside a docker container, you must mount ```/var/run/docker.sock``` as shown above.

### Main Custom Prometheus Exporter endpoints

The actual Custom Prometheus Exporter provides its own endpoints.  By default, the Custom Prometheus Exporter listens on port ```9530``` but it can be changed using the ```-p``` command-line parameter.  This port is not related to the exporters you define, but only to the global Custom Prometheus Exporter endpoints.

You can obtain a list of main endpoints by navigating to ```http://localhost:9530```.

### Configuration API

The format of the YAML configuration is the following:

```
name: string          # A name for the exporter - MANDATORY
port: int             # The TCP port serving the metrics - OPTIONAL, defaults to main port
endpoint: string      # The endpoint serving the metrics - OPTIONAL, defaults to /metrics
metrics:              # An array of metrics to be generated - MANDATORY
- name: string        # The published name of the metric - MANDATORY
  help: string        # The published help message of the metric - MANDATORY
  type: gauge         # Only Prometheus "gauge" is currently supported - MANDATORY
  executions:         # An array of executions to generate the metric - MANDATORY
  - type: sh || bash || tcsh || zsh
                      # The syntax used in the 'command' field must be
                      #   compatible with the shell specified - OPTIONAL, defaults to bash
    command: string   # An sh command that will be run exactly as-specified - MANDATORY
                      #   Shell pipes (|) are allowed.
                      #   The result of the command must be the single
                      #      integer to be used in the metric
    timeout: uint     # Timeout in milliseconds for the command execution - OPTIONAL, defaults to 1000
    labels: map(string, string)
                      # A map of label to value.
                      # The labels qualify further an instance of the metric
                      # This field is MANDATORY if there are more than one execution in
                      #   the executions array of the metric; otherwise it it optional
```

### Backwards-compatibility considerations

Once your YAML-defined exporter is being used, you should be careful when making modifications to its YAML-definition.  It may seem harmless to change the configuration, but changes to some fields could cause consumers to break (such as Prometheus alerts, or Grafana dashboards).

Below are the fields that you should treat as API:
```
name: string
port: int             # API - Changes affect Prometheus configuration
endpoint: string      # API - Changes affect Prometheus configuration
metrics:
- name: string        # API - Changes affect consumers of metrics
  help: string
  type: gauge
  executions:
  - type: sh
    command: string
    timeout: uint
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

### Docker
You can also use Docker.  An example Dockerfile is provided for the example exporters.  However, you may need to modify that Dockerfile for your own exporter needs, to make sure all tools your exporters need will be part of the docker image:
```
docker build -t custom-prometheus-exporter -f <yourDockerfile> .
docker run --rm -d \
    --name custom-prometheus-exporter -p <port>:<yourExporterPort> \
    -v <yourExporterConfigFile.yaml>:/tmp/exporter.yaml \
    custom-prometheus-exporter -f /tmp/exporterConfig.yaml
```
To run the customer-prometheus-exporter example exporters in docker see the example further above.

### Running automated tests

To run the automated tests you will also need to install the [go assert package](https://godoc.org/gotest.tools/assert):

```
go get gotest.tools/assert
```

Then run the tests:
```
go test ./...
```

### TODO list

- [ ] Add more automated Tests
- [ ] Support other types of metrics (e.g., Counter)
- [ ] Support for native execution instead of shell command (e.g., running a script)
- [ ] Add a Kubernetes Helm chart
