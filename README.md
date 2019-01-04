# A YAML-defined Custom Prometheus exporter

This is a generic Prometheus exporter which can be configured using YAML to create metrics.  The goal of this project is to allow to easily create a new Prometheus exporter without having to know how to write a native prometheus exporter.

## Configuration

## Running the custom-prometheus-exporter

Natively:
```
go build
./custom-prometheus-exporter [--endpoint <endpoint>] [--port <port>]
```

By default, the metrics are exposed on port 9555 under the /metrics endpoint.  The port can be changed using the ```--port``` flag, while the endpoint can be changed using the ```--endpoint``` parameter.

Using docker:
```
docker build -t custom-prometheus-exporter .
docker run --rm -d --name custom-prometheus-exporter [-p <port>:9555] custom-prometheus-exporter
```
Using Kubernetes:
```
helm upgrade -i custom-prometheus-exporter ./chart
```

Note that you will need to specify the correct image name and version using the --set flag.
