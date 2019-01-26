# Custom Prometheus Exporter Helm Chart

This chart install the [Custom Prometheus Exporter](https://github.com/marckhouzam/custom-prometheus-exporter).

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ helm install --name my-release chart/custom-prometheus-exporter
```

## Configuration

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
$ helm install --name my-release -f values.yaml chart/custom-prometheus-exporter
```

> **Tip**: You can use the default [values.yaml](values.yaml)

